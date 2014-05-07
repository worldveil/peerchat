package example

import "net"
import "net/rpc"
import "log"
import "encoding/gob"

type Node struct {
	Address string
	Dead bool
}

type PingArgs struct {
	Message string
}

type PingReply struct {
	OK bool
}

func (node *Node) PingHandler(args *PingArgs, reply *PingReply) error {
	log.Printf("Node (%s) recieved message: %s", node.Address, args.Message)
	reply.OK = true
	return nil
} 

func (node *Node) Ping(address, message string) {
	/*
		Ping another server, include a message
	*/
	ok := false
	args := &PingArgs{Message: message}
	var reply PingReply
	
	for !ok {
		log.Printf("Sending to node %s message: %s", address, message)
		ok = call(address, "Node.PingHandler", args, &reply)
		if !ok {
			log.Printf("Failed! Will try again.")
		}
	}
}

func (node *Node) AsyncPing(address, message string, doneChannel chan *PingReply) {
	ok := false
	args := &PingArgs{Message: message}
	var reply PingReply
	
	for !ok {
		log.Printf("Sending to node %s message: %s", address, message)
		ok = call(address, "Node.PingHandler", args, &reply)
		if !ok {
			log.Printf("Failed! Will try again.")
		}
	}
	
	doneChannel <- &reply
}

func MakeNode(hostname, port string) *Node {
	/*
		Takes a hostname and port as an argument for
		its own address and then creates a Node. 
	*/

	// register which objects RPC can serialize/deserialize
	gob.Register(PingArgs{})
	gob.Register(PingReply{})
	
	// construct our node struct
	address := hostname +":"+ port
	node := &Node{Address: address}
	node.Dead = false
	
	// register the exported methods and
	// create an RPC server
	rpcs := rpc.NewServer()
	rpcs.Register(node)
	
	// set up a connection listener
	l, e := net.Listen("tcp", address)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	
	// spin off go routine to listen for connections
	go func() {
		for !node.Dead {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("listen error: ", err);
			}
			
			// spin off goroutine to handle
			// RPC requests from other nodes
			go rpcs.ServeConn(conn)
		}
		
		log.Printf("Server %s shutting down...", address)
	}()
	
	return node
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

	log.Println(err)
	return false
}

func async(srv string, rpcname string, args interface{}, reply interface{}, doneChannel chan *rpc.Call) bool {
	client, errx := rpc.Dial("tcp", srv)
	if errx != nil {
		return false
	}
	defer client.Close()
	
	// call async, then outer routine will
	// have to wait for doneChannel to reply
	client.Go(rpcname, args, reply, doneChannel)
	
	return true
}
