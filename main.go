package main

import (
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/p2p/tcp"
	"log"
)

func main() {

	tcpOpts := tcp.TCPTransportOpt{
		ListenAddr:    ":3999",
		HandshakeFunc: p2p.NopHandshakeFunc,
	}

	transport := tcp.NewTCPTransport(tcpOpts)

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	// block
	select {}
}
