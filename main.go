package main

import (
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/server"
	"github.com/roylic/go-distributed-file-storage/storage"
	"log"
	"time"
)

func main() {

	// 1. tcp options
	tcpOpts := p2p.TCPTransportOpt{
		ListenAddr:    ":3999",
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO OnPeer:        func(p2p.Peer) error { return fmt.Errorf("failed") },
	}
	transport := p2p.NewTCPTransport(tcpOpts)

	// 2. file server options
	fileServerOpts := server.FileServerOpts{
		StorageRoot:       "fileServerStorage",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         transport,
	}

	// 3. construct server
	s := server.NewFileServer(fileServerOpts)

	// stop testing
	go func() {
		time.Sleep(time.Second * 10)
		s.Stop()
	}()

	// start the server
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

}
