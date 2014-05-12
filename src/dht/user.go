package dht

import "time"
import "os"
import "encoding/gob"
import "sync"
import "net"
import "log"
import "fmt"
import "github.com/pmylund/sortutil"

const UserTag = "USER"
const SendingTag = "SENDING"

type User struct {
	l net.Listener
	mu sync.Mutex
	Node *DhtNode
	Name string
	MessageHistory map[string][]*SendMessageArgs // username => messages we've gotten so far
	PendingMessages map[string][]*SendMessageArgs // username => slice of pending messages to apply
	ReceivedMessageIdentifiers map[int64]bool // messageIdentifier (int64) => true if seen messageIdentifier before
	//Notifications map[string]chan *SendMessageArgs
	dead bool
}

const PEERCHAT_USERDATA_DIR = "/tmp"
const PERSIST_EVERY = 30

func UsernameToPath(username string) string {
	/*
		Given a username, returns the filepath
		where the backup will be located.
		
		NOTE: Go treats "/" as the path separator on
		all platforms.
	*/
	return PEERCHAT_USERDATA_DIR + "/" + username + ".gob"
}

func (user *User) GetMessagesFrom(other *User) []*SendMessageArgs {
	/*
		Returns the list of SendMessageArgs
	*/
	if _, ok := user.MessageHistory[other.Name]; ok {
		return user.MessageHistory[other.Name]
	}
	return make([]*SendMessageArgs, 0)
}

func (user *User) Serialize() {
	/*
		Serializes this User struct.
	*/
	path := UsernameToPath(user.Name)
	Print(UserTag, "Serializing path=%s for User %+v", path, user)
	encodeFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	// encode and write to file
	encoder := gob.NewEncoder(encodeFile)
	if err := encoder.Encode(user); err != nil {
		panic(err)
	}
	encodeFile.Close()
	Print("USER", "Written to file successfully")
}

func Deserialize(username string) (bool, *User) {
	/*
		Deserializes a DhtNode and loads
		it into a new DhtNode, which is 
		returned. 
	*/
	newUser := new(User)
	
	// check if this file exists
	path := UsernameToPath(username)
	Print(UserTag, "Loading user from path=%s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// this file does not exist
		Print(UserTag, "File %s does not exist!", path)
		return false, newUser
	}
	
	decodeFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer decodeFile.Close()

	// create decoder
	decoder := gob.NewDecoder(decodeFile)
	decoder.Decode(&newUser)
	return true, newUser
}

// might return nil- handled by Application
func Login(username string, userIpAddr string) *User {
	/*
		Attempts to log into the Peerchat network by loading a previous configuration
		and defaulting to creating a new one. 
	*/
	
	Print(UserTag, "Attempting to log on with username=%s and ip=%s...", username, userIpAddr) 
	user := loadUser(username, userIpAddr)
	if user != nil {
		user.setupUser()
		time.Sleep(10*time.Millisecond)
		user.Node.AnnounceUser(username, userIpAddr)
		go user.startSender()
		go user.startPersistor()
	}
	return user
}

func RegisterAndLogin(username string, userIpAddr string, bootstrapIpAddr string) *User { 
	/*
		Attempts to register as a new user on the Peerchat network using 
		a known IP address as a boostrap. 
	*/
	Print(UserTag, "Bootstraping register with %s using username=%s, ip=%s, and ...", bootstrapIpAddr, username, userIpAddr) 
	user := MakeUser(username, userIpAddr)
	user.setupUser()
	
	// check status of user we are about to bootstrap from
	status := user.CheckStatus(bootstrapIpAddr)
	if status == Offline {
		Print(UserTag, "Could not boostrap: %s was not online!", bootstrapIpAddr)
		//return some error
	}

	time.Sleep(10*time.Millisecond)
	user.Node.AnnounceUser(username, userIpAddr)
	go user.startSender()
	go user.startPersistor()
	return user
}

func (user *User) Logoff() {
	user.Serialize()
	user.dead = true
	user.l.Close()
}

func (user *User) setupUser(){
	rpcs := user.Node.SetupNode()
	rpcs.Register(user)

	// set up a connection listener
	l, e := net.Listen("tcp", user.Node.IpAddr)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	user.l = l
	
	// spin off go routine to listen for connections
	go func() {
		Print(StartTag, "Connection listener for %s starting...", user.Node.IpAddr)
		for user.dead == false{
			conn, err := l.Accept()
			if err == nil && ! user.dead{
				// spin off goroutine to handle
				// RPC requests from other nodes
				go rpcs.ServeConn(conn)
			} else if err == nil {
				conn.Close()
			}
			if err != nil && ! user.dead{
				fmt.Println(err)
				user.Logoff()
			}			
		}
		
		Print(StartTag, "!!!!!!!!!!!!!!!!!! Server %s shutting down...", user.Node.IpAddr)
		fmt.Println("Server shutting down")
	}()
}

