package main

func main() {
		
    peer := new(Peer)
    rpc.Register(peer)
    listener, e := net.Listen("tcp", ":1234")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    
    for {
        if conn, err := listener.Accept(); err != nil {
            log.Fatal("accept error: " + err.Error())
        } else {
            log.Printf("new connection established\n")
            go rpc.ServeConn(conn)
        }
    }
}