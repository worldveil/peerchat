package dht
import "log"


type DhtNode struct {
	ipAddr string
	nodeId ID // sha1(ip)
	routingTable [IDLen][]RoutingEntry // map from nodeId to IP- a IDLen X K matrix
	// set routing table cap to bucket := make([]RoutingEntry, 0,K)
	kv map[string]string // map from username to IP
}

//this gets called when another node is contacting this node through any API method!
func (node *DhtNode) updateRoutingTable(nodeId ID, ipAddr string) {
	entry := RoutingEntry{nodeId: nodeId, ipAddr: ipAddr}
	n := find_n(nodeId, node.nodeId)
	bucket := &routingTable[n]
	//check if node is in routing table
	if len(bucket) < K {
		bucket[len(bucket)] = entry
	} else {
		for i := K -1; i >=0; i--{
			if ! node.Ping(bucket[i].ipAddr) {
				bucket[i] = entry
				break
			}
		}
	}
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
	reply.PingedNodeId = node.nodeId
	return nil
}

// Ping RPC API
//assume you already have them in routing table
func (node *DhtNode) Ping(ipAddr string) {
	args = &PingArgs{PingingNodeId: node.nodeId}
	var reply PingReply
	ok := call(ipAddr, "DhtNode.PingHandler", args, &reply)
	return ok
}


func MakeNode(myIpAddr string, routingTable [IDLen][]RoutingEntry) *DhtNode{
	node := &DhtNode{ipAddr: myIpAddr, nodeId: Sha1(myIpAddr), routingTable: routingTable}
	node.kv = make(map[string]string)
	return node
}