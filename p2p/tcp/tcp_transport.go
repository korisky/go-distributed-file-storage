package tcp

import (
	"github.com/roylic/go-distributed-file-storage/p2p"
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]p2p.Peer
}

func NewTCPTransport(listenAdd string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAdd,
	}
}
