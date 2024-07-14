package storage

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

// PathTransformFunc 路径转换
type PathTransformFunc func(string) string

// CASPathTransformFunc 对key进行hash获取分级路径
func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 8
	sliceLen := len(hashStr) / blockSize
	path := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		path[i] = hashStr[from:to]
	}
	return strings.Join(path, "/")
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

// StorageOpt 存储Opt
type StorageOpt struct {
	PathTransformFunc PathTransformFunc
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

	// 将文件流写入buffer
	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	// 对文件名进行hash
	filenameByes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameByes[:])

	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	fullFilename := pathName + "/" + filename
	f, err := os.Create(fullFilename)
	if err != nil {
		return err
	}

	// buffer写入文件
	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}
	log.Printf("writtern %d bytes to disk: %s",
		n, fullFilename)
	return nil
}
