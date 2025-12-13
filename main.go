package main

import (
	"bytes"
	"fmt"
	"github.com/roylic/go-distributed-file-storage/crypto"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/server"
	"github.com/roylic/go-distributed-file-storage/storage"
	"io"
	"log"
	"time"
)

// makeServer extract the server opts
func makeServer(encKey []byte, listenAddr string, nodes ...string) *server.FileServer {
	// 1. tcp options
	tcpOpts := p2p.TCPTransportOpt{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	transport := p2p.NewTCPTransport(tcpOpts)
	// 2. file server options
	fileServerOpts := server.FileServerOpts{
		EncKey:            encKey,
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
	sharedKey := crypto.NewAesKey()

	s1 := makeServer(sharedKey, ":3999", "")
	s2 := makeServer(sharedKey, ":4999", ":3999")
	s3 := makeServer(sharedKey, ":5999", ":3999", ":4999")

	servers := []*server.FileServer{s1, s2, s3}

	for i, s := range servers {
		if err := s.Start(); err != nil {
			log.Fatalf("Server %d failed to start:%v", i+1, err)
		}
		time.Sleep(time.Millisecond * 100)
	}
	//
	//go func() {
	//	log.Fatal(s1.Start())
	//}()
	//time.Sleep(time.Millisecond * 500)
	//
	//go func() {
	//	log.Fatal(s2.Start())
	//}()
	//time.Sleep(time.Millisecond * 500)
	//
	//go func() {
	//	log.Fatal(s3.Start())
	//}()
	//time.Sleep(time.Second * 1)

	for i := range 5 {
		fmt.Println()
		key := fmt.Sprintf("cool-pic_%d.png", i)

		// store 1 file
		data := bytes.NewReader([]byte(fmt.Sprintf("my big data file here! %d", i)))
		s3.Store(key, data)
		time.Sleep(5 * time.Millisecond)

		// delete
		if err := s3.Storage.Delete(key); err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)

		// get from network
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Main func -> read:", string(b))
	}

	//// C) examine multiple calling
	//for i := 0; i < 3; i++ {
	//	data := bytes.NewReader([]byte("my big data file here!"))
	//	s2.Store(fmt.Sprintf("myprivatedata_%d", i), data)
	//	time.Sleep(5 * time.Millisecond)
	//}

	//// A) examine the broadcast feature
	//key := "MyPrivateData"
	//data := bytes.NewReader([]byte("my big data file"))
	//_ = s2.Store(key, data)
	//
	//// check file storage
	//time.Sleep(time.Millisecond * 500)
	//fileReader, err := s2.Get(key)
	//if err != nil {
	//	log.Fatal("Err", err)
	//}
	//
	//storedFileBytes, err := io.ReadAll(fileReader)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf(string(storedFileBytes))

	//// B) examine the getFile (download) feature
	//// get the file reader
	////r, err := s2.Get("foo") // no file found
	////r, err := s2.Get("MyPrivateData") // file found, stored before
	//r, err := s2.Get("S1_Server_Only") // file not found, need request to other
	//if err != nil {
	//	log.Fatal(err)
	//}
	//// reader -> read the whole file
	//b, err := io.ReadAll(r)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("file downloaded: " + string(b))
	//select {}
}
