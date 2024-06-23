package tcp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_TCPTransport(t *testing.T) {
	listenAddr := ":4000"
	tr := NewTCPTransport(listenAddr)

	assert.Equal(t, tr.listenAddress, listenAddr)
}
