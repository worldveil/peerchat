package dht

import "testing"
import "runtime"
import "fmt"
import "time"
import "math"
import "strconv"
import "math/rand"
import "os"

// Signal failures with the following:
// t.Fatalf("error message here")

const localIp = "127.0.0.1"


func assertEqual(t *testing.T, out, ans interface{}) {
	if out != ans {
		t.Fatalf("wanted %v, got %v", ans, out)
	}
}

func checkLookup(t *testing.T, user1 *User, user2 *User) {
	resp := user1.Node.FindUser(user2.Name)
	assertEqual(t, resp, user2.Node.IpAddr)
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

func registerMany(num_users int) []*User{
	users := make([]*User, num_users)

	bootstrap := ""

	for i :=0; i < num_users; i++{
		username := strconv.Itoa(i)
		ipAddr := localIp + ":" + strconv.Itoa(i + 8000)
		user := RegisterAndLogin(username, ipAddr, bootstrap)
		bootstrap = localIp + ":" + strconv.Itoa(i + 8000)
		time.Sleep(time.Millisecond * 5)
		users[i] = user
	}

	time.Sleep(time.Second)

	for _, user := range(users) {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
	time.Sleep(time.Second)
	return users

}

func killAll(users []*User){
	for _, user := range users {
		user.Logoff()
	}
	time.Sleep(time.Millisecond * 400)
}

func TestGobbing(t *testing.T) {
	
	port1 := ":4444"
	port2 := ":5555"
	username1 := "Alice"
	username2 := "Frans"
	
	// remove any serialized users
	os.Remove(UsernameToPath(username1))
	os.Remove(UsernameToPath(username2))

	// user1 starts the Peerchat network, and
	// user2 joins by bootstrapping
	user1 := RegisterAndLogin(username1, localIp + port1, "")
	time.Sleep(time.Millisecond * 50)
	user2 := RegisterAndLogin(username2, localIp + port2, localIp + port1)

	time.Sleep(time.Millisecond * 50)

	// tests that we can find both users!
	u1_ip := user2.Node.FindUser(username1)
	assertEqual(t, u1_ip, localIp+port1)
	u2_ip := user1.Node.FindUser(username2)
	assertEqual(t, u2_ip, localIp+port2)
	
	// users exchange messages
	msg1 := "Hi Frans! Wanna play squash?"
	msg2 := "Sure Alice, what time?"
	
	user1.SendMessage(username2, msg1)
	time.Sleep(time.Second * 1)
	user2.SendMessage(username1, msg2)
	
	time.Sleep(1 * time.Second)
	
	// remove any serialized users
	os.Remove(UsernameToPath(username1))
	os.Remove(UsernameToPath(username2))
	
	user1.Serialize()
	user2.Serialize()
	
	success, one := Deserialize(user1.Name)
	success2, two := Deserialize(user2.Name)
	if success && success2 {
		assertEqual(t, one.Name, user1.Name)
		assertEqual(t, two.Name, user2.Name)
		assertEqual(t, len(one.MessageHistory[user2.Name]), len(user1.MessageHistory[user2.Name]))
		assertEqual(t, len(two.MessageHistory[user1.Name]), len(user2.MessageHistory[user1.Name]))
		assertEqual(t, len(one.PendingMessages[user2.Name]), len(user1.PendingMessages[user2.Name]))
		assertEqual(t, len(two.PendingMessages[user1.Name]), len(user2.PendingMessages[user1.Name]))
		assertEqual(t, len(one.ReceivedMessageIdentifiers), len(user1.ReceivedMessageIdentifiers))
		assertEqual(t, len(two.ReceivedMessageIdentifiers), len(user2.ReceivedMessageIdentifiers))
	}
}

/*
**  Unit tests for helper functions in common.go
*/
func TestCommonUnit(t *testing.T) {
	//common unit tests
	runtime.GOMAXPROCS(4)

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
**  4) Verify messages were recieved and 
**     saved intact with no duplication.
**	
**	We verify the messages are not lost
**	and arrive unaltered. 
*/
func TestBasic(t *testing.T) {
	fmt.Println("Running TestBasic")
	

	port1 := ":4444"
	port2 := ":5555"
	username1 := "Alice"
	username2 := "Frans"
	
	// remove any serialized users
	os.Remove(UsernameToPath(username1))
	os.Remove(UsernameToPath(username2))

	// user1 starts the Peerchat network, and
	// user2 joins by bootstrapping
	user1 := RegisterAndLogin(username1, localIp + port1, "")
	time.Sleep(time.Millisecond * 50)
	user2 := RegisterAndLogin(username2, localIp + port2, localIp + port1)

	time.Sleep(time.Millisecond * 50)

	// tests that we can find both users!
	u1_ip := user2.Node.FindUser(username1)
	assertEqual(t, u1_ip, localIp+port1)
	u2_ip := user1.Node.FindUser(username2)
	assertEqual(t, u2_ip, localIp+port2)
	
	// users exchange messages
	msg1 := "Hi Frans! Wanna play squash?"
	msg2 := "Sure Alice, what time?"
	
	user1.SendMessage(username2, msg1)
	time.Sleep(time.Second * 1)
	user2.SendMessage(username1, msg2)
	
	time.Sleep(1 * time.Second)
	
	// ensure the messages got there
	assertEqual(t, user2.GetMessagesFrom(user1)[0].Content, msg1)
	assertEqual(t, user1.GetMessagesFrom(user2)[0].Content, msg2)
	
	assertEqual(t, len(user2.MessageHistory[username1]), 1)
	assertEqual(t, len(user1.MessageHistory[username2]), 1)
	
	// kill user nodes
	user1.Logoff()
	user2.Logoff()
}

/*
**  RegisterAndLogin 50 users and make sure that each user can lookup
**  the IP address of every other user
*/
func TestManyRegistrations(t *testing.T) {
	fmt.Println("Running TestManyRegistrations")	
	users := registerMany(10)
	defer killAll(users)
	for _, user := range users{
		for _, targetUser := range users{
			checkLookup(t, user, targetUser)
		}
	}
	fmt.Println("Passed!")
}

/*
**  RegisterAndLogin 100 users. Choose 20 random pairs of users
**  and make sure they can look each other up
*/
func TestManyMoreRegistrations(t *testing.T) {
	fmt.Println("Running TestManyMoreRegistrations")
	size := 20
	users := registerMany(size)
	defer killAll(users)
	for i:=0; i<20; i++ {
		idx :=  rand.Int() % size
		idx2 := rand.Int() % size
		fmt.Println("idx: ", idx, " idx2: ", idx2)
		checkLookup(t, users[idx], users[idx2])
		checkLookup(t, users[idx2], users[idx])
	}
	fmt.Println("Passed!")
}
func sendAndCheck(t *testing.T, sender *User, receiver *User) {
	msg := "message " + strconv.Itoa(rand.Int() % 1000)
	idx := len(receiver.MessageHistory[sender.Name])
	sender.SendMessage(receiver.Name, msg)
	time.Sleep(time.Second)
	assertEqual(t, receiver.MessageHistory[sender.Name][idx].Content, msg)
}

/*
**  RegisterAndLogin 5 users. Make sure they can all send messages
**  to each other
*/
func TestSends(t *testing.T) {
	size := 5
	users := registerMany(5)
	defer killAll(users)
	for i:=0; i < size; i++ {
		for j:=0; j<size; j++ {
			go sendAndCheck(t, users[i], users[j])
		}
	}
	fmt.Println("Passed!")
}

/*
**  RegisterAndLogin 10 users. Have 3 go offline. Make sure the remaining 
**  users can look each other up. Make sure the remaining users can
**  tell that logged-off users are not online
*/
func TestSomeFailures(t* testing.T) {
	size := 10
	failures := 3
	users := registerMany(size)
	defer killAll(users)
	for i:=0; i < size-1; i++ {
		go sendAndCheck(t, users[i], users[i+1])
	}
	go sendAndCheck(t, users[size-1], users[0])
	killAll(users[:failures])
	time.Sleep(time.Second)
	for _, aliveNode := range users[failures:] {
		for _, otherAliveNode := range users[failures:] {
			checkLookup(t, aliveNode, otherAliveNode)
		}
		for _, deadNode := range users[:failures] {
			assertEqual(t, aliveNode.IsOnline(deadNode.Name), false)
		}
	}

}

func switchIp(users []*User, startPort int) {
	p := startPort
	for i:=0;i<len(users);i++ {
		user := users[i]
		name := user.Name
		user.Logoff()
		ipAddr := localIp + ":" + strconv.Itoa(p + 8000)
		p++
		newUser := Login(name, ipAddr)
		users[i] = newUser
	}
	time.Sleep(time.Second)
	for _, user := range users {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
}

/*
**  RegisterAndLogin 3 users. Have one log off and then log back on
**  with the same IP address. Make sure that user can still
**  lookup the other users
*/
func TestPersistance(t* testing.T) {
	users := registerMany(3)
	user := users[0]
	fmt.Println(user.Name)
	users[0].Serialize()
	sucess, new_user := Deserialize("0")
	fmt.Println(sucess, new_user.Name)
	// defer killAll(users)
	// users[0].Logoff()
	// Login("0", localIp + ":8000")
	// newUser := Login("0", localIp + ":8000")
	// defer killAll([]*User{newUser})
	time.Sleep(time.Second)
	// checkLookup(t, newUser, users[1])
	// checkLookup(t, newUser, users[2])
}

/*
**  RegisterAndLogin 3 users. Have one log off and then log back on
**  with a new IP address. Make sure other users can lookup
**  that user's new address.
*/
func TestNewIP(t* testing.T) {
	size := 3
	newIpNum := 1
	users := registerMany(size)
	defer killAll(users)
	switchIp(users[:newIpNum], size+1)
	time.Sleep(time.Second)
	for _, olduser := range users[newIpNum:] {
		for _, newuser := range users[:newIpNum] {
			checkLookup(t, olduser, newuser)
		}
	}
}

func slowRegisterMany(n int, t int) []*User{
	users := make([]*User, n)

	bootstrap := ""

	for i :=0; i < n; i++{
		username := strconv.Itoa(i)
		ipAddr := localIp + ":" + strconv.Itoa(i + 8000)
		user := RegisterAndLogin(username, ipAddr, bootstrap)
		bootstrap = localIp + ":" + strconv.Itoa(i + 8000)
		time.Sleep(time.Millisecond * time.Duration(rand.Int() % (t*1000/n)))
		users[i] = user
	}

	time.Sleep(time.Second)

	for _, user := range(users) {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
	time.Sleep(time.Second)
	return users
}

const sendprob = 80
const logoffprob = 20

const logonprob = 50
const changeipprob = 30

func randomOnAction(t *testing.T, idx int, on_users, off_users []*User) ([]*User, []*User) {
	val := rand.Int() % 100
	user := on_users[idx]
	if val < sendprob {
		ridx := rand.Int() % len(on_users)
		go sendAndCheck(t, on_users[idx], on_users[ridx])
	} else {
		user.Logoff()
		on_users = append(on_users[:idx], on_users[idx+1:]...)
		off_users = append(off_users, user)
	}

	return on_users, off_users
}

func randomOffAction(t *testing.T, idx int, on_users, off_users []*User) ([]*User, []*User) {
	val := rand.Int() % 100
	if val < logonprob {
		
	}
	return on_users, off_users
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
	size := 30
	rounds := 5
	on_users := slowRegisterMany(size, 10)
	off_users := make([]*User, size)
	for r := 0; r<rounds; r++ {
		for i:=0; i<len(on_users); i++  {
			on_users, off_users = randomOnAction(t, i, on_users, off_users)
		}
		time.Sleep(time.Millisecond * 10)
	}
}

/*
**  TestRealLife but with offline messaging. 
**
*/
func TestSomething(t* testing.T) {
    //TODO: implement and name this test
}
