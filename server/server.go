package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/roylic/go-distributed-file-storage/p2p"
)

// Start the server
func (s *FileServer) Start() error {
	// port listening
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	// bootstrap the network
	_ = s.bootstrapNetwork()

	// looping accept msg
	s.loop()

	return nil
}

// Stop will use to close a channel
func (s *FileServer) Stop() {
	close(s.quitCh)
}

// OnPeer handle peer connection
func (s *FileServer) OnPeer(p p2p.Peer) error {
	// lock for adding & unlock for later
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	// put into map
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("server[%s] connected with remote:%s \n",
		p.LocalAddr(), p.RemoteAddr())
	return nil
}

// Get file from storage
func (s *FileServer) Get(key string) (io.Reader, error) {
	// have key, just return
	if s.store.Has(key) {
		log.Printf("server[%s] found file %s locally, sending...",
			s.FileServerOpts.StorageRoot, key)
		return s.store.Read(key)
	}

	// do not have key, broadcast for finding
	log.Printf("server[%s] Do not have file %s locally, fetching...",
		s.FileServerOpts.StorageRoot, key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range s.peers {
		fileBuffer := new(bytes.Buffer)
		n, err := io.CopyN(fileBuffer, peer, 10)
		if err != nil {
			// TODO
			return nil, err
		}
		log.Printf("server[%s] received %d bytes over network\n",
			s.FileServerOpts.StorageRoot, n)
		log.Println(fileBuffer.String())
	}

	select {}

	// do not have key
	return nil, nil
}

// Store contains below duties
// 1) *Store* this file to disk
// 2) *Broadcast* this file to all known peers in the network
func (s *FileServer) Store(key string, r io.Reader) error {

	// use teeReader to copy the reader, or else
	// the read could only be used once, later
	// broadcast only received empty
	var (
		fileBuf   = new(bytes.Buffer)
		teeReader = io.TeeReader(r, fileBuf)
	)

	// 1) store, after write, the reader r is empty
	size, err := s.store.Write(key, teeReader)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	// 2) broadcast
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// TODO simulate big file keep sending
	// TODO time consuming
	// TODO user multi writer
	time.Sleep(time.Millisecond * 300)
	for _, peer := range s.peers {
		// first byte for message type indication
		peer.Send([]byte{p2p.INCOMING_STREAM})
		// then send the file
		n, err := io.Copy(peer, fileBuf)
		if err != nil {
			return err
		}
		fmt.Printf("server[%s] recv & writtern %d bytes\n",
			s.FileServerOpts.StorageRoot, n)
	}
	return nil

	//// use teeReader to copy the reader, or else
	//// the read could only be used once, later
	//// broadcast only received empty
	//fileBuf := new(bytes.Buffer)
	//teeReader := io.TeeReader(r, fileBuf)
	//
	//// 1) store, after write, the reader r is empty
	//if err := s.store.Write(key, teeReader); err != nil {
	//	return err
	//}
	//
	//// 2) broadcast (PayLoad struct)
	//p := &DataMessage{
	//	Key:  key,
	//	Data: fileBuf.Bytes(),
	//}
	//return s.broadcast(&Message{
	//	From:    "TODO",
	//	Payload: p,
	//})
}

// loop is for continuing retrieve msg from Transport channel
// receive the *Broadcast* & *Store* the data
func (s *FileServer) loop() {
	defer func() {
		log.Printf("File Server Stop")
		_ = s.Transport.Close()
	}()
	// crucial looping
	for {
		select {
		// retrieve msg from read only channel
		case rpc := <-s.Transport.Consume():

			// 1) decode the msg (blocking)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error:", err)
			}

			// 2) handle message (store)
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handling message error:", err)
			}

		// server stop
		case <-s.quitCh:
			return
		}
	}
}

// broadcast will send message to all peers
func (s *FileServer) broadcast(m *Message) error {
	// form msg
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(m); err != nil {
		log.Fatal("Error during encoding message", err)
	}
	// send
	for _, peer := range s.peers {
		// first byte to indicate the msg type
		peer.Send([]byte{p2p.INCOMING_MESSAGE})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// stream all coding msg to cur node's peer
// Deprecated: use  broadcast instead
func (s *FileServer) stream(m *Message) error {

	// append to temp slice
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	// concurrently streaming the same msg to all peers
	// go and check description for MultiWriter()
	// 由于net.Conn实现了Writer接口的Write方法, 所以继承了net.Conn的Peer
	// 可以传入允许Writer的方法
	mu := io.MultiWriter(peers...)

	// so now by insert the multi-writer and pass the DataMessage in it,
	// it will encode the DataMessage to bytes, and pass the bytes to
	// multi-writer (only one copy). The multi-writer will copy it to
	// all net.Conn connection (might be Zero-Copy in it)
	return gob.NewEncoder(mu).Encode(m)
}

// bootstrapNetwork is for dialing to other port
func (s *FileServer) bootstrapNetwork() error {
	// init the gob, for encode & decoding
	initTypeRegistration()
	// for each node, make a new goroutine for dialing it
	for _, addr := range s.BootstrapNodes {
		if len(strings.TrimSpace(addr)) == 0 {
			continue
		}
		// only when addr is not empty
		log.Printf("server[%s] is attempting to connect with remote:%s\n", s.FileServerOpts.StorageRoot, addr)
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error during BootstrapNetwork(): ", err)
			}
		}(addr)
	}
	return nil
}

// initTypeRegistration 需要将可能deserializing的struct进行提前注册
func initTypeRegistration() {
	gob.Register(Message{})
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
