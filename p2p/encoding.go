package p2p

import "io"

type Decoder interface {
	Decoder(io.Reader, any) error
}

type GoDecoder struct{}
