package keystore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV3_EncodeDecode(t *testing.T) {
	data := []byte{0x1, 0x2}
	password := "abcd"

	encrypted, err := EncryptV3(data, password)
	assert.NoError(t, err)

	found, err := DecryptV3(encrypted, password)
	assert.NoError(t, err)

	assert.Equal(t, data, found)
}
