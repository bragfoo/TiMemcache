package process

import (
	"log"
	"net"
	"os"
)

// CreateServer is creat tcp server
func CreateServer() {
	host := "0.0.0.0"
	port := "11212"
	serv, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Println("Error listening:", err)
		os.Exit(1)
	}
	defer serv.Close()
	log.Println("Memcache proxy for TiKV on " + host + ":" + port)
	// read from connect
	for {
		// client
		conn, err := serv.Accept()
		if err != nil {
			log.Println("Error accepting: ", err)
			os.Exit(1)
		}
		log.Printf("Create connection %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		// dispose request
		go disposeRequest(conn)
	}
}
