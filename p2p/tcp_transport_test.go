package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_TCPTransport(t *testing.T) {

	// check address

	opt := TCPTransportOpt{
		ListenAddr:    ":4000",
		HandshakeFunc: NopHandshakeFunc,
		Decoder:       GoDecoder{},
	}

	tr := NewTCPTransport(opt)
	assert.Equal(t, opt.ListenAddr, tr.ListenAddr)

	// check accept return no error
	assert.Nil(t, tr.ListenAndAccept())
}
