package dht

import "fmt"
import "time"


type User struct {
	node *DhtNode
	name string
	// messageHistory map[string]string
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

func loadTable(username string) [IDLen][]RoutingEntry{
	table := [IDLen][]RoutingEntry{}
	// todo: load user specific routing from file or hard code or w/e
	return table
}

func Login(username string, userIpAddr string) *User{
	routingTable := loadTable(username)
	node := MakeNode(userIpAddr, routingTable)
	user := &User{node: node, name: username}

	node.AnnounceUser(username, userIpAddr) //bootstrap to store in hash table

	return user
}

