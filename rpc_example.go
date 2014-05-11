package main

import "fmt"
import "bufio"
import "os"
import "dht"
import "io/ioutil"
import "time"

import "path/filepath"

func main() {	
	startChat()
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
} 
func input(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
    input = input[:len(input)-1]
    return input
}

func startChat() {
	reader := bufio.NewReader(os.Stdin)
	
	// load the Peerchat banner
	content, err := ioutil.ReadFile("peerchat.txt")
	if err != nil {
	    //Do something
	}
	fmt.Println(string(content))
	
	// get username
	fmt.Printf("Enter username: ")
	username := input(reader)
	fmt.Printf("Enter IP Address (xxx.xxx.xxx.xxx:yyyy): ")
	address := input(reader)
	
	// search for this user's history
	files, _ := filepath.Glob("/tmp/*.gob")
	userfile := ""
	user := new(dht.User)
    for _, file := range files {
    	if file == dht.UsernameToPath(username) {
    		userfile = file
    	}
    }
    
    // did we find the file?
    if userfile != "" {
    	// we found this user, load from file
    	user = dht.Login(username, address)
    	
    } else {
    	// we did not find a matching user
		fmt.Printf("\n[*] Looks like you haven't logged in on this computer before! Would you like to create a new network, or join an existing one?\n")
		fmt.Printf("Join existing? Type (Y/N):")
		join := input(reader)
		
		if join == "Y" || join == "y" {
			// join existing network, ask for boostrap IP
			fmt.Printf("\n[*] To join an existing network, please enter an IP address/port of a friend (xxx.xxx.xxx.xxx:yyyy): ")
			boostrapAddress := input(reader)
			user = dht.RegisterAndLogin(username, address, boostrapAddress)
			
		} else {
			// creating a new network, simply create new user
			// use bogus address and start with empty routing table
			user = dht.RegisterAndLogin(username, address, "")
			
		}
    }	
    
    // we're now logged in with a user
    fmt.Printf("Connecting to Peerchat")
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".")
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".")
    time.Sleep(300 * time.Millisecond)
    fmt.Printf(".\n")
    
    // prompt them to chat
    fmt.Printf("Connected as user: %+v\n", user)
    
    // never ending loop
    peer := ""
    newPeer := ""
    for {
    	
    	// State 1) get a user to chat with
    	if peer == "" {
    		fmt.Printf("User to chat with: ")
    		newPeer = input(reader)
    		
    		// verify this user is online
    		if ! user.IsOnline(newPeer) {
    			continue
    		}
    		
    		// if they are, set peer to new peer
    		peer = newPeer
    		fmt.Printf("%s> ", peer)
    	
    	// State 2) continue chatting
    	} else {
	    
	    	text := input(reader)
	    	
	    	if text == "" {
	    		// do nothing
	    		fmt.Printf("%s> \n", peer)
	    	
	    	} else if text[0] == 92 {
	    		// switching users to chat with
	    		newPeer = text[1:]
	    		
	    		// verify this user `newPeer` is online
	    		fmt.Printf("Attempting to connect to %s...\n", newPeer)
	    		if ! user.IsOnline(newPeer) {
	    			fmt.Printf("Attempting to connect to %s...\n", newPeer)
	    			continue
				}
	    		
	    		// if they are, set peer to new peer
	    		peer = newPeer
	    		fmt.Printf("Swtiching to talk to: %s\n", peer)
	    		fmt.Printf("%s> ", peer)
	    		
	    	} else if text == "exit" {
	    		// exit peerchat
	    		fmt.Printf("Exiting Peerchat!\n")
	    		break
	    		
	    	} else {
	    		// send the message!
	    		user.SendMessage(peer, text)
	    		
	    		// maybe some indication of whether they are 
	    		// offline and that the message will be sent later?
	    		// ...
	    	}
	    }
    }
}
