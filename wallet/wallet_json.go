package wallet

import (
	"os"

	"github.com/umbracle/ethgo/keystore"
)

func NewJSONWalletFromFile(path string, password string) (*Key, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewJSONWalletFromContent(data, password)
}

func NewJSONWalletFromContent(content []byte, password string) (*Key, error) {
	dst, err := keystore.DecryptV3(content, password)
	if err != nil {
		return nil, err
	}
	key, err := NewWalletFromPrivKey(dst)
	if err != nil {
		return nil, err
	}
	return key, nil
}
