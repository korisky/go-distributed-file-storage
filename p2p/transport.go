package p2p

// Peer is an interface that represents the remote node
type Peer interface {
}

// Transport is anything that handles the communications
// between the nodes in the network. This can be of the
// form (TCP, UDP, WebSocket, etc.)
type Transport interface {
}
