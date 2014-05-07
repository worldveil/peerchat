package dht

import "fmt"
import "time"
import "os"
import "encoding/gob"

type User struct {
	node *DhtNode
	name string
	// messageHistory map[string]string
}

func (node *DhtNode) Serialize(path string) {
	/*
		Serializes this DHT node into a file at 
		location path provided.
	*/
	// Create a file for IO
	encodeFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	// encode and write to file
	encoder := gob.NewEncoder(encodeFile)
	if err := encoder.Encode(node); err != nil {
		panic(err)
	}
	encodeFile.Close()
}

func Deserialize(path string) *DhtNode {
	/*
		Deserializes a DhtNode and loads
		it into a new DhtNode, which is 
		returned. 
	*/
	decodeFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer decodeFile.Close()

	// create decoder
	decoder := gob.NewDecoder(decodeFile)
	newNode := new(DhtNode)
	decoder.Decode(&newNode)
	return newNode
}

func LoadTable(username string) [IDLen][]RoutingEntry{
	table := [IDLen][]RoutingEntry{}
	// todo: load user specific routing from file or hard code or w/e
	return table
}

//SendMessage RPC Handler
func (user *User) SendMessageHandler(args *SendMessageArgs, reply *SendMessageReply) error{
	fmt.Println("My Message:", args.Content)
	return nil
}

//SendMessage API
func (user *User) SendMessage(username string, content string){
	ipAddr := user.node.GetUser(username)
	switch user.CheckStatus(ipAddr){
		case Online:
			args := &SendMessageArgs{Content: content, ToUsername: username, FromUsername: user.name, Timestamp: time.Now()}
			var reply SendMessageReply
			call(ipAddr, "User.SendMessageHandler", args, &reply)
		case Offline:
			// if not, queue?
	}
}

func (user *User) CheckStatus(ipAddr string) Status {
	return Online
}

func Login(username string, userIpAddr string) *User{
	routingTable := loadTable(username)
	node := MakeNode(userIpAddr, routingTable)
	user := &User{node: node, name: username}

	node.AnnounceUser(username, userIpAddr) //bootstrap to store in hash table

	return user
}

