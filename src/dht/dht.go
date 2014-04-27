package dht

import "crypto/sha1"
import "math/big"

type User struct {
	Node DhtNode
	Name string
}

type DhtNode struct {
	# we store the username => ip addr
	# sha1(username)
	Hostname string
	Port int
	NodeId *big.Int # sha1(hostname)
	Table interface{}
}

type AnnouceUserArgs struct {
	QueryingNodeId *big.Int
	QueryingHostname string
	AnnoucedUsername string
}

type AnnouceUserReply struct {
	QueriedNodeId *big.int
}

type FindNodeArgs struct {
	QueryingNodeId *big.Int
	TargetNodeId *big.Int
}

type FindNodeArgs struct {
	
}

//=======

func (node *DhtNode) AnnouceUser(AnnouceUserArgs *args, AnnouceUserReply *reply) {
}

func (node *DhtNode) FindNode() {
}

func (node *DhtNode) GetUser() {
}

//=========

func MakeNode(host string, port int, username string) {
	nodeid := Sha1(username)
	table := &DHT{}
	peer := Peer{Address: host, Port: port, NodeID: nodeid, Table: table}
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
