package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/btcsuite/btcd/btcec/v2"
	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// S256 is the secp256k1 elliptic curve
var S256 = btcec.S256()

var _ ethgo.Key = &Key{}

// Key is an implementation of the Key interface with a private key
type Key struct {
	priv *btcec.PrivateKey
	pub  *btcec.PublicKey
	addr ethgo.Address
}

func (k *Key) Address() ethgo.Address {
	return k.addr
}

func (k *Key) MarshallPrivateKey() ([]byte, error) {
	return (*btcec.PrivateKey)(k.priv).Serialize(), nil
}

func (k *Key) SignMsg(msg []byte) ([]byte, error) {
	return k.Sign(ethgo.Keccak256(msg))
}

func (k *Key) Sign(hash []byte) ([]byte, error) {
	sig, err := btcecdsa.SignCompact(k.priv, hash, false)
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
func NewKey(prv *ecdsa.PrivateKey) (*Key, error) {
	var priv btcec.PrivateKey
	if overflow := priv.Key.SetByteSlice(prv.D.Bytes()); overflow || priv.Key.IsZero() {
		return nil, fmt.Errorf("invalid key: overflow")
	}

	k := &Key{
		priv: &priv,
		pub:  priv.PubKey(),
		addr: pubKeyToAddress(priv.PubKey().ToECDSA()),
	}
	return k, nil
}

func pubKeyToAddress(pub *ecdsa.PublicKey) (addr ethgo.Address) {
	b := ethgo.Keccak256(elliptic.Marshal(S256, pub.X, pub.Y)[1:])
	copy(addr[:], b[12:])
	return
}

// GenerateKey generates a new key based on the secp256k1 elliptic curve.
func GenerateKey() (*Key, error) {
	priv, err := ecdsa.GenerateKey(S256, rand.Reader)
	if err != nil {
		return nil, err
	}
	return NewKey(priv)
}

func EcrecoverMsg(msg, signature []byte) (ethgo.Address, error) {
	return Ecrecover(ethgo.Keccak256(msg), signature)
}

func Ecrecover(hash, signature []byte) (ethgo.Address, error) {
	pub, err := RecoverPubkey(signature, hash)
	if err != nil {
		return ethgo.Address{}, err
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
	pub, _, err := btcecdsa.RecoverCompact(sig, hash)
	if err != nil {
		return nil, err
	}
	return pub.ToECDSA(), nil
}
