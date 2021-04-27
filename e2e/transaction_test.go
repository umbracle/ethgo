package e2e

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/testutil"
	"github.com/umbracle/go-web3/wallet"
)

func TestSendSignedTransaction(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	key, err := wallet.GenerateKey()
	assert.NoError(t, err)

	// add value to the new key
	value := big.NewInt(10000000000)
	s.Transfer(key.Address(), value)

	c, _ := jsonrpc.NewClient(s.HTTPAddr())

	found, _ := c.Eth().GetBalance(key.Address(), web3.Latest)
	assert.Equal(t, found, value)

	chainID, err := c.Eth().ChainID()
	assert.NoError(t, err)

	// send a signed transaction
	to := web3.Address{0x1}
	transferVal := big.NewInt(1000)

	txn := &web3.Transaction{
		To:    &to,
		Value: transferVal,
	}

	{
		msg := &web3.CallMsg{
			From:  key.Address(),
			To:    to,
			Value: transferVal,
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

	data := txn.MarshalRLP()
	hash, err := c.Eth().SendRawTransaction(data)
	assert.NoError(t, err)

	_, err = s.WaitForReceipt(hash)
	assert.NoError(t, err)

	balance, err := c.Eth().GetBalance(to, web3.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, transferVal)
}
