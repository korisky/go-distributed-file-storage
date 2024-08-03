package p2p

import (
	"encoding/gob"
	"io"
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
	// 将Message内容读取到Decoder中
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	// copy from buffer
	rpc.Payload = buf[:n]

	return nil
}
