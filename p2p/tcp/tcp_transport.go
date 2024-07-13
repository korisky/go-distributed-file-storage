package tcp

import (
	"fmt"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"net"
	"sync"
)

// TCPPeer 代表一个通过TCP连接的远程node
type TCPPeer struct {
	// conn is the underline connection with the peer
	conn net.Conn
	// dail & retrieve a conn -> outbound = true
	// accept & retrieve a conn -> inbound= true, outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

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

// ListenAndAccept 进行TCP连接的Accept启动
func (t *TCPTransport) ListenAndAccept() error {
	// 绑定port到listener
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	// 另启线程, 开始循环accept
	go t.startAcceptLoop()
	return nil
}

// startAcceptLoop 循环监听Accept请求
func (t *TCPTransport) startAcceptLoop() {
	for {
		// accept请求
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept errors:%s\n", err)
		}
		// 另起线程, 处理conn
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	fmt.Printf("new incoming connection %+v\n", conn)

	// 针对新连接, 创建Peer

}
