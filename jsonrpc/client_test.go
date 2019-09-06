package jsonrpc

import (
	"math/big"
	"testing"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/testutil"
)

type TestServer struct {
	server   *testutil.TestServer
	client   *Client
	accounts []string
}

func NewTestServer(t *testing.T, cb testutil.ServerConfigCallback) *TestServer {
	s := testutil.NewTestServer(t, cb)
	c := NewClient(s.HttpAddr())

	accounts, err := c.Eth().Accounts()
	if err != nil {
		t.Fatal(err)
	}

	tt := &TestServer{
		server:   s,
		client:   c,
		accounts: accounts,
	}
	return tt
}

func (t *TestServer) ProcessBlock() error {
	_, err := t.SendTxn(&web3.Transaction{
		From:  t.accounts[0],
		To:    "0x015f68893a39b3ba0681584387670ff8b00f4db2",
		Value: big.NewInt(10),
	})
	return err
}

func (t *TestServer) Account(i int) string {
	return t.accounts[i]
}

const (
	defaultGasPrice = 1879048192
	defaultGasLimit = 5242880
)

func (t *TestServer) SendTxn(txn *web3.Transaction) (*web3.Receipt, error) {
	if txn.GasPrice == 0 {
		txn.GasPrice = defaultGasPrice
	}
	if txn.Gas == 0 {
		txn.Gas = defaultGasLimit
	}

	hash, err := t.client.Eth().SendTransaction(txn)
	if err != nil {
		return nil, err
	}
	var receipt *web3.Receipt
	for {
		receipt, err = t.client.Eth().GetTransactionReceipt(hash)
		if err != nil {
			return nil, err
		}
		if receipt != nil {
			break
		}
	}
	return receipt, nil
}

func (t *TestServer) Client() *Client {
	return t.client
}

func (t *TestServer) Close() {
	t.server.Stop()
}
