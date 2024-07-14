package storage

import (
	"io"
	"log"
	"os"
)

// PathTransformFunc 路径转换
type PathTransformFunc func(string) string

type StorageOpt struct {
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type Storage struct {
	StorageOpt
}

func NewStore(opts StorageOpt) *Storage {
	return &Storage{
		StorageOpt: opts,
	}
}

func (s *Storage) writeStream(key string, r io.Reader) error {
	// 转换路径 + 创建路径
	pathName := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	filename := "somefilename"
	fullFilename := pathName + "/" + filename
	f, err := os.Create(fullFilename)
	if err != nil {
		return err
	}

	// 数据流写入文件
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("writtern %d bytes to disk: %s",
		n, fullFilename)
	return nil
}
