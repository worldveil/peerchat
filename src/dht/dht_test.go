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

func checkLookup(t *testing.T, user1 User, user2 User) {
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

func sliceEqual(a, b [IDLen][]RoutingEntry) bool {
	for idx, bucketa := range a{
		bucketb := b[idx]
		if ! isEqualRE(bucketa, bucketb){
			return false
		}
	}
	return true
}

func registerMany(num_users int) []User{
	users := make([]User, num_users)

	bootstrap := ""

	for i :=0; i < num_users; i++{
		base := 8000
		username := strconv.Itoa(i)
		ipAddr := localIp + ":" + strconv.Itoa(i + base)
		user := RegisterAndLogin(username, ipAddr, bootstrap)
		bootstrap = localIp + ":" + strconv.Itoa(i + base)
		time.Sleep(time.Millisecond * 5)
		users[i] = *user
	}

	time.Sleep(time.Second)

	for _, user := range(users) {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
	time.Sleep(time.Second)
	return users

}

func killAll(users []User){
	for _, user := range users {
		user.Logoff()
	}
	time.Sleep(time.Millisecond * 400)
}


func slowRegisterMany(n int, t int) []User{
	users := make([]User, n)

	bootstrap := ""

	for i :=0; i < n; i++{
		username := strconv.Itoa(i)
		ipAddr := localIp + ":" + strconv.Itoa(i + 8000)
		user := RegisterAndLogin(username, ipAddr, bootstrap)
		bootstrap = localIp + ":" + strconv.Itoa(i + 8000)
		time.Sleep(time.Millisecond * time.Duration(rand.Int() % (t*1000/n)))
		users[i] = *user
	}

	time.Sleep(time.Millisecond*200)

	for _, user := range(users) {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
	time.Sleep(time.Millisecond*200)
	return users
}

const sendprob = 80
const logoffprob = 20

const logonprob = 50
const changeipprob = 30

func randomOnAction(t *testing.T, idx int, on_users []User, off_users []DeadUser) ([]User, []DeadUser) {
	val := rand.Int() % 100
	user := on_users[idx]
	if val < sendprob {
		ridx := rand.Int() % len(on_users)
		if idx != ridx {
			sender:= on_users[idx]
			receiver := on_users[ridx]
			sendAndCheck(t, sender, receiver)
		}
		
	} else {
		du := DeadUser{name: user.Name, ipAddr: user.Node.IpAddr}
		user.Logoff()
		on_users = append(on_users[:idx], on_users[idx+1:]...)
		off_users = append(off_users, du)
	}

	return on_users, off_users
}

func randomOffAction(t *testing.T, idx int, on_users []User, off_users []DeadUser, ipCounter int) ([]User, []DeadUser, int) {
	val := rand.Int() % 100
	if val < logonprob {
		off_user := off_users[idx]
		ip := off_user.ipAddr
		if rand.Int() %100 < changeipprob {
			ipCounter++
			int_ip := 8000 + ipCounter
			ip = localIp + ":" + strconv.Itoa(int_ip)
		}
		newUser := Login(off_user.name,ip)
		time.Sleep(time.Millisecond * 10)
		newUser.Node.AnnounceUser(newUser.Name, ip)
		time.Sleep(time.Millisecond * 10)
		on_users = append(on_users, *newUser)
		off_users = append(off_users[:idx], off_users[idx+1:]...)
	}
	return on_users, off_users, ipCounter
}

type DeadUser struct {
	name string
	ipAddr string
}

/*
**  Register 40 users over 10 seconds. 
**  For 10 rounds, have online users randomly send messages and logoff, and
**  have offline users randomly log back in, sometimes with a new ip address
*/
func TestRealLife(t* testing.T) {
	fmt.Println("Running TestRealLife")
	defer fmt.Println("Passed!")

	rounds := 5
	ipCounter := 30
	on_users := slowRegisterMany(30, 10)
	off_users := make([]DeadUser, 0)
	for r := 0; r<rounds; r++ {
		fmt.Println("Round", r)
		for i:=0; i<len(on_users); i++  {
			if len(on_users) <= 3{
				break
			}
			on_users, off_users = randomOnAction(t, i, on_users, off_users)
		}
		time.Sleep(time.Millisecond*200)
		for i:=0; i<len(off_users); i++  {
			on_users, off_users, ipCounter = randomOffAction(t, i, on_users, off_users, ipCounter)
		}
		time.Sleep(time.Millisecond*200)
	}
	killAll(on_users)
}

func TestSerialization(t *testing.T) {
	fmt.Println("Running TestSerialization")
	defer fmt.Println("passed")
	
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
		assertEqual(t, sliceEqual(one.Node.RoutingTable, user1.Node.RoutingTable), true)
		assertEqual(t, sliceEqual(two.Node.RoutingTable, user2.Node.RoutingTable), true)
	}

	user1.Logoff()
	user2.Logoff()
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
**	 saved intact with no duplication.
**	
**	We verify the messages are not lost
**	and arrive unaltered. 
*/
func TestBasic(t *testing.T) {
	fmt.Println("Running TestBasic")
	defer fmt.Println("passed")

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
	assertEqual(t, user1.GetMessagesFrom(user2)[1].Content, msg2)
	
	assertEqual(t, len(user2.MessageHistory[username1]), 2)
	assertEqual(t, len(user1.MessageHistory[username2]), 2)
	
	// kill user nodes
	user1.Logoff()
	user2.Logoff()
}

/*
**  RegisterAndLogin 20 users and make sure that each user can lookup
**  the IP address of every other user
*/
func TestManyRegistrations(t *testing.T) {
	fmt.Println("Running TestManyRegistrations")	
	users := registerMany(20)
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
	
	size := 100
	//filename := fmt.Sprintf("/Users/will/Code/Go/peerchat/writeup/plots/SWEEP.csv")
	//os.Create(filename)

	fmt.Println("Running TestManyMoreRegistrations")
	users := registerMany(size)
	defer killAll(users)
	for i:=0; i<20; i++ {
		idx :=  rand.Int() % size
		idx2 := rand.Int() % size
		checkLookup(t, users[idx], users[idx2])
		checkLookup(t, users[idx2], users[idx])
	}
	fmt.Println("Passed!")
}
func sendAndCheck(t *testing.T, sender User, receiver User) {
	msg := "message " + strconv.Itoa(rand.Int() % 1000)
	idx := len(receiver.MessageHistory[sender.Name])
	sender.SendMessage(receiver.Name, msg)
	
	for i:=0; i<100; i++{
		if len(receiver.MessageHistory[sender.Name]) > idx {
			assertEqual(t, receiver.MessageHistory[sender.Name][idx].Content, msg)
			return
		}
		time.Sleep(time.Millisecond*50)
	}
	t.Fatalf("message not received")
	
}

/*
**  RegisterAndLogin 5 users. Make sure they can all send messages
**  to each other
*/
func TestSends(t *testing.T) {
	fmt.Println("Running TestSends")
	size := 20
	users := registerMany(80)
	defer killAll(users)
	for i:=0; i < size; i++ {
		for j:=0; j<size; j++ {
			sendAndCheck(t, users[i], users[j])
		}
	}
	fmt.Println("Passed!")
}

/*
**  RegisterAndLogin 10 users. Have 3 go offline. Make sure the remaining 
**  users can look each other up. Make sure the remaining users can
**  tell that logged-off users are not online
*/
func TestSomeLogoffs(t* testing.T) {
	fmt.Println("Running TestSomeLogoffs")
	defer fmt.Println("Passed!")
	size := 10
	failures := 3
	users := registerMany(size)
	defer killAll(users)
	for i:=0; i < len(users)-1; i++ {
		go sendAndCheck(t, users[i], users[i+1])
	}
	sendAndCheck(t, users[size-1], users[0])
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

func switchIp(users []User, startPort int) []User{
	p := startPort
	for i:=0;i<len(users);i++ {
		user := users[i]
		name := user.Name
		user.Logoff()
		time.Sleep(time.Second)
		ipAddr := localIp + ":" + strconv.Itoa(p + 8000)
		p++
		newUser := Login(name, ipAddr)
		users[i] = *newUser
	}
	time.Sleep(time.Second)
	for _, user := range users {
		user.Node.AnnounceUser(user.Name, user.Node.IpAddr)
	}
	return users
}

/*
**  RegisterAndLogin 3 users. Have one log off and then log back on
**  with the same IP address. Make sure that user can still
**  lookup the other users
*/
func TestPersistance(t* testing.T) {
	fmt.Println("Running TestPersistance")
	defer fmt.Println("Passed!")

	users := registerMany(10)
	defer killAll(users)
	users[1].Logoff()
	time.Sleep(time.Millisecond * 50)
	newUser := Login("1", localIp + ":8001")
	time.Sleep(time.Millisecond * 50)
	newUser.Logoff()
	time.Sleep(time.Millisecond * 50)
	newUser = Login("1", localIp + ":8001")
	defer killAll([]User{*newUser})
	time.Sleep(time.Millisecond * 50)
	checkLookup(t, *newUser, users[0])
	checkLookup(t, *newUser, users[2])
	sendAndCheck(t, *newUser, users[2])
	sendAndCheck(t, users[2], *newUser)
}

/*
**  RegisterAndLogin 10 users. Have one log off and then log back on
**  with a new IP address. Make sure other users can lookup
**  that user's new address.
*/
func TestNewIP(t* testing.T) {
	fmt.Println("Running TestNewIP")
	defer fmt.Println("Passed!")

	size := 10
	newIpNum := 5
	users := registerMany(size)
	users = append(switchIp(users[:newIpNum], size+1), users[newIpNum:]...)
	defer killAll(users)
	time.Sleep(time.Second)
	for _, olduser := range users[newIpNum:] {
		for _, newuser := range users[:newIpNum] {
			checkLookup(t, olduser, newuser)
		}
	}
}

func TestOfflineChat(t* testing.T) {
	fmt.Println("Running TestOfflineChat")
	defer fmt.Println("Passed!")

	size := 10
	users := registerMany(size)
	defer killAll(users)
	oldip := users[0].Node.IpAddr
	users[0].Logoff()
	users[1].SendMessage("0", "hello")
	time.Sleep(time.Second)
	newUser := Login("0", oldip)
	time.Sleep(time.Second)
	assertEqual(t, newUser.MessageHistory["1"][0].Content, "hello")
	newUser.Logoff()
	time.Sleep(time.Second)
}

//receiver goes offline, then sender goes offline, then receiver comes back- should get message
func TestDualOfflineChat(t* testing.T) {
	fmt.Println("Running TestDualOfflineChat")
	defer fmt.Println("Passed!")

	size := 50
	users := registerMany(size)
	defer killAll(users)
	user0 := users[0]
	user1 := users[1]
	oldip := user0.Node.IpAddr
	sendAndCheck(t, user0, user1)
	sendAndCheck(t, user1, user0)
	user0.Logoff()
	time.Sleep(time.Millisecond*200)
	user1.SendMessage("0", "hello")
	time.Sleep(time.Millisecond*200)
	user1.Logoff()
	time.Sleep(time.Millisecond*200)
	user0 = *Login("0", oldip)
	time.Sleep(time.Millisecond*200)
	// assert that user0 can see message, even though 1 is offline!
	assertEqual(t, user0.MessageHistory["1"][len(user0.MessageHistory["1"]) - 1].Content, "hello")
	user0.SendMessage("1", "hi")
	time.Sleep(time.Millisecond*200)
	user0.Logoff()
	time.Sleep(time.Millisecond*200)
	ipAddr := localIp + ":" + strconv.Itoa(8100)
	user1 = *Login("1", ipAddr)
	time.Sleep(time.Millisecond*500)
	// assert that user1 can see message, even though 0 is offline! Note that user1 has changed ips
	assertEqual(t, user1.MessageHistory["0"][len(user1.MessageHistory["0"]) - 1].Content, "hi")
	user1.Logoff()
} 
