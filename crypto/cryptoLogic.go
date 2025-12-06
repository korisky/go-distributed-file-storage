package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {

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
	)

	// looping
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			if _, err := dst.Write(buf[:n]); err != nil {
				return 0, err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}

	return 0, err
}
