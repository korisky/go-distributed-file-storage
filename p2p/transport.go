package p2p

import "net"

// Peer is an interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport is anything that handles the communications
// between the nodes in the network. This can be of the
// form (TCP, UDP, WebSocket, etc.)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC // consume RPC (read only)
	Close() error
}
