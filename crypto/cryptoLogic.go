package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func NewAesKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	// AES encryption prepare
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	iv := make([]byte, cipherBlock.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// prepend the IV to the file
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	// buffer
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(cipherBlock, iv)
		nw     = cipherBlock.BlockSize()
	)

	// looping
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn // add extra bytes
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, err
}

func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	// AES encryption prepare
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	//if _, err := io.ReadFull(rand.Reader, iv); err != nil {
	//	return 0, err
	//}

	// read IV from the file
	iv := make([]byte, cipherBlock.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	// buffer
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(cipherBlock, iv)
		nw     = cipherBlock.BlockSize()
	)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn // add extra bytes
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}
