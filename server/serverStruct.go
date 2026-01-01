package server

import (
	"github.com/roylic/go-distributed-file-storage/crypto"
	"github.com/roylic/go-distributed-file-storage/p2p"
	"github.com/roylic/go-distributed-file-storage/storage"
	"sync"
)

// FileServerOpts inner Transport is for accepting the p2p communication
type FileServerOpts struct {
	ID                string // server identifier
	EncKey            []byte
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	Storage *storage.Storage

	errCh  chan error    // 出错时的停止
	quitCh chan struct{} // 退出时的停止
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := storage.StorageOpt{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	if len(opts.ID) == 0 {
		opts.ID = crypto.GenerateID()
	}
	return &FileServer{
		FileServerOpts: opts,
		peers:          make(map[string]p2p.Peer),
		Storage:        storage.NewStore(storageOpts),
		quitCh:         make(chan struct{}),
	}
}
