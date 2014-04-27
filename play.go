package main

import "dht"
import "fmt"

func main() {
	one := dht.Sha1("This is a string")
	two := dht.Sha1("Also a string")
	result := dht.Xor(one, two)
	fmt.Println(one, two, result)
}
