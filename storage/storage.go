package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/roylic/go-distributed-file-storage/crypto"
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
	Root string
	// ID of the owner of the storage,
	// used for all files in that loc
	ID                string
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
	if len(opts.ID) == 0 {
		opts.ID = crypto.GenerateID()
	}
	return &Storage{
		StorageOpt: opts,
	}
}

// Write 添加一个Write允许外部访问
func (s *Storage) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

// WriteDecrypt encKey:AES-Key, key: fileKey, r: io.Reader
func (s *Storage) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {
	// 打开文件
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	// 写入文件 (连接时由于每次传入的是Stream, 没有EOF, 会导致Blocking)
	// 可以使用 CopyN 指定拷贝大小 / 使用limitReader
	// copy with decrypt
	n, err := crypto.CopyDecrypt(encKey, r, f)
	return int64(n), err
}

// writeStream 从reader写入文件
func (s *Storage) writeStream(key string, r io.Reader) (int64, error) {
	// 打开文件
	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	// 写入文件 (连接时由于每次传入的是Stream, 没有EOF, 会导致Blocking)
	// 可以使用 CopyN 指定拷贝大小 / 使用limitReader
	return io.Copy(f, r)
}

// openFileForWriting 打开文件 (附带路径转换)
func (s *Storage) openFileForWriting(key string) (*os.File, error) {
	// 转换路径 + 创建路径 (path = root/id/path)
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	// 创建文件 (由于pkg是在storage, 创建的也会在此之下)
	fullPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.FullPath())
	return os.Create(fullPathNameWithRoot)
}

// Read 从文件读取
func (s *Storage) Read(key string) (int64, io.Reader, error) {
	// 不需要额外buffer, 直接返回文件流 (disk->network)
	return s.readStream(key)
}

// readStream 读取字节流, 注意返回的应该使用ReadCloser可以关闭
func (s *Storage) readStream(key string) (int64, io.ReadCloser, error) {
	// 转换路径
	pathKey := s.PathTransformFunc(key)
	fullPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.FullPath())
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

// Has 判断是否存在
func (s *Storage) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.PathName)
	_, err := os.Stat(pathNameWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

// Delete 删除文件
func (s *Storage) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.PathName)
	defer func() {
		log.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()
	// TODO 暂时不做递归删除无用文件夹, 避免hash碰撞导致删除另外文件
	return os.RemoveAll(pathNameWithRoot)
}
