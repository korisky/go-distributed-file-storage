package storage

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "SydneyPic"
	pathKey := CASPathTransformFunc(key)
	fmt.Println(pathKey.FileName)
	fmt.Println(pathKey.fileFullPath())
}

func TestStorage(t *testing.T) {
	opts := StorageOpt{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)

	data := bytes.NewReader([]byte("some jpg file byes"))
	err := store.writeStream("SydneyPic", data)
	if err != nil {
		log.Fatal(err)
	}
}
