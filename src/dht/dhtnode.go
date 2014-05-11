package dht

import "math"
import "github.com/pmylund/sortutil"
//import "strings"
import "net/rpc"
import "encoding/gob"

const ApiTag = "API"
const DHTHelperTag = "HELPER"
const HandlerTag = "HANDLER"
const StartTag = "START"
const Temp = "Temp"

type DhtNode struct {
	IpAddr string
	NodeId ID // sha1(ip)
	RoutingTable [IDLen][]RoutingEntry // map from NodeId to IP- a IDLen X K matrix
	kv map[ID]string // map from username to IP
	port string
	Dead chan bool
}

//this gets called when another node is contacting this node through any API method!
func (node *DhtNode) updateRoutingTable(entry RoutingEntry) {
	Print(DHTHelperTag, "Node %v calling updateRoutingTable for node: %v, ip: %s", Short(node.NodeId), Short(entry.NodeId), entry.IpAddr)
	// ordering of K bucket is from LRS to MRS
	n := find_n(entry.NodeId, node.NodeId) // n is the bucket index- index of first bit that doesn't match
	bucket := node.RoutingTable[n]
	//check if node is in routing table
	for idx, r_entry := range(bucket){
		if r_entry == entry{
			bucket = moveToEnd(bucket, idx)
			node.RoutingTable[n] = bucket
			Print(DHTHelperTag, "Node %v done updateRoutingTable Routing table is: %v", Short(node.NodeId), node.RoutingTable)
			return
		}
	} // new entry is not in bucket
	if len(bucket) < K { // bucket is not full
		bucket = append(bucket, entry)
	} else { // bucket is full
		// ping the front of list (LRS)
		if ! node.Ping(bucket[0]){
			bucket[0] = entry //if does not respond, replace
		}
		bucket = moveToEnd(bucket, 0)	// move to end
	}
	node.RoutingTable[n] = bucket
	Print(DHTHelperTag, "Node %v done updateRoutingTable Routing table is: %v", Short(node.NodeId), node.RoutingTable)
}

// get the alpha closest nodes to node ID in order to find user/node
// returns a slice of RoutingEntriesDist sorted in increasing order of dist from 
func (node *DhtNode) getClosest(target_result_len int, targetNodeId ID) []RoutingEntryDist{
	Print(DHTHelperTag, "Node %v calling getClosest to get %d closest to %v", Short(node.NodeId), target_result_len, Short(targetNodeId))
	empty := true
	for _, bucket := range node.RoutingTable {
		if len(bucket) > 0{
			empty = false
		}
	}
	if empty {
		Print(DHTHelperTag, "Warning: Node %v has empty routing table!", Short(node.NodeId))
	}
	res := make([]RoutingEntryDist, 0, target_result_len)
	orig_bucket_idx := find_n(targetNodeId, node.NodeId)
	bucket_idx := orig_bucket_idx
	increasing_bucket := true
	for len(res) < target_result_len{ //need to keep looping over buckets until res is filled
		bucket := node.RoutingTable[bucket_idx]

		for _, value := range(bucket){
			xor := Xor(targetNodeId, value.NodeId)
			if len(res) < target_result_len {
				res = append(res, RoutingEntryDist{RoutingEntry: value, Distance: xor})
			} else { //bucket is full	
				sortutil.AscByField(res, "Distance")
				if xor < res[len(res) - 1].Distance{
					res[len(res) - 1] = RoutingEntryDist{RoutingEntry: value, Distance: xor}
				}
			}
		}
		if bucket_idx < IDLen - 1 && increasing_bucket{ // starts increasing
			bucket_idx++
		} else if bucket_idx == IDLen - 1 && increasing_bucket{ // stops increasing
			increasing_bucket = false
			if orig_bucket_idx == 0 {
				break
			}
			bucket_idx = orig_bucket_idx - 1
		} else if bucket_idx == 0 && ! increasing_bucket{
			break
		} else {
			bucket_idx--
		}
	}
	// fmt.Println(res)
	sortutil.AscByField(res, "Distance")
	return res
}

