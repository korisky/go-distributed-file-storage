package storage

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "SydneyHoliday"
	pathKey := CASPathTransformFunc(key)
	fmt.Println(pathKey.FileName)
	fmt.Println(pathKey.FullPath())
}

func TestStorage(t *testing.T) {
	// init
	key := "SydneyHoliday"
	store, data := storeFile(key, t)
	// read file
	r, err := store.Read(key)
	if err != nil {
		t.Error(err)
	}
	// 读取并使用String比对
	b, _ := io.ReadAll(r)
	fmt.Println(string(b))
	if string(b) != string(data) {
		t.Errorf("Want %s but got %s", string(data), string(b))
	}
}

func TestStorage_Delete(t *testing.T) {
	key := "SydneyHoliday"
	store, _ := storeFile(key, t)
	// check exist
	if exist := store.Has(key); !exist {
		t.Errorf("Does not exist file %s\n", key)
		return
	}
	// delete
	err := store.Delete(key)
	if err != nil {
		t.Error(err)
	}
	// check exist
	if exist := store.Has(key); exist {
		t.Errorf("Does not delete the file %s\n", key)
		return
	}
}

// storeFile 测试使用, 保存文件
func storeFile(key string, t *testing.T) (*Storage, []byte) {
	opts := StorageOpt{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(opts)
	data := []byte("some jpg file byes")
	if _, err := store.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	return store, data
}
