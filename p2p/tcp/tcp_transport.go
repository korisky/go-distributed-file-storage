package tcp

import (
	"fmt"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"net"
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

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpt struct {
	ListenAddr    string
	HandshakeFunc p2p.HandshakeFunc
	Decoder       p2p.Decoder
	OnPeer        func(peer p2p.Peer) error
}

type TCPTransport struct {
	TCPTransportOpt // 直接放入, 类似Java继承的意思, 可直接操作属性
	listener        net.Listener
	rpcCh           chan p2p.RPC
}

func NewTCPTransport(opts TCPTransportOpt) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpt: opts,
		rpcCh:           make(chan p2p.RPC),
	}
}

// Consume 只能从Channel中读, 不能写
// chan   // read-write
// <-chan // read only
// chan<- // write only
func (t *TCPTransport) Consume() <-chan p2p.RPC {
	return t.rpcCh
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

	// 该handleConn方法内所有异常导致return前都会执行conn.Close()
	var err error
	defer func() {
		fmt.Printf("dropping peer connection:%s\n", err)
		conn.Close()
	}()

	// 针对新连接, 创建Peer
	peer := NewTCPPeer(conn, true)

	// 尝试握手
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	// 存在OnPeer方法时进行调用
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// 循环读取
	rpc := p2p.RPC{}
	for {

		if err := t.Decoder.Decoder(conn, &rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcCh <- rpc
	}

}
