package main

import "fmt"
import "bufio"
import "os"
import "time"

func main() {	
	startChat()
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
