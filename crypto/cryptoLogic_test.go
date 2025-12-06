package crypto

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	src := bytes.NewReader([]byte("please read enc msg"))
	dst := new(bytes.Buffer)
	key := newAesKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("cypher in str %s ", string(dst.Bytes()))
}
