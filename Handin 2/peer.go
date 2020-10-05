package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
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

var peers map[string]bool        // map of all known peers and if we are connected to them
var conns map[string]*rpc.Client // map of all connected peers
var ledger *Ledger
var pastTransactions map[string]bool // transaction id to bools

const debugCalls = false // debug rpc information
const debug = true       // debug information

type Listener int

func (l *Listener) MergePeers(request map[string]bool, reply *map[string]bool) error {
	if debugCalls {
		fmt.Println("MergePeers called!")
	}
	*reply = peers
	merge(request)
	return nil
}

func merge(cmap map[string]bool) {
	for k, _ := range cmap { // Iterating throgh clients map
		_, exists := peers[k] // Check if the key exists in our map
		if !exists {
			peers[k] = false // If it does not, it is added
		}
	}
	if debug {
		fmt.Println(peers)
	}
}

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

func (l *Listener) MakeTransaction(request Transaction, reply *bool) error {
	if debugCalls {
		fmt.Println("MakeTransaction called!")
	}
	makeTransaction(request)
	return nil
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

func makeTransaction(t Transaction) {
	_, exists := pastTransactions[t.ID]
	if exists {
		return // we have already seen this transaction
	}
	if ledger.Accounts[t.From]-t.Amount < 0 {
		fmt.Println(t.From + " has insufficiant balance.")
		return // insufficient cash
	}
	ledger.lock.Lock()
	ledger.Accounts[t.From] -= t.Amount
	ledger.Accounts[t.To] += t.Amount
	ledger.lock.Unlock()
	pastTransactions[t.ID] = true
	broadcastTransaction(t)
	if debug { // print the updated ledgers
		fmt.Println("New ledger state: ")
		fmt.Println(ledger.Accounts)
	}
}

func broadcastTransaction(t Transaction) {
	for _, v := range conns {
		var reply bool
		v.Call("Listener.MakeTransaction", t, &reply)
	}
}

func peer() {
	// Setting up
	peers = make(map[string]bool)
	conns = make(map[string]*rpc.Client)
	pastTransactions = make(map[string]bool)
	ledger = MakeLedger()

	fmt.Println("Please enter the address of a peer")
	addr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	addr = strings.TrimRight(addr, "\r\n") // os-independent way of removing newline characters

	myaddr := startServer()
	connect(formatAddr(addr), myaddr, true)

	// handle input
	fmt.Println("Ready to handle transactions. The format is [port] [port] [amount]. \nFor your convenience, a list of all known ports will be shown after each new transaction.")
	for {
		if debug {
			fmt.Println(peers)
		}
		msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		msg = strings.TrimRight(msg, "\n\r") // remove any trailing characters

		b := make([]byte, 16) // used to generate uuid for the transaction
		rand.Read(b)
		s := strings.Split(msg, " ") // [from, to, amount]
		v, _ := strconv.Atoi(s[2])   // convert amount to int
		from := "[::]:" + s[0]       // since everything is local, it is enough to only input port numbers
		to := "[::]:" + s[1]
		t := Transaction{ID: fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), From: from, To: to, Amount: v}
		makeTransaction(t)
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
		for float64(n) < math.Min(10.0, float64(len(peers))) && m < 99 {
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
	if err == nil { // if succesfull
		peers[remote] = true
		conns[remote] = conn
		cpeers := make(map[string]bool) // connections peers
		var reply bool
		conn.Call("Listener.MergePeers", peers, &cpeers)
		conn.Call("Listener.MergeLedger", ledger, &ledger)
		conn.Call("Listener.BiConnect", local, &reply)
		if recursive {
			recConnect(cpeers)
		}
		merge(cpeers)
		fmt.Println("Connected to " + remote)
	} else {
		fmt.Println("No peer at address")
	}
}

// start our own server on a random port, and return the address
func startServer() string {
	ln, _ := net.Listen("tcp", ":0")             // Listening on random port
	peers[formatAddr(ln.Addr().String())] = true // by setting our own entry to true, we won't try to connect to it later
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

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
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

func main() {
	peer()
}
