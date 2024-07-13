package main

import (
	"fmt"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/p2p/tcp"
	"log"
)

func main() {

	tcpOpts := tcp.TCPTransportOpt{
		ListenAddr:    ":3999",
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        func(p2p.Peer) error { return fmt.Errorf("failed") },
	}

	transport := tcp.NewTCPTransport(tcpOpts)

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	// block
	select {}
}
