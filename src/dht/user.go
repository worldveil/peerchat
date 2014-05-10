package dht

import "time"
import "os"
import "encoding/gob"
import "sync"
import "net"
import "log"
import "fmt"

const UserTag = "USER"
const SendingTag = "SENDING"

type User struct {
	mu sync.Mutex
	node *DhtNode
	name string
	MessageHistory map[string][]*SendMessageArgs // username => messages we've gotten so far
	pendingMessages map[string][]*SendMessageArgs // username => slice of pending messages to apply
}

const PEERCHAT_USERDATA_DIR = "/tmp"

func UsernameToPath(username string) string {
	/*
		Given a username, returns the filepath
		where the backup will be located.
		
		NOTE: Go treats "/" as the path separator on
		all platforms.
	*/
	return PEERCHAT_USERDATA_DIR + "/" + username + ".gob"
}

func (u *User) GetMessagesFrom(username string) []*SendMessageArgs {
	/*
		Returns the list of SendMessageArgs
	*/
	if _, ok := u.MessageHistory[username]; ok {
		return u.MessageHistory[username]
	}
	return make([]*SendMessageArgs, 0)
}

func (u *User) Serialize() {
	/*
		Serializes this User struct.
	*/
	path := UsernameToPath(u.name)
	Print(UserTag, "Serializing path=%s for User %+v", path, u)
	encodeFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	// encode and write to file
	encoder := gob.NewEncoder(encodeFile)
	if err := encoder.Encode(u); err != nil {
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
		user.node.AnnounceUser(username, userIpAddr)
		go user.startSender()
	}
	return user
}

func RegisterAndLogin(username string, userIpAddr string, bootstrapIpAddr string) *User { 
	/*
		Attempts to register as a new user on the Peerchat network using 
		a known IP address as a boostrap. 
	*/
	Print(UserTag, "Bootstraping register with %s using username=%s, ip=%s, and ...", bootstrapIpAddr, username, userIpAddr) 
	user := makeUser(username, userIpAddr)
	user.setupUser()
	
	// check status of user we are about to bootstrap from
	status := user.CheckStatus(bootstrapIpAddr)
	if status == Offline {
		Print(UserTag, "Could not boostrap: %s was not online!", bootstrapIpAddr)
		//return some error
	}

	time.Sleep(10*time.Millisecond)
	user.node.AnnounceUser(username, userIpAddr)
	go user.startSender()
	return user
}

func (u *User) setupUser(){
	rpcs := u.node.SetupNode()
	rpcs.Register(u)

	// set up a connection listener
	l, e := net.Listen("tcp", u.node.IpAddr)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	
	// spin off go routine to listen for connections
	go func() {
		Print(StartTag, "Connection listener for %s starting...", u.node.IpAddr)
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("listen error: ", err);
			}
			
			// spin off goroutine to handle
			// RPC requests from other nodes
			go rpcs.ServeConn(conn)
		}
		
		Print(StartTag, "!!!!!!!!!!!!!!!!!! Server %s shutting down...", u.node.IpAddr)
		fmt.Println("here for no reason")
	}()
}

func makeUser(username string, ipAddr string) *User{
	Print(UserTag, "Creating a new User...")
	emptyPendingMessages := make(map[string][]*SendMessageArgs)
	history := make(map[string][]*SendMessageArgs)
	node := MakeNode(username, ipAddr)
	user := &User{node: node, name: username, pendingMessages: emptyPendingMessages, MessageHistory: history}
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
		if user.node.IpAddr != myIpAddr {
			
			// otherwise, create a new nodeId
			user.node.NodeId = Sha1(myIpAddr)
			
			Print(UserTag, "IP Address changed to %s, creating new NodeID=%s", myIpAddr, user.node.NodeId)
			
			// and rearrange the table based on new nodeId
			// first, get a list of all (nodeId, ipAddr) pairs
			routingEntries := make([]RoutingEntry, 0)
			
			// for each k-buckets row
			for _, row := range user.node.RoutingTable {
			
				// for each RoutingEntry in row
				for _, entry := range row {
					routingEntries = append(routingEntries, entry)
				}
			}
			
			// now delete old routing table and replace 
			// with a new (empty) one
			user.node.MakeEmptyRoutingTable()
			
			// then, for each pair, call:
			// updateRoutingTable(nodeId ID, IpAddr string)
			for _, entry := range routingEntries {
				if user.node.Ping(entry) {
					Print(UserTag, "RoutingEntry %+v is online, updating routing table...", entry)
					user.node.updateRoutingTable(entry)
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
	
	Print(UserTag, "%s recieved: %s, from %s at %v", user.name, args.Content, args.FromUsername, args.Timestamp)
	if _, ok := user.MessageHistory[args.FromUsername]; !ok {
	    user.pendingMessages[args.FromUsername] = make([]*SendMessageArgs, 0)
	}
	user.MessageHistory[args.FromUsername] = append(user.MessageHistory[args.FromUsername], args)
	
	return nil
}

//SendMessage API
func (user *User) SendMessage(username string, content string) {
	/*
		Sends message with content to username. In offline case, we save the message for later
	*/
	user.mu.Lock()
	Print(UserTag, "Queuing message \"%s\" to %s", content, username)
	if _, ok := user.pendingMessages[username]; !ok {
	    user.pendingMessages[username] = make([]*SendMessageArgs, 0)
	} 
	pendingMessage := &SendMessageArgs{Content: content, Timestamp: time.Now(), ToUsername: username, FromUsername: user.name}
	user.pendingMessages[username] = append(user.pendingMessages[username], pendingMessage)
	user.mu.Unlock()
}

func (user *User) startSender() {
	/*
		A separate thread which waits until Nodes 
		are up to send them messages
	*/
	Print(SendingTag, "Sender process for %s starting...", user.name)
	for {
		select {
			case <- user.node.Dead:
			break
			
			default:
			for username, _ := range user.pendingMessages {
				for len(user.pendingMessages[username]) > 0 {
					
					ip := user.node.FindUser(username)
					status := user.CheckStatus(ip)
					if status == Online {
						
						// pop first message args off of slice
						user.mu.Lock()
						var args SendMessageArgs
						if len(user.pendingMessages[username]) == 1 { 
							args, user.pendingMessages[username] = *user.pendingMessages[username][0], make([]*SendMessageArgs, 0)
						} else if len(user.pendingMessages[username]) > 1 {
							// "Pop(0)" slice operation taken from:
							// https://code.google.com/p/go-wiki/wiki/SliceTricks
							args, user.pendingMessages[username] = *user.pendingMessages[username][0], user.pendingMessages[username][1:]
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
							user.pendingMessages[username] = append(
								user.pendingMessages[username][:0], 
								append([]*SendMessageArgs{&args}, user.pendingMessages[username][0:]...)...)
							user.mu.Unlock()
						}
					}
				}
			}
		}
		
		Print(SendingTag, "Sender loop for %s running, pending: %v...", user.name, user.pendingMessages)
		time.Sleep(500 * time.Millisecond)
	}
}

func (user *User) CheckStatus(ipAddr string) string {
	/*
		Returns status of IP Address endpoint. 
	*/
	status := Online
	routingEntry := RoutingEntry{IpAddr: ipAddr, NodeId: Sha1(ipAddr)}
	ok := user.node.Ping(routingEntry)
	if !ok {
		status = Offline
	}
	Print(UserTag, "Checking status: %s is %s", ipAddr, status) 
	return status
}

