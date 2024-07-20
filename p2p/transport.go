package p2p

// Peer is an interface that represents the remote node
type Peer interface {
	Close() error
}

// Transport is anything that handles the communications
// between the nodes in the network. This can be of the
// form (TCP, UDP, WebSocket, etc.)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC // consume RPC (read only)
	Close() error
}
