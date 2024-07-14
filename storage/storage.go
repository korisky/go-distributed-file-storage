package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// PathTransformFunc 路径转换
type PathTransformFunc func(string) PathKey

// CASPathTransformFunc 对key进行hash获取分级路径
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 8
	sliceLen := len(hashStr) / blockSize
	path := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		path[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(path, "/"),
		FileName: hashStr,
	}
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

// PathKey 保存转换与原始的pathname
type PathKey struct {
	PathName string
	FileName string
}

// fileFullPath 获取文件全路径名
func (k PathKey) fullPath() string {
	return fmt.Sprintf("%s/%s", k.PathName, k.FileName)
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

// Read 串联流程:
// 1) 根据key调用转换路径, 2) 打开文件流, 3)通过buffer将文件流读出
func (s *Storage) Read(key string) (io.Reader, error) {
	// 读取
	stream, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	// 延迟关闭字节流
	defer stream.Close()
	// 将输入流写入buffer
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, stream)
	return buf, err
}

// readStream 读取字节流, 注意返回的应该使用ReadCloser可以关闭
func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	// 转换路径
	pathKey := s.PathTransformFunc(key)
	// 打开文件
	return os.Open(pathKey.fullPath())
}

// writeStream 从reader写入文件
func (s *Storage) writeStream(key string, r io.Reader) error {
	// 转换路径 + 创建路径
	pathKey := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}
	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	fullPath := pathKey.fullPath()
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	// 写入文件
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("writtern %d bytes to disk: %s", n, fullPath)
	return nil
}
