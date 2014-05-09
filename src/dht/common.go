package dht

import "time"
import "crypto/sha1"
import "net/rpc"
import "fmt"
import "strconv"

// Configurable constants
const (
	IDLen = 64
	K = 8
	Alpha = 3
)

const (
	Online = "Online"
	Offline = "Offline"
)
type Status string

const Debug=1

func Print(tag string, format string, a ...interface{}) (n int, err error) {
	tag = "["+tag+"]		"
	if Debug > 0 {
		n, err = fmt.Printf(tag + format + "\n", a...)
	}
	return
}

func Short(id ID) string {
	my_int := int(id)
	if my_int < 0 {
		my_int = my_int * -1
	}
	my_string := strconv.Itoa(my_int)
	return my_string[:4]
}

// const (
// 	OK = "OK"
// 	WrongNodeID = "WrongNodeID"
// )
// type Err string

type RoutingEntry struct {
	IpAddr string
	NodeId ID
}

type RoutingEntryDist struct {
	Distance ID
	RoutingEntry RoutingEntry
}

type ID uint64

type SendMessageArgs struct {
	Content string
	Timestamp time.Time
	ToUsername string
	FromUsername string
}

type SendMessageReply struct {
	
}

type StoreUserArgs struct {
	QueryingNodeId ID
	QueryingIpAddr string
	AnnouncedUsername string
}

type StoreUserReply struct {
	QueriedNodeId ID
}

type FindIdArgs struct {
	QueryingNodeId ID
	QueryingIpAddr string
	TargetId ID
}

type FindIdReply struct {
	TryNodes []RoutingEntryDist // if list is of length 1, then we found it
	TargetIpAddr string
}

type PingArgs struct {
	QueryingNodeId ID
	QueryingIpAddr string
}

type PingReply struct {
	QueriedNodeId ID
}

func Sha1(s string) ID {
	/*
		Returns a 160 bit integer based on a
		string input. 
	*/
    h := sha1.New()
    h.Write([]byte(s))
    bs := h.Sum(nil)
    l := len(bs)
    var a ID
	for i, b := range bs {
	    shift := ID((l-i-1) * 8)	
	    a |= ID(b) << shift
   	}
   	return a
}

func Xor(a, b ID) ID {
	/*
		Zors together two big.Ints and
		returns the result.
	*/
	return a ^ b
}

func find_n(a, b ID) uint{
	var IDLen uint
	IDLen = 64
	var d, diff ID
	diff = a ^ b
	var i uint
	for i = 0; i < IDLen; i++{
		d = 1<<(IDLen - 1 - i)
		if d & diff != 0 { // if true, return i
			return i
		}
	}
	return IDLen - 1
}

func isEqual(entry1 []RoutingEntryDist, entry2 []RoutingEntryDist) bool{
	if len(entry1) != len(entry2){
		return false
	}
	for i, v := range entry1{
		if v != entry2[i] {
			return false
		}
	}
	return true
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
	client, errx := rpc.Dial("tcp", srv)
	if errx != nil {
		return false
	}
	defer client.Close()
		
	err := client.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}