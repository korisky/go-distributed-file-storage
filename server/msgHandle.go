package server

import (
	"fmt"
	"io"
	"log"

	"github.com/roylic/go-distributed-file-storage/p2p"
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
	log.Printf("server[%s] recv %+v\n", s.Transport.Addr(), msg)

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
	log.Printf("server[%s], writtern %d recv bytes to disk\n",
		s.Transport.Addr(), size)

	// callback to this Conn's loop
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

// handleMessageGetFile handle get file request from other node
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	log.Printf("server[%s] recv key:%s\n", s.Transport.Addr(), msg)

	// 1) 如果本地没有 (通过map确认), 也就返回内容了
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("server[%s] not found file :%s, end handleMessage logic\n",
			s.Transport.Addr(), msg.Key)
	}

	// 2) 如果本地有, 需要向请求方write回去数据流 (peer.Send 通过TCP传输s)
	log.Printf("server[%s] serving file with key %s over the network\n",
		s.Transport.Addr(), msg.Key)
	// 获取目标的文件的reader
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}
	// 找到该peer的conn连接
	requestPeer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("server[%s] found peer %s had not connected",
			s.Transport.Addr(), from)
	}
	// copy 数据流 & 传输回去
	n, err := io.Copy(requestPeer, r)
	if err != nil {
		return err
	}
	log.Printf("server[%s] written %d byets over the network\n",
		s.Transport.Addr(), n)
	return nil
}
