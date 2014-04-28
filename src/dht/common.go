package dht

type AnnouceUserArgs struct {
	QueryingNodeId *big.Int
	QueryingIpAddr string
	AnnoucedUsername string
}

type AnnouceUserReply struct {
	QueriedNodeId *big.int
}

type FindNodeArgs struct {
	QueryingNodeId *big.Int
	TargetNodeId *big.Int
}

type FindNodeReply struct {
	QueriedNodeId *big.Int
	TryNodes string[] // if list is of length 1, then we found it
}

type GetUserArgs struct {
	QueryingNodeId *big.Int
	TargetUsername *big.Int
}

type GetUserReply struct {
	QueriedNodeId *big.Int
	TryNodes string[] // if list is of length 1, then we found it
}

type PingArgs struct {
	PingingNodeId *big.Int
}

type PingReply struct {
	PingedNodeId *big.Int
}

func Sha1(s string) *big.Int {
	/*
		Returns a 160 bit integer based on a
		string input. 
	*/
    h := sha1.New()
    h.Write([]byte(s))
    bs := h.Sum(nil)
    bi := new(big.Int).SetBytes(bs)
    return bi
}

func Xor(a, b *big.Int) *big.Int {
	/*
		Zors together two big.Ints and
		returns the result.
	*/
	return new(big.Int).Xor(a, b)
}