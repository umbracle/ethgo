package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWallet_Mnemonic(t *testing.T) {
	_, err := NewWalletFromMnemonic("sound practice disease erupt basket pumpkin truck file gorilla behave find exchange napkin boy congress address city net prosper crop chair marine chase seven")
	assert.NoError(t, err)
}

func TestWallet_MnemonicDerivationPath(t *testing.T) {
	cases := []struct {
		path       string
		derivation DerivationPath
	}{
		{"m/44'/60'/0'/0", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}},
		{"m/44'/60'/0'/128", DerivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 128}},
	}

	for _, c := range cases {
		path, err := parseDerivationPath(c.path)
		assert.NoError(t, err)
		assert.Equal(t, *path, c.derivation)
	}
}
