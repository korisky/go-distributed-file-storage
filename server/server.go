package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/storage"
	"io"
	"log"
	"strings"
	"sync"
)

// FileServerOpts inner Transport is for accepting the p2p communication
type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *storage.Storage
	quitCh chan struct{}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := storage.StorageOpt{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		peers:          make(map[string]p2p.Peer),
		store:          storage.NewStore(storageOpts),
		quitCh:         make(chan struct{}),
	}
}

// Start the server
func (s *FileServer) Start() error {
	// port listening
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	// bootstrap the network
	_ = s.bootstrapNetwork()

	// init the gob, for encode & decoding
	gob.Register(Message{})
	gob.Register(MessageStoreFile{})

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
	log.Printf("Connected with remote:%s, cur local:%s\n",
		p.RemoteAddr(), p.LocalAddr())
	return nil
}

// StoreData contains below duties
// 1) *Store* this file to disk
// 2) *Broadcast* this file to all known peers in the network
func (s *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	msg := &Message{
		Payload: MessageStoreFile{Key: key},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Fatal("Error during encoding message", err)
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	// TODO simulate big file keep sending
	// TODO time consuming
	//time.Sleep(time.Second * 3)
	//payload := []byte("This is Large File")
	//for _, peer := range s.peers {
	//	if err := peer.Send(payload); err != nil {
	//		log.Fatal(err)
	//	}
	//}

	return nil

	//// use teeReader to copy the reader, or else
	//// the read could only be used once, later
	//// broadcast only received empty
	//buf := new(bytes.Buffer)
	//teeReader := io.TeeReader(r, buf)
	//
	//// 1) store, after write, the reader r is empty
	//if err := s.store.Write(key, teeReader); err != nil {
	//	return err
	//}
	//
	//// 2) broadcast (PayLoad struct)
	//p := &DataMessage{
	//	Key:  key,
	//	Data: buf.Bytes(),
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

			// decode the msg (blocking)
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Receive: %s\n", msg.Payload)

			// check msg source
			peer, exist := s.peers[rpc.From]
			if !exist {
				panic("Peer not found in Mapping")
			}

			// TODO check peer
			//b := make([]byte, 1000)
			//if _, err := peer.Read(b); err != nil {
			//	panic(err)
			//}
			//fmt.Printf("%s\n", string(b))
			peer.(*p2p.TCPPeer).Wg.Done()
			//// handle the received broadcast msg
			//if err := s.handleMessage(&m); err != nil {
			//	log.Fatal(err)
			//}
		// server stop
		case <-s.quitCh:
			return
		}
	}
}

// bootstrapNetwork is for dialing to other port
func (s *FileServer) bootstrapNetwork() error {
	// for each node, make a new goroutine for dialing it
	for _, addr := range s.BootstrapNodes {
		if len(strings.TrimSpace(addr)) == 0 {
			continue
		}
		// only when addr is not empty
		log.Println("Attempting to connect with remote: ", addr)
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error during BootstrapNetwork(): ", err)
			}
		}(addr)
	}
	return nil
}

// broadcast will help send all coding msg to cur node's peer
func (s *FileServer) broadcast(m *Message) error {
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

//// handleMessage will store the message from broadcast
//func (s *FileServer) handleMessage(m *Message) error {
//	switch v := m.Payload.(type) {
//	case *DataMessage:
//		fmt.Printf("received data %+v\n", v)
//	}
//	return nil
//}
