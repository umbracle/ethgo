package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/btcsuite/btcd/btcec"
	web3 "github.com/umbracle/ethgo"
)

// S256 is the secp256k1 elliptic curve
var S256 = btcec.S256()

var _ web3.Key = &Key{}

// Key is an implementation of the Key interface with a private key
type Key struct {
	priv *ecdsa.PrivateKey
	pub  *ecdsa.PublicKey
	addr web3.Address
}

func (k *Key) Address() web3.Address {
	return k.addr
}

func (k *Key) MarshallPrivateKey() ([]byte, error) {
	return (*btcec.PrivateKey)(k.priv).Serialize(), nil
}

func (k *Key) SignMsg(msg []byte) ([]byte, error) {
	return k.Sign(web3.Keccak256(msg))
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	sig, err := btcec.SignCompact(S256, (*btcec.PrivateKey)(k.priv), hash, false)
	if err != nil {
		return nil, err
	}
	term := byte(0)
	if sig[0] == 28 {
		term = 1
	}
	return append(sig, term)[1:], nil
}

// NewKey creates a new key with a private key
func NewKey(priv *ecdsa.PrivateKey) *Key {
	return &Key{
		priv: priv,
		pub:  &priv.PublicKey,
		addr: pubKeyToAddress(&priv.PublicKey),
	}
}

func pubKeyToAddress(pub *ecdsa.PublicKey) (addr web3.Address) {
	b := web3.Keccak256(elliptic.Marshal(S256, pub.X, pub.Y)[1:])
	copy(addr[:], b[12:])
	return
}

// GenerateKey generates a new key based on the secp256k1 elliptic curve.
func GenerateKey() (*Key, error) {
	priv, err := ecdsa.GenerateKey(S256, rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewKey(priv), nil
}

func EcrecoverMsg(msg, signature []byte) (web3.Address, error) {
	return Ecrecover(web3.Keccak256(msg), signature)
}

func Ecrecover(hash, signature []byte) (web3.Address, error) {
	pub, err := RecoverPubkey(signature, hash)
	if err != nil {
		return web3.Address{}, err
	}
	return pubKeyToAddress(pub), nil
}

func RecoverPubkey(signature, hash []byte) (*ecdsa.PublicKey, error) {
	size := len(signature)
	term := byte(27)
	if signature[size-1] == 1 {
		term = 28
	}

	sig := append([]byte{term}, signature[:size-1]...)
	pub, _, err := btcec.RecoverCompact(S256, sig, hash)
	if err != nil {
		return nil, err
	}
	return pub.ToECDSA(), nil
}
