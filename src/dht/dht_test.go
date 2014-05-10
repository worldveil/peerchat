package dht

import "testing"
import "runtime"
import "fmt"
import "time"
import "math"
import "strconv"
import "math/rand"
// Signal failures with the following:
// t.Fatalf("error message here")

const localIp = "127.0.0.1"

func assertEqual(t *testing.T, out, ans interface{}) {
	if out != ans {
		t.Fatalf("wanted %v, got %v", ans, out)
	}
}

func checkLookup(t *testing.T, user1 *User, user2 *User) {
	resp := user1.node.FindUser(user2.name)
	assertEqual(t, resp, user2.node.IpAddr)
} 

func isEqualRE(entry1 []RoutingEntry, entry2 []RoutingEntry) bool{
	if len(entry1) != len(entry2){
		return false
	}
	for i, v := range entry1{
		if v != entry2[i] {
			return false
		}
	}
	return true
}

func registerMany(num_users int) map[string]*User{
	users := make(map[string]*User)

	bootstrap := ""

	for i :=0; i < num_users; i++{
		username := strconv.Itoa(i)
		ipAddr := localIp + ":" + strconv.Itoa(i + 7000)
		user := RegisterAndLogin(username, ipAddr, bootstrap)
		bootstrap = localIp + ":" + strconv.Itoa(i + 7000)
		time.Sleep(time.Millisecond * 5)
		users[username] = user
	}

	return users

}

/*
**  Unit tests for helper functions in common.go
*/
func TestCommonUnit(t *testing.T) {
	//common unit tests

	//Sha1 Test
	assertEqual(t, Sha1("abc"), Sha1("abc"))
	if Sha1("fjkels") == Sha1("qwewqi") {
		t.Fatalf("Sha1 collision")
	}
	//reference Sha1 computed at www.sha1-online.com and lowest 8 bytes converted to decimal
	assertEqual(t, Sha1("Forrest"), ID(10556789446649181072))
	assertEqual(t, Sha1("testing testing 123"), ID(16871972680281001427))

	//find_n
	a := ID(0)
	b := ID(1)
	c := ID(math.MaxUint64)
	d := ID(1 << 15)
	assertEqual(t, find_n(a, b), uint(63))
	assertEqual(t, find_n(a, c), uint(0))
	assertEqual(t, find_n(a, d), uint(48))
}

/*
**  Unit tests for helper functions in dhtnode.go
*/
func TestDhtNodeUnit(t *testing.T) {
	//DhtNode Unit Tests

	//moveToEnd Test
	id := Sha1("hi")
	in0 := []RoutingEntry{RoutingEntry{"a", id}, RoutingEntry{"b", id}, RoutingEntry{"c", id}}
	ans1 := []RoutingEntry{in0[1], in0[2], in0[0]}  //moveToEnd(in1, 0)
	ans2 := []RoutingEntry{in0[0], in0[2], in0[1]}  //moveToEnd(in1, 1)
	out1 := make([]RoutingEntry, 3)
	out2 := make([]RoutingEntry, 3)
	out3 := make([]RoutingEntry, 3)
	copy(out1, in0)
	copy(out2, in0)
	copy(out3, in0)
	moveToEnd(out1, 0)
	if !isEqualRE(ans1, out1) {
		t.Fatalf("wanted %v, got %v", ans1, out1)
	}
	moveToEnd(out2, 1)
	if !isEqualRE(ans2, out2) {
		t.Fatalf("wanted %v, got %v", ans2, out2)
	}
	moveToEnd(out3, 2)
	if !isEqualRE(in0, out3) {
		t.Fatalf("wanted %v, got %v, in0, out3")
	}
}

