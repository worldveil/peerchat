package dht

import "crypto/sha1"
import "math/big"
import "list"


type DhtNode struct {
	IpAddr string
	Port int
	NodeId *big.Int # sha1(ip)
	RoutingTable interface{}
	KV map[string]string
}

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