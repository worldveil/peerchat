package dht

import "testing"
import "runtime"
import "fmt"
import "time"

// Signal failures with the following:
// t.Fatalf("error message here")

func TestBasic(t *testing.T) {
	/*
		TestBasic:
		1) Starts two nodes
		2) Introduces node1 to node2
		3) Nodes send messages
		
		We verify the messages are not lost
		and arrive unaltered. 
	*/
	runtime.GOMAXPROCS(4)

	localIp := "127.0.0.1"
	port1 := ":4444"
	port2 := ":5555"
	username1 := "Alice"
	username2 := "Frans"


	user1 := Register(username1, localIp + port1, "")
	
	time.Sleep(time.Second * 1)
	
	user2 := Register(username2, localIp + port2, localIp + port1)

	fmt.Println(user1, user2)
}

