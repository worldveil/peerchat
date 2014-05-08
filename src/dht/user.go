package dht

import "time"
import "os"
import "encoding/gob"

const UserTag = "USER"

type User struct {
	node *DhtNode
	name string
	// messageHistory map[string]string
	
	pendingMessages map[string][]string // ipAddr => slice of messages to send
}

const PEERCHAT_USERDATA_DIR = "/tmp/"

func usernameToPath(username string) string {
	/*
		Given a username, returns the filepath
		where the backup will be located.
		
		NOTE: Go treats "/" as the path separator on
		all platforms.
	*/
	return PEERCHAT_USERDATA_DIR + "/" + username + ".gob"
}

func (u *User) Serialize() {
	/*
		Serializes this User struct.
	*/
	path := usernameToPath(u.name)
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
	path := usernameToPath(username)
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
		Print(UserTag, "Could not load user from file, creating a new User...")
		emptyPendingMessages := make(map[string][]string)
		node := MakeNode(username, myIpAddr)
		user = &User{node: node, name: username, pendingMessages: emptyPendingMessages}		
	}
	user.node.AnnounceUser(username, myIpAddr)
	
	return user
}

//SendMessage RPC Handler
func (user *User) SendMessageHandler(args *SendMessageArgs, reply *SendMessageReply) error{
	Print(UserTag, "My Message: %s", args.Content)
	return nil
}

//SendMessage API
func (user *User) SendMessage(username string, content string){
	ipAddr := user.node.FindUser(username)
	switch user.CheckStatus(ipAddr) {
		
		case Online:
			args := &SendMessageArgs{Content: content, ToUsername: username, FromUsername: user.name, Timestamp: time.Now()}
			var reply SendMessageReply
			call(ipAddr, "User.SendMessageHandler", args, &reply)
			Print(UserTag, "Sending \"%s\" to %s at %s...", content, username, ipAddr)
			
		case Offline:
			// if not, queue?
			Print(UserTag, "Recipient user=%s, ipAddr=%s was not online, queuing \"%s\" for later...", username, ipAddr, content)
	}
}

func (user *User) CheckStatus(ipAddr string) Status {
	/*
		Returns status of IP Address endpoint. 
	*/
	
	Print(UserTag, "Checking status: %s is Online (##HARDCODED##)", ipAddr) 
	return Online
}

func Login(username string, userIpAddr string) *User {
	/*
		Attempts to log into the Peerchat network by loading a previous configuration
		and defaulting to creating a new one. 
	*/
	
	Print(UserTag, "Attempting to log on with username=%s and ip=%s...", username, userIpAddr) 
	user := loadUser(username, userIpAddr)
	return user
}

func Register(username string, userIpAddr string, bootstrapIpAddr string) *User { 
	/*
		Attemps to register as a new user on the Peerchat network using 
		a known IP address as a boostrap. 
	*/
	
	Print(UserTag, "Bootstraping register with %s using username=%s, ip=%s, and ...", bootstrapIpAddr, username, userIpAddr) 
	user := Login(username, userIpAddr)
	routingEntry := RoutingEntry{ipAddr: bootstrapIpAddr, nodeId: Sha1(bootstrapIpAddr)}
	ok := user.node.Ping(routingEntry)
	if !ok {
		Print(UserTag, "Could not boostrap: %s was not online!", bootstrapIpAddr) 
		//return some error
	}
	user.node.AnnounceUser(username, userIpAddr)
	return user
}

