package server

import "github.com/roylic/go-distributed-file-storage/storage"

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
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

func (s *FileServer) start() {

}
