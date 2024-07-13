package tcp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_TCPTransport(t *testing.T) {

	// check address
	listenAddr := ":4000"
	tr := NewTCPTransport(listenAddr)
	assert.Equal(t, tr.listenAddress, listenAddr)

	// check accept return no error
	assert.Nil(t, tr.ListenAndAccept())

	// block
	select {}
}
