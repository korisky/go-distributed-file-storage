package storage

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "SydneyPic"
	fmt.Println(CASPathTransformFunc(key))
}

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