// StoreUser RPC handler
//this just stores the user in your kv
func (node *DhtNode) StoreUserHandler(args *StoreUserArgs, reply *StoreUserReply) error {
	Print(HandlerTag, "Node %v StoreUserHandler called by %v. kv[%v]=%v", Short(node.NodeId), Short(args.QueryingNodeId), Short(args.AnnouncedUserId), args.QueryingIpAddr)
	node.updateRoutingTable(RoutingEntry{NodeId: args.QueryingNodeId, IpAddr: args.QueryingIpAddr})
	node.kv[args.AnnouncedUserId] = args.AnnouncedIpAddr
	return nil
}

// called by User
// tells the entire network: I'm a node and I'm online
func (node *DhtNode) AnnounceUser(username string, ipAddr string) {
	//put myself in routing table
	node.kv[Sha1(username)] = ipAddr

	Print(ApiTag, "Node %v calling AnnounceUser, username: %v, ipAddr: %v", Short(node.NodeId), username, ipAddr)
	// does lookup(node.NodeId) in order to populate other node's routing table with my info
	node.idLookup(node.NodeId, "Node")
	// does lookup(hash(username)) to find K closest nodes to username then calls StoreUserHandler RPC on each node
	kClosestEntryDists, _ := node.idLookup(Sha1(username), "Node")
	args := &StoreUserArgs{QueryingNodeId: node.NodeId, QueryingIpAddr: ipAddr, AnnouncedUserId: Sha1(username), AnnouncedIpAddr: ipAddr}
	for _, entryDist := range kClosestEntryDists{
		var reply PingReply
		call(entryDist.RoutingEntry.IpAddr, "DhtNode.StoreUserHandler", args, &reply)
	}
}

// FindNode RPC handler
// all this does is call getClosest on K nodes
// returns k sorted slice of RoutingEntryDist from my routing table
func (node *DhtNode) FindNodeHandler(args *FindIdArgs, reply *FindIdReply) error {
	Print(HandlerTag, "Node %v FindNodeHandler called by %v, TargetId: %v", Short(node.NodeId), Short(args.QueryingNodeId), Short(args.TargetId))
	reply.QueriedNodeId = node.NodeId
	reply.QueriedIpAddr = node.IpAddr
	node.updateRoutingTable(RoutingEntry{NodeId: args.QueryingNodeId, IpAddr: args.QueryingIpAddr})
	reply.TryNodes = node.getClosest(K, args.TargetId)
	return nil
}

// FindUser RPC handlers
//checks if user is in, if not, return false
func (node *DhtNode) FindUserHandler(args *FindIdArgs, reply *FindIdReply) error {
	Print(HandlerTag, "Node %v FindUserHandler called by %v, TargetId: %v. My kv is %v", Short(node.NodeId), Short(args.QueryingNodeId), Short(args.TargetId), node.kv)
	reply.QueriedNodeId = node.NodeId
	reply.QueriedIpAddr = node.IpAddr
	node.updateRoutingTable(RoutingEntry{NodeId: args.QueryingNodeId, IpAddr: args.QueryingIpAddr})
	ipAddr, exists := node.kv[args.TargetId]
	if exists {
		reply.TargetIpAddr = ipAddr
		Print(HandlerTag, "Node %v FindUserHandler (finished) called by %v, TargetId: %v. Target user is in my map! returning user", Short(node.NodeId), Short(args.QueryingNodeId), args.TargetId)
	} else{
		reply.TryNodes = node.getClosest(K, args.TargetId)
		Print(HandlerTag, "Node %v FindUserHandler (finished) called by %v, TargetId: %v. Target user is NOT in my map! returning closest nodes", Short(node.NodeId), Short(args.QueryingNodeId), args.TargetId)
	}	
	return nil
}

