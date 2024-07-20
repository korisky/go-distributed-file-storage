package server

import (
	"fmt"
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
	// for quit
	quitCh chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storageOpts := storage.StorageOpt{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          storage.NewStore(storageOpts),
		quitCh:         make(chan struct{}),
	}
}

// Start the server
func (s *FileServer) Start() error {
	// port listening
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	// looping accept msg
	s.loop()

	return nil
}

// Stop will use to close a channel
func (s *FileServer) Stop() {
	close(s.quitCh)
}

// loop is for continuing retrieve msg from Transport channel
func (s *FileServer) loop() {
	for {
		select {
		// retrieve msg from read only channel
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		// server stop
		case <-s.quitCh:
			return
		}
	}
}
