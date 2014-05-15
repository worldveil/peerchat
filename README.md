peerchat
=====

Distributed, P2P, realtime chat application written in golang for MIT's 6.824 distributed systems class. 

### Testing

Run the tests with:

	$ cd src/dht
	$ go get github.com/pmylund/sortutil
	$ go test 2> /dev/null
	
You should see this:

	(ml)31-35-161:dht will$ go test 2> /dev/null
	Running TestSerialization
	passed
	Running TestBasic
	passed
	Running TestManyRegistrations
	Passed!
	Running TestManyMoreRegistrations
	Passed!
	Running TestSends
	Passed!
	Running TestSomeLogoffs
	Passed!
	Running TestPersistance
	Passed!
	Running TestNewIP
	Passed!
	Running TestOfflineChat
	Passed!
	Running TestDualOfflineChat
	Passed!
	Running TestDualOfflineChat
	Passed!
	Running TestRealLife
	Passed!
	
so long as K and Alpha are set appropriately as we have them defaulted to. You may see that on less powerful computers that `TestRealLife` will fail. 

### Chatting
	
Start a chat node with:

	$ go build
	$ ./peerchat
	
Or alternatively:

	$ go run chat.go
	


