package dht

import "crypto/sha1"
import "math/big"
import "list"

type User struct {
	Node DhtNode
	Name string
	Message map[string]string
}

type DhtNode struct {
	IpAddr string
	Port int
	NodeId *big.Int # sha1(ip)
	RoutingTable interface{}
	KV map[string]string
}

type AnnouceUserArgs struct {
	QueryingNodeId *big.Int
	QueryingIpAddr string
	AnnoucedUsername string
}

type AnnouceUserReply struct {
	QueriedNodeId *big.int
}

type FindNodeArgs struct {
	QueryingNodeId *big.Int
	TargetNodeId *big.Int
}

type FindNodeReply struct {
	QueriedNodeId *big.Int
	TryNodes string[] // if list is of length 1, then we found it
}

type GetUserArgs struct {
	QueryingNodeId *big.Int
	TargetUsername *big.Int
}

type GetUserReply struct {
	QueriedNodeId *big.Int
	TryNodes string[] // if list is of length 1, then we found it
}

type PingArgs struct {
	PingingNodeId *big.Int
}

type PingReply struct {
	PingedNodeId *big.Int
}

//=======

func (node *DhtNode) AnnouceUser(AnnouceUserArgs *args, AnnouceUserReply *reply) {
}

func (node *DhtNode) FindNode(FindNodeArgs *args, FindNodeReply *reply) {
}

func (node *DhtNode) GetUser(GetUserArgs *args, GetUserReply *reply) {
}

func (node *DhtNode) Ping(PingArgs *args, PingReply *reply) {
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
