package tcp

import (
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_TCPTransport(t *testing.T) {

	// check address

	opt := TCPTransportOpt{
		ListenAddr:    ":4000",
		HandshakeFunc: p2p.NopHandshakeFunc,
		Decoder:       p2p.GoDecoder{},
	}

	tr := NewTCPTransport(opt)
	assert.Equal(t, opt.ListenAddr, tr.ListenAddr)

	// check accept return no error
	assert.Nil(t, tr.ListenAndAccept())
}
