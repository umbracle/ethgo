package testcases

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) {
	var walletSpec []struct {
		Address    string  `json:"address"`
		Checksum   string  `json:"checksumAddress"`
		Name       string  `json:"name"`
		PrivateKey *string `json:"privateKey,omitempty"`
	}
	ReadTestCase(t, "accounts", &walletSpec)

	msg := []byte("msg")

	for _, spec := range walletSpec {
		// test that an string address can be checksumed
		addr := ethgo.HexToAddress(spec.Address)
		assert.Equal(t, addr.String(), spec.Checksum)

		if spec.PrivateKey != nil {
			// test that we can decode the private key
			priv, err := hex.DecodeString(strings.TrimPrefix(*spec.PrivateKey, "0x"))
			assert.NoError(t, err)

			key, err := wallet.NewWalletFromPrivKey(priv)
			assert.NoError(t, err)

			assert.Equal(t, key.Address().String(), spec.Checksum)

			// test that we can sign and recover address
			sig, err := key.SignMsg(msg)
			require.NoError(t, err)

			recoveredAddr, err := wallet.EcrecoverMsg(msg, sig)
			require.NoError(t, err)
			require.Equal(t, recoveredAddr, addr)
		}
	}
}
