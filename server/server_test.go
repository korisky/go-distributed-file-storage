package server

import (
	"bytes"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"testing"
	"time"
)

// Test_FilePassing 简单建立2个服务，确认文件上传单个节点后的broadcast效果
func Test_FilePassing(t *testing.T) {
	// multi-server setting up
	// S2 监听4999端口, 主动将:3999作为连接的node
	s2 := makeServer(":4999", ":3999")
	// S1 监听3999端口
	s1 := makeServer(":3999", "")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(time.Millisecond * 500)

	go func() {
		log.Fatal(s2.Start())
	}()
	time.Sleep(time.Second * 1)

	key := "MyPrivateData"
	data := []byte("my big data file")

	// examine the broadcast feature
	_ = s2.Store(key, bytes.NewReader(data))

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

	assert.Equal(t, data, storedFileBytes)
}

// makeServer extract the server opts
func makeServer(listenAddr string, nodes ...string) *FileServer {
	// 1. tcp options
	tcpOpts := p2p.TCPTransportOpt{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	transport := p2p.NewTCPTransport(tcpOpts)
	// 2. file server options
	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         transport,
		BootstrapNodes:    nodes,
	}
	// 3. construct server
	s := NewFileServer(fileServerOpts)
	transport.OnPeer = s.OnPeer
	return s
}
