package crypto

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {

	payload := "please read enc msg"

	// encryption
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := NewAesKey()
	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("cypher in str %s\n", dst.String())
	fmt.Println(len(payload))
	fmt.Println(len(dst.String()))

	// decryption
	out := new(bytes.Buffer)
	nw, err := CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("plain text in str %s\n", out.String())
	fmt.Printf("NW = 16 + len(payload), result:%t", nw == len(payload)+16)

	// comparison
	if payload != out.String() {
		t.Error("decryption failed")
	}

}
