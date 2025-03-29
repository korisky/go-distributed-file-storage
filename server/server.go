package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/storage"
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
	Key  string
	Size int64
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
	log.Printf("server %s connected with remote:%s \n",
		p.LocalAddr(), p.RemoteAddr())
	return nil
}

// Get file from storage
func (s *FileServer) Get(key string) (io.Reader, error) {
	// have key
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	// do not have key
	panic("server Do not have file locally")
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
		n, err := io.Copy(peer, fileBuf)
		if err != nil {
			return err
		}
		fmt.Printf("server %s recv & writtern %d bytes\n",
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
				log.Fatal(err)
				return
			}

			// 2) handle message (store)
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Fatal(err)
				return
			}

		// server stop
		case <-s.quitCh:
			return
		}
	}
}

// handleMessage will store the message from broadcast
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
		//case
	}
	return nil
}

// handleMessageStoreFile specific handle message for store file
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	log.Printf("server [%s] recv %+v\n", s.FileServerOpts.StorageRoot, msg)

	// got the peer & let Conn receive the consumption result
	peer, exist := s.peers[from]
	if !exist {
		return fmt.Errorf("peer (%s) not found in Mapping, end handleMessage logic", from)
	}

	// store the input stream from peer
	// 由于TCPPeer包含net.Conn, 并且net.Conn接口实现了Read接口,
	// 所以可以被当作是io.Reader放入, 可以被读出内容
	// 由于网络流并不包含EOF, 使用LimitReader进行封装
	size, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	log.Printf("server %s, writtern %d recv bytes to disk\n",
		s.FileServerOpts.StorageRoot, size)

	// callback to this Conn's loop
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
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
	gob.Register(Message{})
	gob.Register(MessageStoreFile{})
	// for each node, make a new goroutine for dialing it
	for _, addr := range s.BootstrapNodes {
		if len(strings.TrimSpace(addr)) == 0 {
			continue
		}
		// only when addr is not empty
		log.Printf("server %s is attempting to connect with remote:%s\n", s.FileServerOpts.StorageRoot, addr)
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error during BootstrapNetwork(): ", err)
			}
		}(addr)
	}
	return nil
}
