package jsonrpc

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/testutil"
)

var (
	addr0 = "0x0000000000000000000000000000000000000000"
	addr1 = "0x0000000000000000000000000000000000000001"
)

func TestEthAccounts(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	_, err := c.Eth().Accounts()
	assert.NoError(t, err)
}

func TestEthBlockNumber(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	for i := uint64(0); i < 10; i++ {
		num, err := c.Eth().BlockNumber()
		assert.NoError(t, err)
		assert.Equal(t, num, i)
		assert.NoError(t, s.ProcessBlock())
	}
}

func TestEthGetBalance(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())

	before, err := c.Eth().GetBalance(s.Account(0), web3.Latest)
	assert.NoError(t, err)

	amount := big.NewInt(10)
	txn := &web3.Transaction{
		From:  s.Account(0),
		To:    "0x015f68893a39b3ba0681584387670ff8b00f4db2",
		Value: amount,
	}
	_, err = s.SendTxn(txn)
	assert.NoError(t, err)

	after, err := c.Eth().GetBalance(s.Account(0), web3.Latest)
	assert.NoError(t, err)

	// the balance in 'after' must be 'before' - 'amount'
	assert.Equal(t, after.Add(after, amount).Cmp(before), 0)

	// get balance at block 0
	before2, err := c.Eth().GetBalance(s.Account(0), 0)
	assert.NoError(t, err)
	assert.Equal(t, before, before2)
}

func TestEthGetBlockByNumber(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())

	block, err := c.Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// block 1 has not been processed yet, do not fail but returns nil
	block, err = c.Eth().GetBlockByNumber(1, true)
	assert.NoError(t, err)
	assert.Nil(t, block)

	// process a new block
	assert.NoError(t, s.ProcessBlock())

	// there exists a block 1 now
	block, err = c.Eth().GetBlockByNumber(1, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(1))
}

func TestEthGetBlockByHash(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())

	// get block 0 first by number
	block, err := c.Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// get block 0 by hash
	block2, err := c.Eth().GetBlockByHash(block.Hash, true)
	assert.NoError(t, err)
	assert.Equal(t, block, block2)
}

func TestEthGasPrice(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	_, err := c.Eth().GasPrice()
	assert.NoError(t, err)
}

func TestEthSendTransaction(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())

	txn := &web3.Transaction{
		From:     s.Account(0),
		GasPrice: testutil.DefaultGasPrice,
		Gas:      testutil.DefaultGasLimit,
		To:       "0x015f68893a39b3ba0681584387670ff8b00f4db2",
		Value:    big.NewInt(10),
	}
	hash, err := c.Eth().SendTransaction(txn)
	assert.NoError(t, err)

	var receipt *web3.Receipt
	for {
		receipt, err = c.Eth().GetTransactionReceipt(hash)
		if err != nil {
			t.Fatal(err)
		}
		if receipt != nil {
			break
		}
	}
}

func TestEthGetLogs(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	fmt.Println(c)

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").
		Add("address", true).
		Add("address", true))

	cc.EmitEvent("setA1", "A", addr0, addr1)
	cc.EmitEvent("setA2", "A", addr1, addr0)

	addr := s.DeployContract(cc)

	r := s.TxnTo(addr, "setA2")

	filter := &web3.LogFilter{
		BlockHash: r.BlockHash,
	}
	logs, err := c.Eth().GetLogs(filter)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	log := logs[0]
	assert.Len(t, log.Topics, 3)
	assert.Equal(t, log.Address, addr)

	// first topic is the signature of the event
	assert.Equal(t, log.Topics[0], cc.GetEvent("A").Sig())

	// topics have 32 bytes and the addr are 20 bytes, then, assert.Equal wont work.
	// this is a workaround until we build some helper function to test this better
	assert.True(t, strings.HasSuffix(log.Topics[1], addr1[2:]))
	assert.True(t, strings.HasSuffix(log.Topics[2], addr0[2:]))
}
