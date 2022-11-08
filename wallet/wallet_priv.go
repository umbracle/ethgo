package wallet

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec/v2"
)

func ParsePrivateKey(buf []byte) (*ecdsa.PrivateKey, error) {
	prv, _ := btcec.PrivKeyFromBytes(buf)
	return prv.ToECDSA(), nil
}

func NewWalletFromPrivKey(p []byte) (*Key, error) {
	priv, err := ParsePrivateKey(p)
	if err != nil {
		return nil, err
	}
	return NewKey(priv), nil
}
