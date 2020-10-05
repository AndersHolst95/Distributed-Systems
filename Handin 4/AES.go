package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"os"
)

var iv []byte
var blocks cipher.Block

// create a new file with the given name and store the ciphertext of m in it
func EncryptToFile(file string, m []byte) {
	c := make([]byte, len(m))        // prepare an array for the ciphertext
	ctr := cipher.NewCTR(blocks, iv) // create the 'O' blocks as indicated in the slides
	ctr.XORKeyStream(c, m)           // XOR these blocks with the message blocks, and output to the ciphertext array

	f, _ := os.Create(file)  // create a new file. if it already exists, it is overwritten
	f.WriteString(string(c)) // write the ciphertext
	f.Close()
}

// decrypt the ciphertext from the given file
func DecryptFromFile(file string) string {
	c, _ := ioutil.ReadFile(file) // read the ciphertext from the file
	md := make([]byte, len(c))    // prepare an array for the message

	ctr := cipher.NewCTR(blocks, iv) // create the 'O' blocks as indicated in the slides
	ctr.XORKeyStream(md, c)          // XOR these blocks with the message blocks, and output to the message array
	return string(md)
}

func main() {
	key := []byte("1122334455667788") // 16 bytes
	m := []byte("this is a test. i want to be encrypted and written to a file. 123456789")
	iv = make([]byte, len(key))    // the iv is required for this to work, so we just supply one filled with zeros.
	blocks, _ = aes.NewCipher(key) // prepare the encryption algorithm

	EncryptToFile("test.txt", m)
	md := DecryptFromFile("test.txt")
	fmt.Println(md)
}
