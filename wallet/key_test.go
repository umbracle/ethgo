package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeySign(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)

	msg := []byte("hello world")
	signature, err := key.SignMsg(msg)
	assert.NoError(t, err)

	addr, err := EcrecoverMsg(msg, signature)
	assert.NoError(t, err)
	assert.Equal(t, addr, key.addr)
}