// helper function called by both FindUser and AnnounceUser
// returns a k-length slice of RoutingEntriesDist sorted in increasing order of dist from 
func (node *DhtNode) idLookup(targetId ID, targetType string) ([]RoutingEntryDist, string) {
	Print(DHTHelperTag, "Node %v calling idLookup, targetId: %v, targetType: %v", Short(node.NodeId), Short(targetId), targetType)
	// get the closest nodes to the desired node ID
	// then add to a stack. we'll 
	closestNodes := node.getClosest(Alpha, targetId)
	if len(closestNodes) == 0 {
		Print(ApiTag, "Node %v found 0 closest nodes- empty routing table!", Short(node.NodeId))
		return []RoutingEntryDist{}, ""
	}
	triedNodes := make(map[ID]bool)
	triedNodes[node.NodeId] = true
	closestNodes= append(closestNodes, RoutingEntryDist{Distance: Xor(node.NodeId, targetId), RoutingEntry: RoutingEntry{NodeId: node.NodeId, IpAddr: node.IpAddr}})
	sortutil.AscByField(closestNodes, "Distance")
	closestNodes = removeDuplicates(closestNodes)
	// send the initial min(Alpha, # of closest Node)
	// messages in flight to start the process
	replyChannel := make(chan *FindIdReply, Alpha)
	sent := 0
	for _, entryDist := range closestNodes{
		_, already_tried := triedNodes[entryDist.RoutingEntry.NodeId]
		if ! already_tried {
			go node.sendFindIdQuery(entryDist.RoutingEntry, replyChannel, targetId, targetType)
			triedNodes[entryDist.RoutingEntry.NodeId] = true
			sent++
		}
	}

	// now process replies as they arrive, spinning off new
	// requests up to alpha requests
	for {
		reply := <-replyChannel
		if reply == nil{
			Print(DHTHelperTag, "Node %v received dropped DhtNode.Find%sHandler packet", Short(node.NodeId), targetType)
			continue
		}
		Print(DHTHelperTag, "Node %v received Find%v response from %v. Response is %v, %v", Short(node.NodeId), targetType, Short(reply.QueriedNodeId), reply.TryNodes, reply.TargetIpAddr)
		//update our routing table with queriedNodeId
		node.updateRoutingTable(RoutingEntry{NodeId: reply.QueriedNodeId, IpAddr: reply.QueriedIpAddr})

		//if we are looking for a user's ip address break early if found
		if targetType == "User" && reply.TargetIpAddr != "" {
			//send user to closest node that did not return value
			for _, entryDist := range closestNodes{
				if triedNodes[entryDist.RoutingEntry.NodeId] && entryDist.RoutingEntry.NodeId != reply.QueriedNodeId{
					args := &StoreUserArgs{QueryingNodeId: node.NodeId, QueryingIpAddr: node.IpAddr, AnnouncedUserId: targetId, AnnouncedIpAddr: reply.TargetIpAddr}
					var reply2 StoreUserReply
					call(entryDist.RoutingEntry.IpAddr, "DhtNode.StoreUserHandler", args, &reply2)
					break //only cache once!
				}
			}
			return []RoutingEntryDist{} , reply.TargetIpAddr
		}
		// process the reply, see if we are done
		// if we need to break because of stop cond: send done channel
		combined := append(closestNodes, reply.TryNodes...)
		combined = removeDuplicates(combined)
		sortutil.AscByField(combined, "Distance")
		combined = combined[: int(math.Min(float64(K), float64(len(combined))))]
		
		sortutil.AscByField(combined, "Distance")
		done := true
		for _, entryDist := range combined {
			_, already_tried := triedNodes[entryDist.RoutingEntry.NodeId]
			if !already_tried {
				done = false
			}
		}

		if isEqual(combined, closestNodes) && done { //closest Nodes have not changed
			Print(DHTHelperTag, "Node %v is exiting ID lookup because it's closest nodes have not changed! %v", Short(node.NodeId), closestNodes)
			return closestNodes, ""
		}
		closestNodes = combined
		sent--
		// then check to see if we are still under
		// the total of alpha messages still in flight
		// and if so, send more
		for i := sent; i < Alpha; i++ {
			for _, entryDist := range closestNodes{
				//find first value not in tried nodes
				_, already_tried := triedNodes[entryDist.RoutingEntry.NodeId]
				if ! already_tried {
					go node.sendFindIdQuery(entryDist.RoutingEntry, replyChannel, targetId, targetType)
					triedNodes[entryDist.RoutingEntry.NodeId] = true
					sent++
					break
				}
			}			
		}		
	}	
}

