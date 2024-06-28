package e2e

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/jsonrpc"
	"github.com/Ethernal-Tech/ethgo/testutil"
	"github.com/Ethernal-Tech/ethgo/wallet"
	"github.com/stretchr/testify/assert"
)

func TestSendSignedTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	key, err := wallet.GenerateKey()
	assert.NoError(t, err)

	// add value to the new key
	value := big.NewInt(1000000000000000000)
	s.Transfer(key.Address(), value)

	c, _ := jsonrpc.NewClient(s.HTTPAddr())

	found, _ := c.Eth().GetBalance(key.Address(), ethgo.Latest)
	assert.Equal(t, found, value)

	chainID, err := c.Eth().ChainID()
	assert.NoError(t, err)

	// send a signed transaction
	to := ethgo.Address{0x1}
	transferVal := big.NewInt(1000)

	gasPrice, err := c.Eth().GasPrice()
	assert.NoError(t, err)

	txn := &ethgo.Transaction{
		To:       &to,
		Value:    transferVal,
		Nonce:    0,
		GasPrice: gasPrice,
	}

	{
		msg := &ethgo.CallMsg{
			From:     key.Address(),
			To:       &to,
			Value:    transferVal,
			GasPrice: gasPrice,
		}
		limit, err := c.Eth().EstimateGas(msg)
		assert.NoError(t, err)

		txn.Gas = limit
	}

	signer := wallet.NewEIP155Signer(chainID.Uint64())
	txn, err = signer.SignTx(txn, key)
	assert.NoError(t, err)

	from, err := signer.RecoverSender(txn)
	assert.NoError(t, err)
	assert.Equal(t, from, key.Address())

	data, err := txn.MarshalRLPTo(nil)
	assert.NoError(t, err)

	hash, err := c.Eth().SendRawTransaction(data)
	assert.NoError(t, err)

	_, err = s.WaitForReceipt(hash)
	assert.NoError(t, err)

	balance, err := c.Eth().GetBalance(to, ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, transferVal)
}
