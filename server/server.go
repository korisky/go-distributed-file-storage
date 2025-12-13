package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/roylic/go-distributed-file-storage/crypto"
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
	// -> using another goroutine (in the background)
	s.errCh = make(chan error, 1)
	go s.loop()

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
	log.Printf("[%s] connected with remote:%s\n", p.LocalAddr(), p.RemoteAddr())
	return nil
}

// Get file from storage
func (s *FileServer) Get(key string) (io.Reader, error) {
	// have key, just return
	if s.Storage.Has(key) {
		log.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, reader, err := s.Storage.Read(key)
		return reader, err
	}

	// do not have key, broadcast for finding
	log.Printf("server[%s] Do not have file %s locally, fetching...",
		s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		// TODO first bytes to read file size, without hanging
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		// decrypt
		n, err := s.Storage.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		//n, err := s.Storage.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		log.Printf("[%s] received (%d) bytes over network from (%s)\n",
			s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	// do not have key
	_, reader, err := s.Storage.Read(key)
	return reader, err
}

// Store contains below duties
// 1) *Store* this file to disk
// 2) *Broadcast* send message to the peers, telling what we got
// 3)
func (s *FileServer) Store(key string, r io.Reader) error {

	// use teeReader to copy the reader, or else
	// the read could only be used once, later
	// broadcast only received empty
	var (
		fileBuf   = new(bytes.Buffer)
		teeReader = io.TeeReader(r, fileBuf)
	)

	// 1) Storage, after write, the reader r is empty
	size, err := s.Storage.Write(key, teeReader)
	if err != nil {
		return err
	}
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size + 16, // with 16 bytes of AES key
		},
	}

	// 2) broadcast message (message type)
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// 3) broadcast file stream concurrently
	time.Sleep(time.Millisecond * 5)

	// writing to peers concurrently (by multi-writer)
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.INCOMING_STREAM})
	n, err := crypto.CopyEncrypt(s.EncKey, fileBuf, mw)
	if err != nil {
		return err
	}

	log.Printf("server[%s] recv & writtern %d bytes to disk\n",
		s.Transport.Addr(), n)
	return nil
}

func (s *FileServer) Err() <-chan error {
	return s.errCh
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
				log.Printf("server[%s] decoding error:%s", s.Transport.Addr(), err)
				continue // don't fatal on decode error
			}

			// 2) handle message (Storage)
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
		// First byte to indicate the msg type
		peer.Send([]byte{p2p.INCOMING_MESSAGE})
		// Second then send the rest (file-buffer)
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
		log.Printf("server[%s] is attempting to connect with remote:%s\n",
			s.Transport.Addr(), addr)
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
