package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"
)

var k = 16          // maximum bitlength of n
const debug = false // print debug information
var e *big.Int
var d *big.Int
var n *big.Int
var p *big.Int
var q *big.Int

// find the greatest common divisor of the two arguments
func GCD(p, q *big.Int) int64 {
	a, b := new(big.Int).Set(p), new(big.Int).Set(q)
	for b.Int64() != 0 {
		t := new(big.Int).Set(b)
		b.Mod(a, b)
		a = new(big.Int).Set(t)
	}
	return a.Int64()
}

// generate everything we need for encryption
func KeyGen(k int) {
	e = new(big.Int).SetInt64(3)
	coprime := false
	temp := new(big.Int).SetInt64(1) // this is just 1

	// find two primes that are both coprime with e
	for !coprime {
		p, _ = rand.Prime(rand.Reader, int(k/2)) // max number is 2^(k/2)
		q, _ = rand.Prime(rand.Reader, int(k/2)) // max number is 2^(k/2)

		// if they are equal, try again
		if p.String() == q.String() {
			continue
		}

		gcdp := GCD(new(big.Int).Sub(p, temp), e)
		gcdq := GCD(new(big.Int).Sub(q, temp), e)

		if debug {
			fmt.Println("p: " + p.String() + " gcd: " + fmt.Sprint(gcdp))
			fmt.Println("q: " + q.String() + " gcd: " + fmt.Sprint(gcdq))
		}

		// if they are both coprime with e, stop searching
		if gcdp == 1 && gcdq == 1 {
			coprime = true
		}
	}

	n = new(big.Int).Mul(p, q)                                                   // max number is 2^k --> k bits
	pq := new(big.Int).Mul(new(big.Int).Sub(p, temp), new(big.Int).Sub(q, temp)) // (p-1)(q-1)

	d = new(big.Int).ModInverse(e, pq) // d*e mod pq = 1
	if debug {
		fmt.Println("p: " + p.String())
		fmt.Println("q: " + q.String())
		fmt.Println("pq: " + pq.String())
		fmt.Println("d: " + d.String())
	}
}

// encrypt a given integer
func Encrypt(m int64) *big.Int {
	a := new(big.Int).SetInt64(m)
	return new(big.Int).Exp(a, e, n)
}

// decrypt a given ciphertext
func Decrypt(c *big.Int) int64 {
	return new(big.Int).Exp(c, d, n).Int64()
}

func hash(m int64) *big.Int { // Hashes the message
	s := strconv.FormatInt(m, 10)
	f := sha256.Sum256([]byte(s))
	return new(big.Int).SetBytes(f[:])
}

func HSign(m int64) int64 { // Hash and sign
	return Decrypt(hash(m))
}

func Sign(m *big.Int) int64 { // Sign is equal to a decryption
	return Decrypt(m)
}

func Verify(s int64, m int64) bool {
	hm := new(big.Int).Mod(hash(m), n).Int64()
	hs := Encrypt(s).Int64() // Verify is equal to an encryption
	return hm == hs
}

func RSATest() {
	fmt.Println("Generating keys for k = " + fmt.Sprint(k))
	fmt.Println("n has size: " + fmt.Sprint(n.BitLen()))
	m := int64(17)
	fmt.Println("m is: " + fmt.Sprint(m))
	c := Encrypt(m)
	fmt.Println("c is: " + c.String())
	m = Decrypt(c)
	fmt.Println("Decrypted m is: " + fmt.Sprint(m))
}

func VerifyTest() {
	m := int64(132645312)
	fmt.Println("The message is " + fmt.Sprint(m))
	s := HSign(m)
	b := Verify(s, m)
	fmt.Println("The verification of the signature and the message is " + fmt.Sprint(b))
	m = int64(64)
	fmt.Println("Changing the message to " + fmt.Sprint(m))
	b = Verify(s, m)
	fmt.Println("The verification of the signature and the message is " + fmt.Sprint(b))
}

func HashSpeedTest() {
	m, _ := ioutil.ReadFile("hashtext.txt")
	ta := time.Now()
	sha256.Sum256([]byte(m))
	t := time.Since(ta).Seconds()
	info, _ := os.Stat("hashtext.txt")
	x := info.Size() * 8 // convert bytes to bits
	fmt.Println("The size of the file is " + fmt.Sprint(x) + " bits")
	fmt.Println("The time required to hash this file is " + fmt.Sprint(t) + "s")
	fmt.Println("Hash speed is " + fmt.Sprintf("%.3e", float64(x)/t) + " bits/s")
}

func SignSpeedTest() {
	KeyGen(2000)
	m := int64(12345)
	hm := hash(m)
	ta := time.Now()
	Sign(hm)
	t := time.Since(ta).Seconds()
	fmt.Println("It took " + fmt.Sprint(t) + " seconds to sign this message, using a 2000 bit RSA key")
}

func main() {
	KeyGen(k)
	//	RSATest()
	fmt.Println("Starting VerifyTest")
	VerifyTest()
	fmt.Println("\nStarting HashSpeedTest")
	HashSpeedTest()
	fmt.Println("\nStarting SignSpeedTest")
	SignSpeedTest()
}
