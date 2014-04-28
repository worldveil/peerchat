package dht

type User struct {
	Node DhtNode
	Name string
	Message map[string]string
}