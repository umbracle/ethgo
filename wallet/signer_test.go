package wallet

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/panyanyany/go-web3"
)

func TestSigner_EIP1155(t *testing.T) {
	signer1 := NewEIP155Signer(1337)

	addr0 := web3.Address{0x1}
	key, err := GenerateKey()
	assert.NoError(t, err)

	txn := &web3.Transaction{
		To:       &addr0,
		Value:    big.NewInt(10),
		GasPrice: 0,
	}
	txn, err = signer1.SignTx(txn, key)
	assert.NoError(t, err)

	from, err := signer1.RecoverSender(txn)
	assert.NoError(t, err)
	assert.Equal(t, from, key.addr)

	/*
		// try to use a signer with another chain id
		signer2 := NewEIP155Signer(2)
		from2, err := signer2.RecoverSender(txn)
		assert.NoError(t, err)
		assert.NotEqual(t, from, from2)
	*/
}
