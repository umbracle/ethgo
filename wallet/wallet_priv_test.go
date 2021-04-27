package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWallet_Priv(t *testing.T) {
	key, err := GenerateKey()
	assert.NoError(t, err)

	raw, err := key.MarshallPrivateKey()
	assert.NoError(t, err)

	key1, err := NewWalletFromPrivKey(raw)
	assert.NoError(t, err)

	assert.Equal(t, key.addr, key1.addr)
}
