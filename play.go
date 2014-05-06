package main

// import "dht"
import "fmt"
// import "crypto/sha1"
// import "math"


func find_n(a, b uint64) uint{
	var IDLen uint
	IDLen = 64
	var d, diff uint64
	diff = a ^ b
	var i uint
	for i = 0; i < IDLen; i++{
		d = 1<<(IDLen - 1 - i)
		if d & diff != 0 { // if true, return i
			break
		}
	}
	fmt.Println(i)
	return i
}

type myStruct struct{
	a int
	b string
}

func main() {
	arr := make([]myStruct, 0, 5)
	fmt.Println(arr)
	


	
}
