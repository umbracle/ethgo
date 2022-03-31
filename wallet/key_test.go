package wallet

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/testsuite"
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

func TestSpec_Accounts(t *testing.T) {
	var walletSpec []struct {
		Address    string  `json:"address"`
		Checksum   string  `json:"checksumAddress"`
		Name       string  `json:"name"`
		PrivateKey *string `json:"privateKey,omitempty"`
	}
	testsuite.ReadTestCase(t, "accounts", &walletSpec)

	for _, spec := range walletSpec {
		if spec.PrivateKey != nil {
			// test that we can decode the private key
			priv, err := hex.DecodeString(strings.TrimPrefix(*spec.PrivateKey, "0x"))
			assert.NoError(t, err)

			key, err := NewWalletFromPrivKey(priv)
			assert.NoError(t, err)

			assert.Equal(t, key.Address().String(), spec.Checksum)
		}

		// test that an string address can be checksumed
		addr := ethgo.HexToAddress(spec.Address)
		assert.Equal(t, addr.String(), spec.Checksum)
	}
}
