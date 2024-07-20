package p2p

import (
	"errors"
	"fmt"
	"log"
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

// Send implement Peer interface
func (p *TCPPeer) Send(bytes []byte) error {
	_, err := p.conn.Write(bytes)
	return err
}

// RemoteAddr implement Peer interface
// return underlining connection's address
func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

// Close implement Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpt struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpt // 直接放入, 类似Java继承的意思, 可直接操作属性
	listener        net.Listener
	rpcCh           chan RPC
}

func NewTCPTransport(opts TCPTransportOpt) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpt: opts,
		rpcCh:           make(chan RPC),
	}
}

// Dial implement the Transport interface,
// use extra goroutine for dialing to the server
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	// also call the handleConn(), but from outbound
	go t.handleConn(conn, true)

	return nil
}

// Close implement the Transport interface it
// take care of port listener's graceful closing
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Consume implement the Transport interface
// only reading from channel, below are golang's trick
// chan   // read-write
// <-chan // read only
// chan<- // write only
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

// ListenAndAccept implement the Transport interface
// 1) bind the port for tcp
// 2) looping accepting the tcp request
// 3) looping parse the incoming msg to RPC & send into channel
func (t *TCPTransport) ListenAndAccept() error {
	// 绑定port到listener
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	// 另启线程, 开始循环accept
	go t.startAcceptLoop()
	log.Printf("TCP Transport listening on port: %s\n", t.ListenAddr)
	return nil
}

// startAcceptLoop 循环监听Accept请求
func (t *TCPTransport) startAcceptLoop() {
	for {
		// accept请求
		conn, err := t.listener.Accept()
		// close error handling -> graceful shutdown
		if errors.Is(err, net.ErrClosed) {
			return
		}
		// normal error handling
		if err != nil {
			fmt.Printf("TCP accept errors:%s\n", err)
		}
		// 另起线程, 处理conn
		fmt.Printf("new incoming connection %+v\n", conn)
		go t.handleConn(conn, false)
	}
}

// handleConn with below procedures
// 1) creating new peer for each new tcp income (accept())
// 2) use customised HandshakeFunc
// 3) decode the incoming msg to RPC and put into channel (in loop)
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	// 该handleConn方法内所有异常导致return前都会执行conn.Close()
	var err error
	defer func() {
		fmt.Printf("dropping peer connection:%s\n", err)
		conn.Close()
	}()

	// 针对新连接, 创建Peer
	peer := NewTCPPeer(conn, outbound)

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
	rpc := RPC{}
	for {
		err = t.Decoder.Decoder(conn, &rpc)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				// 连接异常停止调用
				return
			} else {
				// 解码异常让其继续
				fmt.Printf("TCP error: %s\n", err)
				continue
			}
		}

		rpc.From = conn.RemoteAddr()
		t.rpcCh <- rpc
	}

}
