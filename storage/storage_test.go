package storage

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "SydneyPic"
	pathKey := CASPathTransformFunc(key)
	fmt.Println(pathKey.FileName)
	fmt.Println(pathKey.fullPath())
}

func TestStorage(t *testing.T) {
	opts := StorageOpt{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)

	key := "SydneyPic"
	data := []byte("some jpg file byes")
	if err := store.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := store.Read(key)
	if err != nil {
		t.Error(err)
	}

	// 读取并使用String比对
	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("Want %s but got %s", string(data), string(b))
	}
}
