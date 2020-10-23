package main

import (
	"bufio"
	"bytes"
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"math"
	rand "math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
)

// notes:
// no mutex is acquired before writing to the maps. this should be implemented in transaction-heavy networks
// bidirectional connections are *required* for the network to work properly
// instead of sorting the peers and connecting to the 10 upper entries, we decided to just connect to 10
// 		random peers instead. otherwise only the lucky few with low port numbers (high on the list) would
// 		really receive connections in large networks. by doing it randomly instead, everyone should be more
// 		or less equally connected to the network.
// everyone is initialized with 100$

var rsakey *rsa.PrivateKey         // our own key
var keys map[string]*rsa.PublicKey // map of all known peers and their public keys
var peers map[string]bool          // map of all known peers and if we are connected to them
var conns map[string]*rpc.Client   // map of all connected peers
var ledger *Ledger
var pastTransactions map[string]bool // transaction id to bools

const debugCalls = true // debug rpc information
const debug = true      // debug information

type Listener int

// merge the peer maps of the caller and the callee
func (l *Listener) MergePeers(request map[string]bool, reply *map[string]bool) error {
	if debugCalls {
		fmt.Println("MergePeers called!")
	}
	*reply = peers
	mergePeers(request)
	return nil
}

// helper method, merges two peer maps
func mergePeers(rpeers map[string]bool) {
	for k, _ := range rpeers {
		_, exists := peers[k]
		if !exists {
			peers[k] = false
		}
	}
	if debug {
		fmt.Println(peers)
	}
}

// merge the key maps of the caller and the callee
func (l *Listener) MergeKeys(request map[string]*rsa.PublicKey, reply *map[string]*rsa.PublicKey) error {
	if debugCalls {
		fmt.Println("MergePeers called!")
	}
	*reply = keys
	mergeKeys(request)
	return nil
}

// helper method, merges two key maps
func mergeKeys(rkeys map[string]*rsa.PublicKey) {
	for k, v := range rkeys {
		keys[k] = v
	}
	if debug {
		fmt.Println("I now know " + fmt.Sprint(len(keys)) + " unique keys")
	}
}

// make the callee broadcast the presence of a new node
func (l *Listener) BroadcastNewNode(request string, reply *bool) error {
	if debugCalls {
		fmt.Println("BroadcastNewNode called!")
	}
	_, exists := peers[request] // do we already know this guy (or gal)
	if !exists {
		peers[request] = false
		broadcastNewNode(request) // if this is a new guy (or gal), tell our friends about him (or her)
	}
	if debug {
		fmt.Println(peers)
	}
	return nil
}

// helper method, broadcasts an ip to all known connections
func broadcastNewNode(ip string) {
	for _, v := range conns {
		var reply bool
		v.Call("Listener.BroadcastNewNode", ip, &reply)
	}
}

// merge the ledger of the caller and the callee
func (l *Listener) MergeLedger(request Ledger, reply *Ledger) error {
	if debugCalls {
		fmt.Println("MergeLedger called!")
	}
	request.lock.Lock()
	ledger.lock.Lock()
	for k, v := range request.Accounts {
		ledger.Accounts[k] = v // assume everything is already synchronized
	}
	request.lock.Unlock()
	ledger.lock.Unlock()
	*reply = *ledger // replace their ledger with ours
	return nil
}

// make the callee perform a transaction
func (l *Listener) MakeSignedTransaction(request SignedTransaction, reply *bool) error {
	if debugCalls {
		fmt.Println("MakeSignedTransaction called!")
	}
	makeSignedTransaction(request)
	return nil
}

// helper method, performs the actual transaction
func makeSignedTransaction(st SignedTransaction) {
	t := st.T
	_, exists := pastTransactions[t.ID]
	if exists {
		return // we have already seen this transaction
	}
	if !validateSignature(st) {
		fmt.Println("Signature was invalid!")
		return // invalid signature
	}
	fmt.Println("Signature was valid!")
	if ledger.Accounts[t.From]-t.Amount < 0 {
		return // insufficient cash
	}
	ledger.lock.Lock()
	ledger.Accounts[t.From] -= t.Amount
	ledger.Accounts[t.To] += t.Amount
	ledger.lock.Unlock()
	pastTransactions[t.ID] = true
	broadcastTransaction(st)
	if debug { // print the updated ledgers
		fmt.Println("New ledger state: ")
		fmt.Println(ledger.Accounts)
	}
}

// validate a given signed transaction
func validateSignature(t SignedTransaction) bool {
	if rsa.VerifyPSS(keys[t.T.From], crypto.SHA256, hashMessage(t.T), t.Signature, nil) == nil {
		return true
	}
	return false
}

// serialize a transaction structure, and hash it
func hashMessage(t Transaction) []byte {
	var buffer = new(bytes.Buffer)
	serializer := json.NewEncoder(buffer)
	serializer.Encode(t) // serialized transaction

	hash := crypto.SHA256
	h := hash.New()
	h.Write(buffer.Bytes())
	hm := h.Sum(nil) // hashed message
	return hm
}

// helper method, broadcast a transaction to all our connections
func broadcastTransaction(t SignedTransaction) {
	for _, v := range conns {
		var reply bool
		v.Call("Listener.MakeSignedTransaction", t, &reply)
	}
}

