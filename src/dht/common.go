package dht

import "time"
import "math/big"
import "crypto/sha1"
import "net/rpc"
import "fmt"

const (
	Online = "Online"
	Offline = "Offline"
)
type Status string

type ID *big.Int

type SendMessageArgs struct {
	Content string
	Timestamp time.Time
	ToUsername string
	FromUsername string
}

type SendMessageReply struct {
	
}

type AnnouceUserArgs struct {
	QueryingNodeId *big.Int
	QueryingIpAddr string
	AnnoucedUsername string
}

type AnnouceUserReply struct {
	QueriedNodeId *big.Int
}

type FindNodeArgs struct {
	QueryingNodeId *big.Int
	TargetNodeId *big.Int
}

type FindNodeReply struct {
	QueriedNodeId *big.Int
	TryNodes []string // if list is of length 1, then we found it
}

type GetUserArgs struct {
	QueryingNodeId *big.Int
	TargetUsername *big.Int
}

type GetUserReply struct {
	QueriedNodeId *big.Int
	TryNodes []string // if list is of length 1, then we found it
}

type PingArgs struct {
	PingingNodeId *big.Int
}

type PingReply struct {
	PingedNodeId *big.Int
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

// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string, args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("unix", srv)
	if errx != nil {
		return false
	}
	defer c.Close()
		
	err := c.Call(rpcname, args, reply)
	if err == nil {
		return true
	} 

	fmt.Println(err)
	return false
}