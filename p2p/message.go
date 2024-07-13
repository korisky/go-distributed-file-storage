package p2p

import "net"

// RPC holds data over transport
type RPC struct {
	From    net.Addr
	Payload []byte
}
