package dht

// import "list"


type DhtNode struct {
	ipAddr string
	nodeId ID // sha1(ip)
	routingTable map[ID] string // map from nodeId to IP
	kv map[string]string // map from username to IP
}

// AnnouceUser RPC handlers
func (node *DhtNode) AnnouceUserHandler(args *AnnouceUserArgs, reply *AnnouceUserReply) error {
	return nil
}

// AnnouceUser API
func (node *DhtNode) AnnounceUser(username string, IpAddr string) {
}

// FindNode RPC handlers
func (node *DhtNode) FindNodeHandler(args *FindNodeArgs, reply *FindNodeReply) error {
	return nil
}

// FindNodeRPC API
func (node *DhtNode) FindNode(nodeId ID) {
}

// GetUser RPC handlers
func (node *DhtNode) GetUserHandler(args *GetUserArgs, reply *GetUserReply) error {
	return nil
}

// GetUser RPC API
// returns IP of username
func (node *DhtNode) GetUser(username string) string {
	return ""
}

// Ping RPC handlers
func (node *DhtNode) PingHandler(args *PingArgs, reply *PingReply) error {
	return nil
}

// Ping RPC API
func (node *DhtNode) Ping(nodeId ID) {
}

func MakeNode(myIpAddr string, routingTable map[ID]string) *DhtNode{
	node := &DhtNode{ipAddr: myIpAddr, nodeId: Sha1(myIpAddr), routingTable: routingTable}
	node.kv = make(map[string]string)
	return node
}