func MakeUser(username string, ipAddr string) *User{
	Print(UserTag, "Creating a new User...")
	emptyPendingMessages := make(map[string][]*SendMessageArgs)
	history := make(map[string][]*SendMessageArgs)
	receivedMessageIdentifiers := make(map[int64]bool)
	//notifications := make(map[string]chan *SendMessageArgs)
	
	node := MakeNode(username, ipAddr)
	user := &User{Node: node, Name: username, PendingMessages: emptyPendingMessages, MessageHistory: history, ReceivedMessageIdentifiers: receivedMessageIdentifiers } //, Notifications: notifications}
	return user
}

func loadUser(username, myIpAddr string) *User {
	/*
		This method loads the User struct for a given 
		username from disk, checking if there needs to 
		be a reconfiguration of the routing table or not
		and acting appropriately. 
	*/
	
	// first deserialize the old User struct from disk
	success, user := Deserialize(username)
	
	// there was a userfile to load
	if success {
	
		Print(UserTag, "Loaded User from disk!")
	
		// check and see if ipaddr is the same as the old one
		// if so, we don't need to change anything
		if user.Node.IpAddr != myIpAddr {
			
			// otherwise, create a new nodeId
			user.Node.NodeId = Sha1(myIpAddr)
			user.Node.IpAddr = myIpAddr

			Print(UserTag, "IP Address changed to %s, creating new NodeID=%s", myIpAddr, user.Node.NodeId)
			
			// and rearrange the table based on new nodeId
			// first, get a list of all (nodeId, ipAddr) pairs
			routingEntries := make([]RoutingEntry, 0)
			
			// for each k-buckets row
			for _, row := range user.Node.RoutingTable {
			
				// for each RoutingEntry in row
				for _, entry := range row {
					routingEntries = append(routingEntries, entry)
				}
			}
			
			// now delete old routing table and replace 
			// with a new (empty) one
			user.Node.MakeEmptyRoutingTable()
			
			// then, for each pair, call:
			// updateRoutingTable(nodeId ID, IpAddr string)
			for _, entry := range routingEntries {
				if user.Node.Ping(entry) {
					Print(UserTag, "RoutingEntry %+v is online, updating routing table...", entry)
					user.Node.updateRoutingTable(entry)
				}			
			}
		}

	// there was not, create a new User	
	} else {
		user = nil
	}
	
	return user
}

//SendMessage RPC Handler
func (user *User) SendMessageHandler(args *SendMessageArgs, reply *SendMessageReply) error {
	
	user.mu.Lock()
	defer user.mu.Unlock()

	// check if message is for you, and you havn’t received it before -> then process
	if args.ToUsername == user.Name{
		_, seenBefore := user.ReceivedMessageIdentifiers[args.MessageIdentifier]
		if ! seenBefore{
			Print(UserTag, "%s recieved a previously unseen message meant for me!: %s, from %s at %v", user.Name, args.Content, args.FromUsername, args.Timestamp)
			//initialize entry in messageHistory if first time hearing from user
			if _, ok := user.MessageHistory[args.FromUsername]; !ok {
				user.PendingMessages[args.FromUsername] = make([]*SendMessageArgs, 0)
			}
			user.MessageHistory[args.FromUsername] = append(user.MessageHistory[args.FromUsername], args)
			user.ReceivedMessageIdentifiers[args.MessageIdentifier] = true
			
			// then notify the UI
			/*
			if _, ok := user.Notifications[args.FromUsername]; !ok {
				user.Notifications[args.FromUsername] = make(chan *SendMessageArgs, 10000)
			}
			user.Notifications[args.FromUsername] <- args
			*/
			
		} else {
			Print(UserTag, "%s recieved a previously seen message meant for me! Disregarding: %s, from %s at %v", user.Name, args.Content, args.FromUsername, args.Timestamp)
		}
	} else {
		// if not for you -> store in pendingMessages map
		if _, ok := user.PendingMessages[args.ToUsername]; !ok {
			user.PendingMessages[args.ToUsername] = make([]*SendMessageArgs, 0)
		}
		user.PendingMessages[args.ToUsername] = append(user.PendingMessages[args.ToUsername], args)
	}
	
	// persist to disk
	user.Serialize()
	return nil
}

//SendMessage API
func (user *User) SendMessage(username string, content string) {
	/*
		Sends message with content to username. In offline case, we save the message for later
	*/
	user.mu.Lock()
	Print(UserTag, "Queuing message \"%s\" to %s", content, username)
	//initilize map entry for a certain user
	if _, ok := user.PendingMessages[username]; !ok {
		user.PendingMessages[username] = make([]*SendMessageArgs, 0)
	} 
	pendingMessage := &SendMessageArgs{Content: content, Timestamp: time.Now(), ToUsername: username, FromUsername: user.Name, MessageIdentifier: nrand()}
	user.PendingMessages[username] = append(user.PendingMessages[username], pendingMessage)
	user.mu.Unlock()
}

