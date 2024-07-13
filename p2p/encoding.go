package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decoder(io.Reader, any) error
}

type GoDecoder struct{}

func (dec GoDecoder) Decoder(r io.Reader, v any) error {
	return gob.NewDecoder(r).Decode(v)
}
