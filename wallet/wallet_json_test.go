package wallet

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWallet_JSON(t *testing.T) {
	raw, err := os.ReadFile("./fixtures/wallet_json.json")
	assert.NoError(t, err)

	var cases []struct {
		Wallet   json.RawMessage
		Password string
		Address  string
	}
	assert.NoError(t, json.Unmarshal(raw, &cases))

	for _, c := range cases {
		key, err := NewJSONWalletFromContent(c.Wallet, c.Password)
		assert.NoError(t, err)
		assert.Equal(t, key.Address().String(), c.Address)
	}
}
