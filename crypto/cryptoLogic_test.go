package crypto

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {

	// encryption
	src := bytes.NewReader([]byte("please read enc msg"))
	dst := new(bytes.Buffer)
	key := newAesKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("cypher in str %s\n", string(dst.Bytes()))

	// decryption
	out := new(bytes.Buffer)
	if _, err := copyDecrypt(key, dst, out); err != nil {
		t.Error(err)
	}
	fmt.Printf("plain text in str %s\n", string(out.Bytes()))

}
