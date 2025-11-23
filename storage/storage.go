package storage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const DefaultRoot = "../NetworkFiles/"

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
	// Root 是保存的根路径
	Root              string
	PathTransformFunc PathTransformFunc
}

type Storage struct {
	StorageOpt
}

func NewStore(opts StorageOpt) *Storage {
	if nil == opts.PathTransformFunc {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DefaultRoot
	}
	return &Storage{
		StorageOpt: opts,
	}
}

// Write 添加一个Write允许外部访问
func (s *Storage) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

// Read 从文件读取的流程:
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

// Has 判断是否存在
func (s *Storage) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	_, err := os.Stat(s.Root + pathKey.fullPath())
	return !errors.Is(err, os.ErrNotExist)
}

// Delete 删除文件
func (s *Storage) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()
	// TODO 暂时不做递归删除无用文件夹, 避免hash碰撞导致删除另外文件
	return os.RemoveAll(s.Root + pathKey.fullPath())
}

// readStream 读取字节流, 注意返回的应该使用ReadCloser可以关闭
func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	// 转换路径
	pathKey := s.PathTransformFunc(key)
	// 打开文件
	return os.Open(s.Root + pathKey.fullPath())
}

// writeStream 从reader写入文件
func (s *Storage) writeStream(key string, r io.Reader) (int64, error) {
	// 转换路径 + 创建路径
	pathKey := s.PathTransformFunc(key)
	if err := os.MkdirAll(s.Root+pathKey.PathName, os.ModePerm); err != nil {
		return 0, err
	}
	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	fullPath := s.Root + pathKey.fullPath()
	f, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	// 写入文件 (连接时由于每次传入的是Stream, 没有EOF, 会导致Blocking)
	// 可以使用 CopyN 指定拷贝大小 / 使用limitReader
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	log.Printf("writtern %d bytes to disk: %s\n", n, fullPath)
	return n, nil
}
