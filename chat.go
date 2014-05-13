package main

import "fmt"
import "bufio"
import "os"
import "dht"
import "io/ioutil"
import "time"
import "github.com/pmylund/sortutil"
import "path/filepath"

func main() {	
	startChat()
	//testsort()
}

type Thing struct {
	Field1 int
	Field2 int64
}

func testsort() {
	slice := make([]Thing, 0)
	for i := 0; i < 19; i++ {
		thing := Thing{Field1: i, Field2: time.Now().Unix()}
		slice = append(slice, thing)
		//time.Sleep(1001 * time.Millisecond)
	}
	
	//print in order
	sortutil.AscByField(slice, "Field1")
	for i := 0; i < len(slice); i++ {
		fmt.Printf("thing: %+v\n", slice[i])
	}
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
    
    // update loop
    go func() {
    	notifications := user.GetNotificationsChannel()
    	for {
	    	<- notifications
	    	paint(user)
	    }
    }()
    
    // input loop
    peer := ""
    for {
    	// State 1) get a user to chat with
    	if peer == "" {
    		fmt.Printf("User to chat with: ")
    		peer = input(reader)
    		fmt.Printf("Starting to talk to: %s\n", peer)
	    	user.UpdateCurrentPeer(peer)
    		fmt.Printf("me> ")
    	
    	// State 2) continue chatting
    	} else {
	    
	    	text := input(reader)
	    	
	    	if text == "" {
	    		// do nothing
	    		fmt.Printf("%s> \n", peer)
	    		paint(user)
	    	
	    	} else if text[0] == 92 {
	    		// switching users to chat with
	    		peer = text[1:]
	    		fmt.Printf("Swtiching to talk to: %v\n", peer)
	    		user.UpdateCurrentPeer(peer)
	    		fmt.Printf("me> ")
	    		paint(user)
	    		
	    		
	    	} else if text == "exit" {
	    		// exit peerchat
	    		fmt.Printf("Exiting Peerchat!\n")
	    		user.Logoff()
	    		break
	    		
	    	} else {
	    		// send the message!
	    		user.SendMessage(user.Current, text)
	    		paint(user)
	    	}
	    }
    }
}

func paint(user *dht.User) {

	// clear space
	for j := 1; j < 100; j++ {
		fmt.Println("")
	}
	
	// new messages?
	usersWithPendingMessages := make([]string, 0)
	for peer, _ := range user.MessageHistory {
		areNew, _ := user.AreNewMessagesFrom(peer)
		if areNew {
			usersWithPendingMessages = append(usersWithPendingMessages, peer)
		}
	}
	
	// print users with pending messages
	for _, peer := range usersWithPendingMessages {
		fmt.Printf("New message(s) from `%s`!\n", peer)
	}
	fmt.Printf("\n\n========================================\n")
	
	// are we current chatting?
	if user.Current != "" {
		fmt.Printf("Conversation with `%s`:\n\n", user.Current)
		
		newMessages := user.AllMessagesFromUser(user.Current)
		
		messages := make([]dht.SendMessageArgs, 0)
		for _, msg := range newMessages {
			messages = append(messages, *msg)
		}
		
		sortutil.AscByField(messages, "Timestamp")
		for i := 0; i < len(messages); i++ {
			msg := messages[i]
			fmt.Printf("%s> %s\n", msg.FromUsername, msg.Content)
		}
	}
	
	fmt.Printf("=========================================\n")
	fmt.Printf("me> ")
}