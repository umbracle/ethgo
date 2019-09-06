package jsonrpc

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3"
)

func TestEthAccounts(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	_, err := s.Client().Eth().Accounts()
	assert.NoError(t, err)
}

func TestEthBlockNumber(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	for i := uint64(0); i < 10; i++ {
		num, err := s.Client().Eth().BlockNumber()
		assert.NoError(t, err)
		assert.Equal(t, num, i)
		assert.NoError(t, s.ProcessBlock())
	}
}

func TestEthGetBalance(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	before, err := s.Client().Eth().GetBalance(s.Account(0), Latest)
	assert.NoError(t, err)

	amount := big.NewInt(10)
	txn := &web3.Transaction{
		From:  s.Account(0),
		To:    "0x015f68893a39b3ba0681584387670ff8b00f4db2",
		Value: amount,
	}
	_, err = s.SendTxn(txn)
	assert.NoError(t, err)

	after, err := s.Client().Eth().GetBalance(s.Account(0), Latest)
	assert.NoError(t, err)

	// the balance in 'after' must be 'before' - 'amount'
	assert.Equal(t, after.Add(after, amount).Cmp(before), 0)

	// get balance at block 0
	before2, err := s.Client().Eth().GetBalance(s.Account(0), 0)
	assert.NoError(t, err)
	assert.Equal(t, before, before2)
}

func TestEthGetBlockByNumber(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	block, err := s.Client().Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// block 1 has not been processed yet, do not fail but returns nil
	block, err = s.Client().Eth().GetBlockByNumber(1, true)
	assert.NoError(t, err)
	assert.Nil(t, block)

	// process a new block
	assert.NoError(t, s.ProcessBlock())

	// there exists a block 1 now
	block, err = s.Client().Eth().GetBlockByNumber(1, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(1))
}

func TestEthGetBlockByHash(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	// get block 0 first by number
	block, err := s.Client().Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// get block 0 by hash
	block2, err := s.Client().Eth().GetBlockByHash(block.Hash, true)
	assert.NoError(t, err)
	assert.Equal(t, block, block2)
}

func TestEthGasPrice(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	_, err := s.Client().Eth().GasPrice()
	assert.NoError(t, err)
}

func TestEthSendTransaction(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	txn := &web3.Transaction{
		From:     s.Account(0),
		GasPrice: defaultGasPrice,
		Gas:      defaultGasLimit,
		To:       "0x015f68893a39b3ba0681584387670ff8b00f4db2",
		Value:    big.NewInt(10),
	}
	hash, err := s.Client().Eth().SendTransaction(txn)
	assert.NoError(t, err)

	var receipt *web3.Receipt
	for {
		receipt, err = s.Client().Eth().GetTransactionReceipt(hash)
		if err != nil {
			t.Fatal(err)
		}
		if receipt != nil {
			break
		}
	}
}
