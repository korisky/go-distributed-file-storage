package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/roylic/go-distributed-file-storage/crypto"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"io"
	"log"
)

// handleMessage will Storage the message from broadcast
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

// handleMessageStoreFile specific handle message for Storage file
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	log.Printf("server[%s] recv %+v\n", s.Transport.Addr(), msg)

	// got the peer & let Conn receive the consumption result
	peer, exist := s.peers[from]
	if !exist {
		return fmt.Errorf("peer (%s) not found in Mapping, end handleMessage logic", from)
	}

	// decrypt first
	var decryptedBuf bytes.Buffer
	// 由于TCPPeer包含net.Conn, 并且net.Conn接口实现了Read接口,
	// 由于网络流并不包含EOF, 使用LimitReader进行封装
	// 所以可以被当作是io.Reader放入, 可以被读出内容
	_, err := crypto.CopyDecrypt(s.EncKey, io.LimitReader(peer, msg.Size), &decryptedBuf)
	if err != nil {
		return err
	}

	// read decrypted bytes
	size, err := s.Storage.Write(msg.Key, &decryptedBuf)

	if err != nil {
		return err
	}
	log.Printf("server[%s], writtern %d recv bytes to disk\n",
		s.Transport.Addr(), size)

	// callback to this Conn's loop
	//peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()

	return nil
}

// handleMessageGetFile handle get file request from other node
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {

	// 1) 如果本地没有 (通过map确认), 也就返回内容了
	if !s.Storage.Has(msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s), but it does not exist on disk\n", s.Transport.Addr(), msg.Key)
	}

	// 2) 如果本地有, 需要向请求方write回去数据流 (peer.Send 通过TCP传输s)
	log.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	// 获取目标的文件的reader & fileSize (记得关闭)
	fSize, r, err := s.Storage.Read(msg.Key)
	if err != nil {
		return err
	}
	// 确认可以转换为ReadCloser接口
	rc, ok := r.(io.ReadCloser)
	if ok {
		// defer关闭流
		defer rc.Close()
	}

	// 找到该peer的conn连接
	requestPeer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("[%s] found peer %s had not connected", s.Transport.Addr(), from)
	}

	// first send the 'incoming-Stream' byte to the peer
	requestPeer.Send([]byte{p2p.INCOMING_STREAM})

	// then can send the file size as an int64
	binary.Write(requestPeer, binary.LittleEndian, fSize)
	n, err := io.Copy(requestPeer, r)
	if err != nil {
		return err
	}

	log.Printf("[%s] written (%d) byets over the network to %s\n", s.Transport.Addr(), n, from)
	return nil
}
