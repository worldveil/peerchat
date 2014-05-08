package dht

import "testing"
import "runtime"
import "github.com/pmylund/sortutil"

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
}

