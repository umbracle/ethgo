package wallet

import (
	"math/big"

	"github.com/cloudwalk/ethgo"
	"github.com/umbracle/fastrlp"
)

type Signer interface {
	// RecoverSender returns the sender to the transaction
	RecoverSender(tx *ethgo.Transaction) (ethgo.Address, error)

	// SignTx signs a transaction
	SignTx(tx *ethgo.Transaction, key ethgo.Key) (*ethgo.Transaction, error)
}

type EIP1155Signer struct {
	chainID uint64
}

func NewEIP155Signer(chainID uint64) *EIP1155Signer {
	return &EIP1155Signer{chainID: chainID}
}

func (e *EIP1155Signer) RecoverSender(tx *ethgo.Transaction) (ethgo.Address, error) {
	v := new(big.Int).SetBytes(tx.V).Uint64()
	v -= e.chainID * 2
	v -= 8
	v -= 27

	sig, err := encodeSignature(tx.R, tx.S, byte(v))
	if err != nil {
		return ethgo.Address{}, err
	}
	addr, err := Ecrecover(signHash(tx, e.chainID), sig)
	if err != nil {
		return ethgo.Address{}, err
	}
	return addr, nil
}

func trimBytesZeros(b []byte) []byte {
	var i int
	for i = 0; i < len(b); i++ {
		if b[i] != 0x0 {
			break
		}
	}
	return b[i:]
}

func (e *EIP1155Signer) SignTx(tx *ethgo.Transaction, key ethgo.Key) (*ethgo.Transaction, error) {
	hash := signHash(tx, e.chainID)

	sig, err := key.Sign(hash)
	if err != nil {
		return nil, err
	}

	vv := uint64(sig[64])
	if tx.Type == 0 {
		vv = vv + 35 + e.chainID*2
	}

	tx.R = trimBytesZeros(sig[:32])
	tx.S = trimBytesZeros(sig[32:64])
	tx.V = new(big.Int).SetUint64(vv).Bytes()
	return tx, nil
}

func signHash(tx *ethgo.Transaction, chainID uint64) []byte {
	a := fastrlp.DefaultArenaPool.Get()

	v := a.NewArray()

	if tx.Type != 0 {
		// either dynamic and access type
		v.Set(a.NewBigInt(tx.ChainID))
	}

	v.Set(a.NewUint(tx.Nonce))

	if tx.Type == ethgo.TransactionDynamicFee {
		// dynamic fee uses
		v.Set(a.NewBigInt(tx.MaxPriorityFeePerGas))
		v.Set(a.NewBigInt(tx.MaxFeePerGas))
	} else {
		// legacy and access type use gas price
		v.Set(a.NewUint(tx.GasPrice))
	}

	v.Set(a.NewUint(tx.Gas))
	if tx.To == nil {
		v.Set(a.NewNull())
	} else {
		v.Set(a.NewCopyBytes((*tx.To)[:]))
	}
	v.Set(a.NewBigInt(tx.Value))
	v.Set(a.NewCopyBytes(tx.Input))

	if tx.Type != 0 {
		// either dynamic and access type
		accessList, err := tx.AccessList.MarshalRLPWith(a)
		if err != nil {
			panic(err)
		}
		v.Set(accessList)
	}

	// EIP155
	if chainID != 0 && tx.Type == 0 {
		v.Set(a.NewUint(chainID))
		v.Set(a.NewUint(0))
		v.Set(a.NewUint(0))
	}

	dst := v.MarshalTo(nil)

	// append the tx type byte
	if tx.Type == ethgo.TransactionAccessList {
		dst = append([]byte{0x1}, dst...)
	} else if tx.Type == ethgo.TransactionDynamicFee {
		dst = append([]byte{0x2}, dst...)
	}

	hash := ethgo.Keccak256(dst)
	fastrlp.DefaultArenaPool.Put(a)
	return hash
}

func encodeSignature(R, S []byte, V byte) ([]byte, error) {
	sig := make([]byte, 65)
	copy(sig[32-len(R):32], R)
	copy(sig[64-len(S):64], S)
	sig[64] = V
	return sig, nil
}
