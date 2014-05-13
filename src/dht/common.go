package dht

import "crypto/sha1"
import "net/rpc"
import "fmt"
import "strconv"
import "crypto/rand"
import "math/big"
import "os"
import "time"

// Configurable constants
const (
	IDLen = 64
	K = 20
	Alpha = 3
)

const (
	Online = "Online"
	Offline = "Offline"
)

const Debug=0

func Print(tag string, format string, a ...interface{}) (n int, err error) {
	tag = "["+tag+"]		"
	if Debug > 0 {
		n, err = fmt.Printf(tag + format + "\n", a...)
	}
	return
}

func Short(id ID) string{
	my_string := strconv.FormatUint(uint64(id), 10)
	// fmt.Println(my_string)
	return my_string[:4]
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

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
	Timestamp int64
	ToUsername string
	FromUsername string
	MessageIdentifier int64
}

type SendMessageReply struct {
	
}

type StoreUserArgs struct {
	QueryingNodeId ID
	QueryingIpAddr string
	AnnouncedUserId ID
	AnnouncedIpAddr string
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
	QueriedNodeId ID
	QueriedIpAddr string
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

func removeDuplicates(slice []RoutingEntryDist) []RoutingEntryDist{
	uniques := make(map[RoutingEntryDist]bool)
	for _, entryDist := range slice{
		uniques[entryDist] = true
	}
	new_slice := []RoutingEntryDist{}
	for key, _ := range uniques{
		new_slice = append(new_slice, key)
	}
	return new_slice
}

func moveToEnd(slice []RoutingEntry, index int) []RoutingEntry{
	return append(slice[:index], append(slice[index + 1:], slice[index])...)
}

func appendToCsv(filename, text string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
	    panic(err)
	}
	defer f.Close()
	
	if _, err = f.WriteString(text); err != nil {
	    panic(err)
	}
}

// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
func call(srv string, rpcname string, args interface{}, reply interface{}) bool {

	// collect data
	/*
	n := 640
	text := fmt.Sprintf("%s, %d, %d, %d, %d\n", rpcname, time.Now().Unix(), K, Alpha, n)
	filename := fmt.Sprintf("/Users/will/Code/Go/peerchat/writeup/plots/SWEEP.csv")
	appendToCsv(filename, text)
	*/

	c:= make(chan bool, 1)
	go func() { 
		client, errx := rpc.Dial("tcp", srv)
		if errx != nil {
			c <- false
			return
		}
		defer client.Close()
			
		err := client.Call(rpcname, args, reply)
		if err == nil {
			c <- true
			return
		}
	}()

	select {
	case result := <- c:
		return result
	case <- time.After(time.Second *2):
		return false
	}
}