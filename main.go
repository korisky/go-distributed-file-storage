package main

import (
	"bytes"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/server"
	"github.com/roylic/go-distributed-file-storage/storage"
	"io"
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
	time.Sleep(time.Millisecond * 500)

	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(time.Second * 2)

	// examine the broadcast feature
	key := "MyPrivateData"
	data := bytes.NewReader([]byte("my big data file"))
	_ = s2.Store(key, data)

	// check file storage
	time.Sleep(time.Millisecond * 500)
	fileReader, err := s2.Get(key)
	if err != nil {
		log.Fatal("Err", err)
	}

	storedFileBytes, err := io.ReadAll(fileReader)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(string(storedFileBytes))

	select {}
}
