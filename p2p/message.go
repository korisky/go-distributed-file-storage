package p2p

const (
	INCOMING_MESSAGE = 0x1
	INCOMING_STREAM  = 0x2
)

// RPC holds data over transport
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
