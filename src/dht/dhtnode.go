package dht
import "fmt"

type DhtNode struct {
	ipAddr string
	nodeId ID // sha1(ip)
	routingTable [IDLen][]RoutingEntry // map from nodeId to IP- a IDLen X K matrix
	// set routing table cap to bucket := make([]RoutingEntry, 0,K)
	kv map[string]string // map from username to IP
}

func moveToEnd(slice []RoutingEntry, index int) []RoutingEntry{
	return append(slice[:index], append(slice[index + 1:], slice[index])...)
}

//this gets called when another node is contacting this node through any API method!
func (node *DhtNode) updateRoutingTable(nodeId ID, ipAddr string) {
	entry := RoutingEntry{nodeId: nodeId, ipAddr: ipAddr}
	n := find_n(nodeId, node.nodeId) // n is the bucket index- index of first bit that doesn't match
	bucket := node.routingTable[n]
	defer func(){node.routingTable[n] = bucket}()
	//check if node is in routing table
	for idx, r_entry := range(bucket){
		if r_entry == entry{
			bucket = moveToEnd(bucket, idx)
			return
		}
	} // new entry is not in bucket
	if len(bucket) < K { // bucket is not full
		bucket = append(bucket, entry)
	} else { // bucket is full
		// ping the front of list (LRS)
		if ! node.Ping(bucket[0].ipAddr){
			bucket[0] = entry //if does not respond, replace
		}
		bucket = moveToEnd(bucket, 0)	// move to end
	}
}

// AnnouceUser RPC handlers
func (node *DhtNode) AnnouceUserHandler(args *AnnouceUserArgs, reply *AnnouceUserReply) error {
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	return nil
}

// AnnouceUser API
func (node *DhtNode) AnnounceUser(username string, IpAddr string) {
}

// FindNode RPC handlers
func (node *DhtNode) FindNodeHandler(args *FindNodeArgs, reply *FindNodeReply) error {
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	return nil
}

func (node *DhtNode) getAlphaClosest(nodeId ID) []RoutingEntry{
	res := make([]RoutingEntry, 0, Alpha)
	orig_n := find_n(nodeId, node.nodeId)
	n := orig_n
	increasing_n := true
	for len(res) < Alpha{ //need to keep looping over buckets until res is filled
		bucket := node.routingTable[n]
		for _, value := range(bucket){
			if len(res) < Alpha {
				res = append(res, value)
			}else{ //bucket is full
				xor := Xor(nodeId, value.nodeId)
				need_to_replace := false
				var max_dist ID
				max_dist = 0
				max_idx := 0
				for idx, res_value := range(res) {
					res_val_distance := Xor(nodeId, res_value.nodeId)
					if xor < res_val_distance{ // current value is closer than what's in res
						need_to_replace = true
					}
					if max_dist < res_val_distance{
						max_dist = res_val_distance
						max_idx = idx
					}
				}
				if need_to_replace{
					res[max_idx] = value
				}
			}
		}
		if n < IDLen - 1 && increasing_n{ // starts increasing
			n++
		} else if n == IDLen - 1 && increasing_n{ // stops increasing
			increasing_n = false
			n = orig_n - 1
		} else if n == 0 && ! increasing_n{
			break
		}else {
			n--
		}
	}
	return res
}

// FindNodeRPC API
func (node *DhtNode) FindNode(nodeId ID) {
	closest_nodes := node.getAlphaClosest(nodeId)
	fmt.Println(closest_nodes)
}

// GetUser RPC handlers
func (node *DhtNode) GetUserHandler(args *GetUserArgs, reply *GetUserReply) error {
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	return nil
}

// GetUser RPC API
// returns IP of username
func (node *DhtNode) GetUser(username string) string {

	return ""
}

// Ping RPC handlers
func (node *DhtNode) PingHandler(args *PingArgs, reply *PingReply) error {
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	reply.QueriedNodeId = node.nodeId
	return nil
}

// Ping RPC API
//assume you already have them in routing table
func (node *DhtNode) Ping(ipAddr string) bool{
	args := &PingArgs{QueryingNodeId: node.nodeId}
	var reply PingReply
	ok := call(ipAddr, "DhtNode.PingHandler", args, &reply)
	return ok
}


func MakeNode(myIpAddr string, routingTable [IDLen][]RoutingEntry) *DhtNode{
	node := &DhtNode{ipAddr: myIpAddr, nodeId: Sha1(myIpAddr), routingTable: routingTable}
	node.kv = make(map[string]string)
	return node
}