// make the target connect to the given address. used to ensure bidirectional connections
func (l *Listener) BiConnect(request string, reply *bool) error {
	if debugCalls {
		fmt.Println("BiConnect called!")
	}
	conn, err := rpc.Dial("tcp", request)
	if err == nil {
		fmt.Println("Bidirectional connection established with " + request)
		peers[request] = true
		conns[request] = conn
		if debug {
			fmt.Println(peers)
		}
	} else {
		log.Fatal(err)
	}
	return nil
}

func peer() {
	// initialize all variables
	keys = make(map[string]*rsa.PublicKey)
	peers = make(map[string]bool)
	conns = make(map[string]*rpc.Client)
	pastTransactions = make(map[string]bool)
	ledger = MakeLedger()
	rsakey, _ = rsa.GenerateKey(crand.Reader, 2048)

	// wait for input, and prepare for operation when received
	fmt.Println("Please enter the address of a peer")
	addr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	addr = strings.TrimRight(addr, "\r\n") // os-independent way of removing newline characters
	myaddr := startServer()
	keys[myaddr] = &rsakey.PublicKey // add our own key to the keyset
	connect(formatAddr(addr), myaddr, true)

	// handle transaction input
	fmt.Println("Ready to handle transactions. The format is [port] [amount]. For your convenience, a list of all known ports will be shown after each new transaction.")
	for {
		if debug {
			fmt.Println(peers)
		}
		msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		msg = strings.TrimRight(msg, "\n\r") // remove any trailing characters

		b := make([]byte, 16) // used to generate uuid for the transaction
		rand.Read(b)
		s := strings.Split(msg, " ") // [to, amount]
		v, _ := strconv.Atoi(s[1])   // convert amount to int
		if v < 0 {
			fmt.Println("You cannot send a negative amount!")
			continue
		}
		to := "[::]:" + s[0]
		from := myaddr // we can only send from ourselves (we do not know any other secret keys)
		t := Transaction{ID: fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), From: from, To: to, Amount: v}

		// broadcast the transaction
		signature, _ := rsa.SignPSS(crand.Reader, rsakey, crypto.SHA256, hashMessage(t), nil)
		st := SignedTransaction{T: t, Signature: signature}
		makeSignedTransaction(st)
	}
}

// we have to keep the same format of our addresses, since they are used to uniquely identify peers
func formatAddr(addr string) string {
	addr = strings.ReplaceAll(addr, "localhost", "[::]")
	addr = strings.ReplaceAll(addr, "127.0.0.1", "[::]")
	return addr
}

// connect to a server that may not be active
func connect(remote string, local string, recursive bool) {
	// recursively connect to the targets set of connections
	recConnect := func(connections map[string]bool) {
		ips := getipset(connections) // get the keyset of peers
		n := 0                       // successes
		m := 0                       // attempts
		for float64(n) < math.Min(2.0, float64(len(peers))) && m < 99 {
			r := rand.Intn(len(peers)) // roll a dice
			if !peers[ips[r]] {        // if we are not connected to this guy
				connect(ips[r], local, false) // connect to him non-recursively
				n += 1
			} else {
				m += 1
			}
		}
	}

	conn, err := rpc.Dial("tcp", remote)
	if err == nil {
		peers[remote] = true
		conns[remote] = conn
		remotePeers := make(map[string]bool)          // remote peer set
		remoteKeys := make(map[string]*rsa.PublicKey) // remote key set
		var reply bool
		conn.Call("Listener.BroadcastNewNode", local, &reply)
		conn.Call("Listener.MergePeers", peers, &remotePeers)
		conn.Call("Listener.MergeKeys", keys, &remoteKeys)
		conn.Call("Listener.MergeLedger", ledger, &ledger)
		conn.Call("Listener.BiConnect", local, &reply)
		if recursive {
			recConnect(remotePeers)
		}
		mergePeers(remotePeers)
		mergeKeys(remoteKeys)
		fmt.Println("Connected to " + remote)
	} else {
		fmt.Println("No peer at address")
	}
}

// start our own server on a random port, and return the address
func startServer() string {
	ln, _ := net.Listen("tcp", ":0")
	peers[ln.Addr().String()] = true // by setting our own entry to true, we won't try to connect to it later
	fmt.Println("Server waiting for connection at " + ln.Addr().String())
	ledger.Accounts[formatAddr(ln.Addr().String())] = 100 // initialize our own account

	// handle incoming method calls
	listener := new(Listener)
	rpc.Register(listener)
	go openConnection(ln)
	return formatAddr(ln.Addr().String())
}

// listen for incoming rpc connections
func openConnection(ln net.Listener) {
	fmt.Println("Waiting for connection...")
	rpc.Accept(ln)     // wait for connection
	openConnection(ln) // go wait for further connections
}

func getipset(c map[string]bool) []string {
	var ips []string
	for k, _ := range c {
		ips = append(ips, k)
	}
	return ips
}

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

type Transaction struct {
	ID     string // Any string
	From   string // A verification key coded as a string
	To     string // A verification key coded as a string
	Amount int    // Amount to transfer
}

type SignedTransaction struct {
	T         Transaction // The transaction
	Signature []byte      // Potential signature coded as string
}

func main() {
	peer()
}
