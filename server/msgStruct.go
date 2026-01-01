package server

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	ID   string // owner's identifier for finding the file
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string // owner's identifier for finding the file
	Key string
}