func (node *DhtNode) sendFindIdQuery(entry RoutingEntry, replyChannel chan *FindIdReply, targetId ID, targetType string) {
	/*
		This function is generally called as a separate goroutine. At the end of the call, 
		the reply is added to the done Channel, which is read by a separate thread. 
	*/
	Print(DHTHelperTag, "Node %v sending find%vQuery to node %v. Looking for ID %v", Short(node.NodeId), targetType, Short(entry.NodeId), Short(targetId))
	
	args := &FindIdArgs{QueryingNodeId: node.NodeId, QueryingIpAddr: node.IpAddr, TargetId: targetId}
	var reply FindIdReply
	call(entry.IpAddr, "DhtNode.Find" + targetType + "Handler", args, &reply) //if failed, reply will be empty!
	
	// add reference to reply onto the channel
	replyChannel <- &reply
}

// FindUser RPC API
// returns IP of username or "" if can't find IP of username
func (node *DhtNode) FindUser(username string) string {
	Print(ApiTag, "Node %v calling FindUser", Short(node.NodeId))
	targetId := Sha1(username)
	//check if have locally
	ipAddr, exists := node.kv[targetId]
	if exists {
		return ipAddr
	} 
	//do a idLookup to get K closest to username- query them in order of ascending distance until finds username
	_, ipAddr = node.idLookup(targetId, "User")	
	return ipAddr
}

// Ping RPC handlers
func (node *DhtNode) PingHandler(args *PingArgs, reply *PingReply) error {
	Print(HandlerTag, "Node %v PingHandler called by %v", Short(node.NodeId), Short(args.QueryingNodeId))
	reply.QueriedNodeId = node.NodeId
	return nil
}

// Ping RPC API
//assume you already have them in routing table
func (node *DhtNode) Ping(routingEntry RoutingEntry) bool{
	Print(ApiTag, "Node %v calling Ping on ip: %s", Short(node.NodeId), routingEntry.IpAddr)
	args := &PingArgs{QueryingNodeId: node.NodeId, QueryingIpAddr: node.IpAddr}
	var reply PingReply
	ok := call(routingEntry.IpAddr, "DhtNode.PingHandler", args, &reply)
	if ok && (reply.QueriedNodeId == routingEntry.NodeId){
		node.updateRoutingTable(routingEntry) // this is needed for bootstrap
		return true
	}
	return false
}

func (node *DhtNode) MakeEmptyRoutingTable() {
	/*
		Creates an empty routing table
	*/
	var routingTable [IDLen][]RoutingEntry
	for i, _ := range routingTable {
		routingTable[i] = make([]RoutingEntry, 0)
	}
	node.RoutingTable = routingTable
}

func (node *DhtNode) SetupNode() *rpc.Server{
	// register which objects RPC can serialize/deserialize
	gob.Register(SendMessageArgs{})
	gob.Register(SendMessageReply{})
	gob.Register(StoreUserArgs{})
	gob.Register(StoreUserReply{})
	gob.Register(FindIdArgs{})
	gob.Register(FindIdReply{})
	gob.Register(PingArgs{})
	gob.Register(PingReply{})
	gob.Register(RoutingEntryDist{})

	// register the exported methods and
	// create an RPC server
	rpcs := rpc.NewServer()
	rpcs.Register(node)

	return rpcs

}

//called when want to make a node from user.go
func MakeNode(username string, myIpAddr string) *DhtNode{
	/*
		Creates a DHTNode with a given username, ip address, and routing table. 
	*/
	node := &DhtNode{IpAddr: myIpAddr, NodeId: Sha1(myIpAddr)}
	node.kv = make(map[ID]string)
	node.MakeEmptyRoutingTable()
	//node.port = strings.Split(myIpAddr, ":")[1]
	node.Dead = make(chan bool, 10)
	
	Print(StartTag, "DHT Node created for username=%s with ip=%s, moving to gob setup...", username, myIpAddr)	
	return node
}
