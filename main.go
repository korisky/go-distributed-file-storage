package main

import (
	"bytes"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/server"
	"github.com/roylic/go-distributed-file-storage/storage"
	"log"
	"time"
)

// makeServer extract the server opts
func makeServer(listenAddr string, nodes ...string) *server.FileServer {
	// 1. tcp options
	tcpOpts := p2p.TCPTransportOpt{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	transport := p2p.NewTCPTransport(tcpOpts)
	// 2. file server options
	fileServerOpts := server.FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         transport,
		BootstrapNodes:    nodes,
	}
	// 3. construct server
	s := server.NewFileServer(fileServerOpts)
	transport.OnPeer = s.OnPeer
	return s
}

func main() {

	// multi-server setting up
	s2 := makeServer(":4999", ":3999")
	s1 := makeServer(":3999", "")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(time.Second * 3)

	go s2.Start()
	time.Sleep(time.Second * 3)

	// examine the broadcast feature
	data := bytes.NewReader([]byte("my big data file"))
	_ = s2.StoreData("MyPrivateData", data)

	select {}
}
