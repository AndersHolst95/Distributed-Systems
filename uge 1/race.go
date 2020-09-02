package main

import (
	"fmt"
	"sync"
)

type DNS struct {
	m    map[string]string
	lock sync.RWMutex
}

func (dns *DNS) Set(key string, val string) {
	dns.lock.Lock()
	defer dns.lock.Unlock()
	dns.m[key] = val
}

func (dns *DNS) Get(key string) string {
	dns.lock.RLock()
	defer dns.lock.RUnlock()
	return dns.m[key]
}

func MakeDNS() *DNS {
	dns := new(DNS)
	dns.m = make(map[string]string)
	return dns
}

func GetAndSet(suf string) {
	for i := 0; i < 10; i++ {
		dns.Set("X", dns.Get("X")+suf)
	}
	c <- 0
}

var c = make(chan int)
var dns = MakeDNS()

func main() {
	go GetAndSet("1")
	go GetAndSet("2")
	go GetAndSet("3")
	<-c
	<-c
	<-c
	fmt.Println(dns.Get("X"))
}
