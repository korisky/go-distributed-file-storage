package main

import (
	"github.com/roylic/go-distributed-file-storage/p2p/tcp"
	"log"
)

func main() {
	transport := tcp.NewTCPTransport(":3999")

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	// block
	select {}
}