/*
**	TestBasic:
**	1) Starts two nodes
**	2) Introduces node1 to node2
**	3) Nodes send messages
**  4) Verify messages were recieved and saved intact.
**	
**	We verify the messages are not lost
**	and arrive unaltered. 
*/
func TestBasic(t *testing.T) {

	runtime.GOMAXPROCS(4)

	port1 := ":4444"
	port2 := ":5555"
	username1 := "Alice"
	username2 := "Frans"

	// user1 starts the Peerchat network, and
	// user2 joins by bootstrapping
	user1 := RegisterAndLogin(username1, localIp + port1, "")
	time.Sleep(time.Millisecond * 50)
	user2 := RegisterAndLogin(username2, localIp + port2, localIp + port1)

	time.Sleep(time.Millisecond * 50)

	// tests that we can find both users!
	u1_ip := user2.node.FindUser(username1)
	assertEqual(t, u1_ip, localIp+port1)
	u2_ip := user1.node.FindUser(username2)
	assertEqual(t, u2_ip, localIp+port2)
	
	// users exchange messages
	msg1 := "Hi Frans! Wanna play squash?"
	msg2 := "Sure Alice, what time?"
	
	user1.SendMessage(username2, msg1)
	time.Sleep(time.Second * 1)
	user2.SendMessage(username1, msg2)
	
	time.Sleep(1 * time.Second)
	
	// ensure the messages got there
	assertEqual(t, user2.MessageHistory[username1][0].Content, msg1)
	assertEqual(t, user1.MessageHistory[username2][0].Content, msg2)
	
	// kill user nodes
	user1.node.Dead <- true
	user2.node.Dead <- true
}

/*
**  RegisterAndLogin 30 users and make sure that each user can lookup
**  the IP address of every other user
*/
func TestManyRegistrations(t *testing.T) {
	
	users := registerMany(50)
	time.Sleep(time.Second)
	for _, user := range users{
		user.node.AnnounceUser(user.name, user.node.IpAddr)
	}
	time.Sleep(time.Second)
	for _, user := range users{
		fmt.Println(user.name, user.node.kv)
		for _, targetUser := range users{
			checkLookup(t, user, targetUser)
			//targetIp := user.node.FindUser(targetUsername)
			//assertEqual(t, targetIp, targetUser.node.IpAddr)
			//fmt.Println("Correct")
		}
	}
	
	for _, user := range users {
		user.node.Dead <- true
	}
}

/*
**  RegisterAndLogin 100 users. Choose 20 random pairs of users
**  and make sure they can look each other up
*/
func TestManyMoreRegistrations(t *testing.T) {
	size := 100
	users := registerMany(size)
	time.Sleep(time.Second)
	for _, user := range users {
		user.node.AnnounceUser(user.name, user.node.IpAddr)
	}
	time.Sleep(time.Second)
	for i:=0; i<20; i++ {
		idx :=  strconv.Itoa(rand.Int() % size)
		idx2 := strconv.Itoa(rand.Int() % size)
		fmt.Println("idx: ", idx, " idx2: ", idx2)
		checkLookup(t, users[idx], users[idx2])
		checkLookup(t, users[idx2], users[idx])
	}
}

/*
**  RegisterAndLogin 5 users. Make sure they can all send messages
**  to each other
*/
func TestSends(t *testing.T) {
	users := registerMany(5)
	fmt.Println("testing lookup...")
	time.Sleep(time.Second)
	users["0"].SendMessage("4", "hello 4")
	time.Sleep(5 * time.Second)
	// fmt.Printf("histoary = %v", users["4"].MessageHistory["0"])
	// assertEqual(t, users["4"].MessageHistory["0"][0].Content, "hello 4")

	//users["4"].SendMessage("0", "hi 0")
	//users["0"].SendMessage("9", "hello 9")

}

/*
**  RegisterAndLogin 10 users. Have 3 go offline. Make sure the remaining 
**  users can look each other up. Make sure the remaining users can
**  tell that logged-off users are not online
*/
func TestSomeFailures(t* testing.T) {
	//TODO: implement this test
}

/*
**  RegisterAndLogin 3 users. Have one log off and then log back on
**  with the same IP address. Make sure that user can still
**  lookup the other users
*/
func TestPersistance(t* testing.T) {
	//TODO: implement this test
}

/*
**  RegisterAndLogin 3 users. Have one log off and then log back on
**  with a new IP address. Make sure other users can lookup
**  that user's new address.
*/
func TestNewIP(t* testing.T) {
	//TODO: implement this test
}

/*
**  Register 10 users. Have them chat with each other for a bit.
**  Register 10 more users. Make sure 3 random pairs can look each
**  other up. Have 5 users from each group log off. Make sure 3 random
**  pairs can look each other up. Register 10 more users. Make sure 5
**  random pairs can look each other up. Have 5 more users log off and
**  5 others log back on with new IP addresses. Make sure 10 random 
**  pairs can look each other up
*/
func TestRealLife(t* testing.T) {
    //TODO: implement this test
}


