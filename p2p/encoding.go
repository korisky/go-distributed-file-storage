package p2p

import (
	"encoding/gob"
	"io"
	"log"
)

type Decoder interface {
	Decoder(io.Reader, *RPC) error
}

type GoDecoder struct{}

func (d GoDecoder) Decoder(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(rpc)
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decoder(r io.Reader, rpc *RPC) error {

	// determine the type by 1st byte
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		log.Fatalf("Decoding error: %+v, returning nil", err)
		return nil
	}

	// if it's stream -> finished
	stream := peekBuf[0] == INCOMING_STREAM
	if stream {
		rpc.Stream = true
		return nil
	}

	// if it's message -> read and copy to buffer (as payload)
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	// copy from buffer
	rpc.Payload = buf[:n]
	return nil
}
