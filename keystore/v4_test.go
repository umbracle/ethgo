package keystore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV4_EncodeDecode(t *testing.T) {
	data := []byte{0x1, 0x2}
	password := "abcd"

	encrypted, err := EncryptV4(data, password)
	assert.NoError(t, err)

	found, err := DecryptV4(encrypted, password)
	assert.NoError(t, err)

	assert.Equal(t, data, found)
}

func TestV4_NormalizePassword(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			"𝔱𝔢𝔰𝔱𝔭𝔞𝔰𝔰𝔴𝔬𝔯𝔡🔑",
			"testpassword🔑",
		},
		{
			string([]byte{
				0x00, 0x1F,
				0x70, 0x61, 0x73, 0x73, 0x77, 0x6F, 0x72, 0x64,
			}),
			"password",
		},
	}

	for _, c := range cases {
		found := normalizePassword(c.input)
		assert.Equal(t, c.output, found)
	}
}