func (user *User) startPersistor() {
	/*
		Saves the user's routing table and message
		history to disk every PERSIST_EVERY seconds.
	*/
	for {
		user.Serialize()
		time.Sleep(PERSIST_EVERY * time.Second)
	}
}

func (user *User) startSender() {
	/*
		A separate thread which waits until Nodes 
		are up to send them messages
	*/
	Print(SendingTag, "Sender process for %s starting...", user.Name)
	for user.dead == false{
		for username, _ := range user.PendingMessages {
			for len(user.PendingMessages[username]) > 0 {
				
				ip := user.Node.FindUser(username)
				status := user.CheckStatus(ip)
				if status == Online {
					
					// pop first message args off of slice
					user.mu.Lock()
					var args SendMessageArgs
					if len(user.PendingMessages[username]) == 1 { 
						args, user.PendingMessages[username] = *user.PendingMessages[username][0], make([]*SendMessageArgs, 0)
					} else if len(user.PendingMessages[username]) > 1 {
						// "Pop(0)" slice operation taken from:
						// https://code.google.com/p/go-wiki/wiki/SliceTricks
						args, user.PendingMessages[username] = *user.PendingMessages[username][0], user.PendingMessages[username][1:]
					}
					user.mu.Unlock()
					
					// create reply and send to user
					var reply SendMessageReply
					Print(SendingTag, "SenderLoop: Sending \"%s\" to %s...", args.Content, args.ToUsername)
					ok := call(ip, "User.SendMessageHandler", args, &reply)
					
					// if our message sending failed, put back on queue
					if ! ok {
						// put message back on front of queue and continue
						// "Insert" slice operation taken from:
						// https://code.google.com/p/go-wiki/wiki/SliceTricks
						user.mu.Lock()
						user.PendingMessages[username] = append(
							user.PendingMessages[username][:0], 
							append([]*SendMessageArgs{&args}, user.PendingMessages[username][0:]...)...)
						//forward to K nearest neighbors
						kClosestEntryDists := user.Node.FindNearestNodes(Sha1(username))
						for _, entryDist := range kClosestEntryDists {
							var replyOther SendMessageReply
							call(entryDist.RoutingEntry.IpAddr, "User.SendMessageHandler", args, &replyOther)
						}
						user.mu.Unlock()
						
						// create reply and send to user
						var reply SendMessageReply
						Print(SendingTag, "SenderLoop: Sending \"%s\" to %s...", args.Content, args.ToUsername)
						ok := call(ip, "User.SendMessageHandler", args, &reply)
						
						// if our message sending failed, put back on queue
						if ! ok {
							// put message back on front of queue and continue
							// "Insert" slice operation taken from:
							// https://code.google.com/p/go-wiki/wiki/SliceTricks
							user.mu.Lock()
							user.PendingMessages[username] = append(
								user.PendingMessages[username][:0], 
								append([]*SendMessageArgs{&args}, user.PendingMessages[username][0:]...)...)
							
							// append to our convo history too
							user.MessageHistory[username] = append(user.MessageHistory[username], &args) 
							
							//forward to K nearest neighbors
							kClosestEntryDists := user.Node.FindNearestNodes(Sha1(username))
							for _, entryDist := range kClosestEntryDists {
								var replyOther SendMessageReply
								call(entryDist.RoutingEntry.IpAddr, "User.SendMessageHandler", args, &replyOther)
							}
							user.mu.Unlock()
						}
					}
				}
			}
		}		
		Print(SendingTag, "Sender loop for %s running, pending: %v...", user.Name, user.PendingMessages)
		time.Sleep(500 * time.Millisecond)
	}
}

func (user *User) CheckStatus(ipAddr string) string {
	/*
		Returns status of IP Address endpoint. 
	*/
	status := Online
	routingEntry := RoutingEntry{IpAddr: ipAddr, NodeId: Sha1(ipAddr)}
	ok := user.Node.Ping(routingEntry)
	if !ok {
		status = Offline
	}
	Print(UserTag, "Checking status: %s is %s", ipAddr, status) 
	return status
}

func (user *User) IsOnline(username string) bool{
	ip := user.Node.FindUser(username)
	return ip != "" && user.CheckStatus(ip) == Online
}

func (user *User) AreNewMessagesFrom(other string, mostRecent time.Time) (bool, []*SendMessageArgs) {
	
	areNew := false
	newMessages := make([]*SendMessageArgs, 0)
	
	// get messages in this conversation, and traverse
	// messages in the conversation in order of timing
	messages := user.MessageHistory[other]
	sortutil.AscByField(messages, "Timestamp")
	for _, message := range messages {
		stamp := message.Timestamp
		
		// http://golang.org/src/pkg/time/time.go?s=2447:2479#L50
		if stamp.After(mostRecent) {
			areNew = true
		}
		
		if areNew {
			newMessages = append(newMessages, message)
		}
	}
	
	return areNew, newMessages
}

