package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decoder(io.Reader, *Message) error
}

type GoDecoder struct{}

func (d GoDecoder) Decoder(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (d DefaultDecoder) Decoder(r io.Reader, msg *Message) error {
	// 将Message内容读取到Decoder中
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	// copy from buffer
	msg.Payload = buf[:n]
	return nil
}
