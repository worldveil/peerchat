package dht

import "log"
import "math"
import "github.com/pmylund/sortutil"

type DhtNode struct {
	IpAddr string
	NodeId ID // sha1(ip)
	RoutingTable [IDLen][]RoutingEntry // map from NodeId to IP- a IDLen X K matrix
	// set routing table cap to bucket := make([]RoutingEntry, 0,K)
	kv map[ID]string // map from username to IP
}

func moveToEnd(slice []RoutingEntry, index int) []RoutingEntry{
	return append(slice[:index], append(slice[index + 1:], slice[index])...)
}

//this gets called when another node is contacting this node through any API method!
func (node *DhtNode) updateRoutingTable(nodeId ID, ipAddr string) {
	// ordering of K bucket is from LRS to MRS
	entry := RoutingEntry{nodeId: nodeId, ipAddr: ipAddr}
	n := find_n(nodeId, node.NodeId) // n is the bucket index- index of first bit that doesn't match
	bucket := node.RoutingTable[n]
	defer func(){node.RoutingTable[n] = bucket}()
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

// get the alpha closest nodes to node ID in order to find user/node
// returns a slice of RoutingEntriesDist sorted in increasing order of dist from 
func (node *DhtNode) getClosest(target_result_len int, targetNodeId ID) []RoutingEntryDist{
	res := make([]RoutingEntryDist, 0, target_result_len)
	orig_bucket_idx := find_n(targetNodeId, node.NodeId)
	bucket_idx := orig_bucket_idx
	increasing_bucket := true
	for len(res) < target_result_len{ //need to keep looping over buckets until res is filled
		bucket := node.RoutingTable[bucket_idx]
		for _, value := range(bucket){
			xor := Xor(targetNodeId, value.nodeId)
			if len(res) < target_result_len {
				res = append(res, RoutingEntryDist{routingEntry: value, distance: xor})
			} else { //bucket is full	
				sortutil.AscByField(res, "distance")
				if xor < res[len(res) - 1].distance{
					res[len(res) - 1] = RoutingEntryDist{routingEntry: value, distance: xor}
				}
			}
		}
		if bucket_idx < IDLen - 1 && increasing_bucket{ // starts increasing
			bucket_idx++
		} else if bucket_idx == IDLen - 1 && increasing_bucket{ // stops increasing
			increasing_bucket = false
			bucket_idx = orig_bucket_idx - 1
		} else if bucket_idx == 0 && ! increasing_bucket{
			break
		}else {
			bucket_idx--
		}
	}
	sortutil.AscByField(res, "dist")
	return res
}

// StoreUser RPC handler
//this just stores the user in your kv
func (node *DhtNode) StoreUserHandler(args *StoreUserArgs, reply *StoreUserReply) error {	
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	node.kv[Sha1(args.AnnouncedUsername)] = args.AnnouncedIpAddr
	return nil
}

// called by makeNode
func (node *DhtNode) announceUser(username string, ipAddr string) {
	// tells the entire network: I'm a node and I'm online
	// loop through routing table, ping each one until one responds
	// does nodeLookup(node.NodeId) in order to populate other node's routing table with my info
	// does nodeLookup(hash(username)) to find K closest nodes to username then calls StoreUserHandler RPC on each node

}

// FindNode RPC handler
// all this does is call getClosest on K nodes
// returns k sorted slice of RoutingEntryDist from my routing table
func (node *DhtNode) FindNodeHandler(args *FindNodeArgs, reply *FindNodeReply) error {
	node.updateRoutingTable(args.QueryingNodeId, args.QueryingIpAddr)
	reply.TryNodes = node.getClosest(K, args.TargetNodeId)
	// reply.QueriedNodeId = node.nodeId //remove1
	return nil
}

// helper function called by both FindUser and AnnounceUser
// returns a k-length slice of RoutingEntriesDist sorted in increasing order of dist from 
func (node *DhtNode) nodeLookup(nodeId ID) []RoutingEntryDist{
	// get the closest nodes to the desired node ID
	// then add to a stack. we'll 
	closestNodes := node.getClosest(Alpha, nodeId)
	triedNodes := make(map[ID]bool)

	// send the initial min(Alpha, # of closest Node)
	// messages in flight to start the process
	replyChannel := make(chan *FindNodeReply, Alpha)
	sent := 0
	for _, entryDist := range closestNodes{
		go node.sendFindNodeQuery(entryDist.routingEntry, replyChannel)
		triedNodes[entryDist.routingEntry.nodeId] = true
		sent++
	}

	// now process replies as they arrive, spinning off new
	// requests up to alpha requests
	for {
		reply := <-replyChannel
		// process the reply, see if we are done
		// if we need to break because of stop cond: send done channel
		combined := append(closestNodes, reply.TryNodes...)
		sortutil.AscByField(combined, "distance")[: int(math.Min(float64(K), float64(len(combined))))]
		if isEqual(combined, closestNodes) { //closest Nodes have not changed
			return closestNodes
		}
		closestNodes = combined
		sent--

		// then check to see if we are still under
		// the total of alpha messages still in flight
		// and if so, send more
		for i := sent; i < Alpha; i++ {
			for idx, entryDist := range closestNodes{
				//find first value not in tried nodes
				_, already_tried := triedNodes[entryDist.routingEntry.nodeId]
				if ! already_tried{
					go node.sendFindNodeQuery(entryDist.routingEntry, replyChannel)
					sent++
					break
				}
			}			
		}		
	}	
}

func (node *DhtNode) sendFindNodeQuery(entry RoutingEntry, replyChannel chan *FindNodeReply) {
	/*
		This function is generally called as a separate goroutine. At the end of the call, 
		the reply is added to the done Channel, which is read by a separate thread. 
	*/
	ok := false
	args := &FindNodeArgs{QueryingNodeId: node.NodeId, QueryingIpAddr: "???", TargetNodeId: entry.nodeId}
	var reply FindNodeReply
	
	for !ok {
		ok = call(entry.ipAddr, "DhtNode.FindNodeHandler", args, &reply)
		if !ok {
			log.Printf("Failed! Will try again.")
		}
	}
	// add reference to reply onto the channel
	replyChannel <- &reply
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
	reply.QueriedNodeId = node.NodeId
	return nil
}

// Ping RPC API
//assume you already have them in routing table
func (node *DhtNode) Ping(ipAddr string) bool{
	args := &PingArgs{QueryingNodeId: node.NodeId}
	var reply PingReply
	ok := call(ipAddr, "DhtNode.PingHandler", args, &reply)
	return ok
}

//called when want to make a node from user.go
func MakeNode(username string, myIpAddr string, RoutingTable [IDLen][]RoutingEntry) *DhtNode{
	node := &DhtNode{IpAddr: myIpAddr, NodeId: Sha1(myIpAddr), RoutingTable: RoutingTable}
	node.kv = make(map[ID]string)
	node.announceUser(username, myIpAddr)
	return node
}