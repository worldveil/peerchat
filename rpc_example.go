package main

import "example"
import "log"

func main() {	
	
	one := example.MakeNode("127.0.0.1", "55555")
	log.Printf("making one...")
	two := example.MakeNode("127.0.0.1", "55554")
	log.Printf("making two...")
	
	one.Ping(two.Address, "Hello, this is one!")
	two.Ping(one.Address, "This is two...")
	one.Ping(two.Address, "Cool")
	two.Ping(one.Address, "Last message.")
}
