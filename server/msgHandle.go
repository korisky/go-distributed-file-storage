package server

import (
	"fmt"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"io"
	"log"
)

// handleMessage will store the message from broadcast
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
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

// handleMessageGetFile handle get file request from other node
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	fmt.Printf("server [%s] recv key:%s\n", s.FileServerOpts.StorageRoot, msg)

	// TODO get the stuff
	return nil
}
