package storage

import (
	"bytes"
	"log"
	"testing"
)

func TestStorage(t *testing.T) {
	opts := StorageOpt{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	store := NewStore(opts)

	data := bytes.NewReader([]byte("some jpg file byes"))
	err := store.writeStream("File1", data)
	if err != nil {
		log.Fatal(err)
	}
}
