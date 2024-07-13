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

type TCPTransportOpt struct {
	ListenAddr    string
	HandshakeFunc p2p.HandshakeFunc
	Decoder       p2p.Decoder
}

type TCPTransport struct {
	TCPTransportOpt // 直接放入, 类似Java继承的意思, 可直接操作属性
	listener        net.Listener
	mu              sync.RWMutex
	peers           map[net.Addr]p2p.Peer
}

func NewTCPTransport(opts TCPTransportOpt) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpt: opts,
	}
}

// ListenAndAccept 进行TCP连接的Accept启动
func (t *TCPTransport) ListenAndAccept() error {
	// 绑定port到listener
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
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
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	// 针对新连接, 创建Peer
	peer := NewTCPPeer(conn, true)
	// 尝试握手
	if err := t.HandshakeFunc(peer); err != nil {
		err := conn.Close()
		if err != nil {
			return
		}
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}
	// 循环读取
	rpc := &p2p.RPC{}
	for {

		if err := t.Decoder.Decoder(conn, rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()

		fmt.Printf("Rpc: %+v\n", rpc)
	}

}
