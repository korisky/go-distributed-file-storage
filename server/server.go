package server

import (
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/storage"
)

// FileServerOpts inner Transport is for accepting the p2p communication
type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts
	store *storage.Storage
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := storage.StorageOpt{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          storage.NewStore(storageOpts),
	}
}

func (s *FileServerOpts) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	return nil
}
