package main

import "example"
import "log"
import "strconv"
import "fmt"
import "bufio"
import "os"
import "time"

func main() {	
	
	//sync()
	async()
	//startChat()
	
	// table()
}

type Entry struct {
	Thing string
}

func table() {
	var t [10][]Entry
	for i, _ := range t {
		t[i] = make([]Entry, 0)
		t[i] = append(t[i], Entry{Thing:strconv.Itoa(i)})
	}
	log.Printf("%+v", t)
}

func sync() {
	one := example.MakeNode("127.0.0.1", "55555")
	log.Printf("making one...")
	two := example.MakeNode("127.0.0.1", "55554")
	log.Printf("making two...")
	
	one.Ping(two.Address, "Hello, this is one!")
	two.Ping(one.Address, "This is two...")
	one.Ping(two.Address, "Cool")
	two.Ping(one.Address, "Last message.")
}

func async() {
	
	one := example.MakeNode("127.0.0.1", "55555")
	log.Printf("making one...")
	two := example.MakeNode("127.0.0.1", "55554")
	log.Printf("making two...")
	
	alpha := 25
	
	doneChannel := make(chan *example.PingReply, alpha)
	for i := 0; i < alpha; i++ {
		go one.AsyncPing(two.Address, "Async message " + strconv.Itoa(i), doneChannel)
	}
	
	for i := 0; i < alpha; i++ {
		reply := <-doneChannel
		log.Printf("Message recieved? %v", reply.OK)
	} 
}

func startChat() {
	fmt.Printf("********************")
	fmt.Printf(" WELCOME TO PEERCHAT! ")
	fmt.Printf("********************\n")
	
	// input a friends IP address 
	// TODO: logic for if you've already registered
	fmt.Printf("Enter a friend's address (ip:port) to get started:\n")
	in := bufio.NewReader(os.Stdin)
    peerAddress, err := in.ReadString('\n')
    peerAddress = peerAddress[:len(peerAddress)-1]
    if err != nil {
    	// handle error
    }
    
    // so sexy
    fmt.Printf("Connecting to %s", peerAddress)
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".")
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".")
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".\n")
    
    for {
    	fmt.Printf("%s> ", peerAddress)
    	input, _ := in.ReadString('\n')
    	input = input[:len(input)-1]
    	
    	if input == "exit" {
    		fmt.Printf("Signing out!\n")
    		break
    		
    	} else if input[0] == 92 { // "\"
    		fmt.Printf("Swtiching to talk to: %s\n", input[1:])
    		peerAddress = input[1:]
    		
    	} else {
    		fmt.Printf("Sending: \"%s\"...\n", input)
    	}
    }
}
