package jsonrpc

import (
	"fmt"
	"math/big"

	"github.com/umbracle/go-web3"
)

type Eth struct {
	c *Client
}

func (c *Client) Eth() *Eth {
	return c.endpoints.e
}

// Accounts returns a list of addresses owned by client.
func (e *Eth) Accounts() ([]string, error) {
	var out []string
	if err := e.c.Call("eth_accounts", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BlockNumber returns the number of most recent block.
func (e *Eth) BlockNumber() (uint64, error) {
	var out string
	if err := e.c.Call("eth_blockNumber", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// GetBlockByNumber returns information about a block by block number.
func (e *Eth) GetBlockByNumber(i uint64, full bool) (*web3.Block, error) {
	var b *web3.Block
	if err := e.c.Call("eth_getBlockByNumber", &b, encodeUintToHex(i), full); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBlockByHash returns information about a block by hash.
func (e *Eth) GetBlockByHash(hash string, full bool) (*web3.Block, error) {
	var b *web3.Block
	if err := e.c.Call("eth_getBlockByHash", &b, hash, full); err != nil {
		return nil, err
	}
	return b, nil
}

// SendTransaction creates new message call transaction or a contract creation.
func (e *Eth) SendTransaction(txn *web3.Transaction) (string, error) {
	var hash string
	err := e.c.Call("eth_sendTransaction", &hash, txn)
	return hash, err
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
func (e *Eth) GetTransactionReceipt(hash string) (*web3.Receipt, error) {
	var receipt *web3.Receipt
	err := e.c.Call("eth_getTransactionReceipt", &receipt, hash)
	return receipt, err
}

// GetBalance returns the balance of the account of given address.
func (e *Eth) GetBalance(addr string, blockNumber BlockNumber) (*big.Int, error) {
	var out string
	if err := e.c.Call("eth_getBalance", &out, addr, blockNumber.String()); err != nil {
		return nil, err
	}
	b, ok := new(big.Int).SetString(out[2:], 16)
	if !ok {
		return nil, fmt.Errorf("failed to convert to big.int")
	}
	return b, nil
}

// GasPrice returns the current price per gas in wei.
func (e *Eth) GasPrice() (uint64, error) {
	var out string
	if err := e.c.Call("eth_gasPrice", &out); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}
