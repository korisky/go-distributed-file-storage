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

// FullPath 获取文件全路径名
func (k PathKey) FullPath() string {
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
// FIXME: instead of copying directly to a reader, first copy to a buffer, then return the file from the readStream
func (s *Storage) Read(key string) (int64, io.Reader, error) {
	// 读取
	n, stream, err := s.readStream(key)
	if err != nil {
		return n, nil, err
	}
	// 延迟关闭字节流
	defer stream.Close()
	// 将输入流写入buffer
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, stream)
	return n, buf, err
}

// Has 判断是否存在
func (s *Storage) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	_, err := os.Stat(pathNameWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

// Delete 删除文件
func (s *Storage) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	defer func() {
		log.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()
	// TODO 暂时不做递归删除无用文件夹, 避免hash碰撞导致删除另外文件
	return os.RemoveAll(pathNameWithRoot)
}

// readStream 读取字节流, 注意返回的应该使用ReadCloser可以关闭
func (s *Storage) readStream(key string) (int64, io.ReadCloser, error) {
	// 转换路径
	pathKey := s.PathTransformFunc(key)
	fullPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	// 打开文件
	fio, err := os.Open(fullPathNameWithRoot)
	if err != nil {
		return 0, nil, err
	}
	// 获取文件信息 (size)
	fi, err := fio.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), fio, nil
}

// writeStream 从reader写入文件
func (s *Storage) writeStream(key string, r io.Reader) (int64, error) {
	// 转换路径 + 创建路径
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}
	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	fullPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	f, err := os.Create(fullPathNameWithRoot)
	if err != nil {
		return 0, err
	}
	// 写入文件 (连接时由于每次传入的是Stream, 没有EOF, 会导致Blocking)
	// 可以使用 CopyN 指定拷贝大小 / 使用limitReader
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	log.Printf("writtern %d bytes to disk: %s\n", n, fullPathNameWithRoot)
	return n, nil
}
