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
	key := newAesKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("cypher in str %s\n", dst.String())

	// decryption
	out := new(bytes.Buffer)
	if _, err := copyDecrypt(key, dst, out); err != nil {
		t.Error(err)
	}
	fmt.Printf("plain text in str %s\n", out.String())

	// comparison
	if payload != out.String() {
		t.Error("decryption failed")
	}

}
