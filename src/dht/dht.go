package dht

import "crypto/sha1"
import "math/big"

type Peer struct {
	NodeID *big.Int
	Address int
	Port int
	Table DHT
}

type DHT struct {
}

func (dht *DHT) Ping() {
}

func (dht *DHT) Store() {
}

func (dht *DHT) FindNode() {
}

func (dht *DHT) FindValue() {
}

func MakePeer() {
	
}

func Sha1(s string) *big.Int {
	/*
		Returns a 160 bit integer based on a
		string input. 
	*/
    h := sha1.New()
    h.Write([]byte(s))
    bs := h.Sum(nil)
    bi := new(big.Int).SetBytes(bs)
    return bi
}

func Xor(a, b *big.Int) *big.Int {
	/*
		Zors together two big.Ints and
		returns the result.
	*/
	return new(big.Int).Xor(a, b)
}
