package contract

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

type jsonRPCNodeProvider struct {
	client *jsonrpc.Client
}

func (j *jsonRPCNodeProvider) Call(addr ethgo.Address, input []byte, opts *CallOpts) ([]byte, error) {
	msg := &ethgo.CallMsg{
		To:   &addr,
		Data: input,
	}
	if opts.From != ethgo.ZeroAddress {
		msg.From = opts.From
	}
	rawStr, err := j.client.Eth().Call(msg, opts.Block)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (j *jsonRPCNodeProvider) Txn(addr ethgo.Address, input []byte, opts *TxnOpts) (Txn1, error) {
	var err error

	// estimate gas price
	if opts.GasPrice == 0 {
		opts.GasPrice, err = j.client.Eth().GasPrice()
		if err != nil {
			return nil, err
		}
	}
	// estimate gas limit
	if opts.GasLimit == 0 {
		msg := &ethgo.CallMsg{
			From:  opts.From,
			To:    &addr,
			Data:  input,
			Value: opts.Value,
		}
		opts.GasLimit, err = j.client.Eth().EstimateGas(msg)
		if err != nil {
			return nil, err
		}
	}

	// send transaction
	rawTxn := &ethgo.Transaction{
		From:     opts.From,
		Input:    input,
		GasPrice: opts.GasPrice,
		Gas:      opts.GasLimit,
		Value:    opts.Value,
	}
	if addr != ethgo.ZeroAddress {
		rawTxn.To = &addr
	}

	signer := wallet.NewEIP155Signer(1)
	signedTxn, err := signer.SignTx(rawTxn, key)
	if err != nil {
		return nil, err
	}

	txn := &jsonrpcTransaction{
		txn:    rawTxn,
		client: j.client,
	}
	return txn, nil
}

type jsonrpcTransaction struct {
	hash   ethgo.Hash
	client *jsonrpc.Client
	txn    *ethgo.Transaction
}

func (j *jsonrpcTransaction) Hash() ethgo.Hash {
	return j.hash
}

func (j *jsonrpcTransaction) EstimatedGas() uint64 {
	return j.txn.Gas
}

func (j *jsonrpcTransaction) GasPrice() uint64 {
	return j.txn.GasPrice
}

func (j *jsonrpcTransaction) Do() error {
	hash, err := j.client.Eth().SendTransaction(j.txn)
	if err != nil {
		return err
	}
	j.hash = hash
	return nil
}

func (j *jsonrpcTransaction) Wait() (*ethgo.Receipt, error) {
	if (j.hash == ethgo.Hash{}) {
		panic("transaction not executed")
	}

	for {
		receipt, err := j.client.Eth().GetTransactionReceipt(j.hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			return receipt, nil
		}
	}
}

// NodeProvider handles the interactions with the Ethereum 1x node
type NodeProvider interface {
	Call(ethgo.Address, []byte, ethgo.BlockNumber) ([]byte, error)
	Txn(ethgo.Address, []byte) (Txn1, error)
}

// Txn1 is the transaction object returned
type Txn1 interface {
	Hash() ethgo.Hash
	EstimatedGas() uint64
	GasPrice() uint64
	Do() error
	Wait() (*ethgo.Receipt, error)
}

func NewAbiCaller(addr ethgo.Address, abi *abi.ABI, provider NodeProvider) *AbiCaller {
	return &AbiCaller{addr: addr, abi: abi, provider: provider}
}

// AbiCaller is a wrapper to make abi calls to contract with a state provider
type AbiCaller struct {
	addr     ethgo.Address
	abi      *abi.ABI
	provider NodeProvider
}

func (a *AbiCaller) ABI() *abi.ABI {
	return a.abi
}

type TxnOpts struct {
	From     ethgo.Address
	Value    *big.Int
	GasPrice uint64
	GasLimit uint64
}

func (a *AbiCaller) Txn(method string, args ...interface{}) (Txn1, error) {
	m := a.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	input, err := m.Encode(args)
	if err != nil {
		return nil, err
	}
	txn, err := a.provider.Txn(a.addr, input)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

type CallOpts struct {
	Block ethgo.BlockNumber
	From  ethgo.Address
}

func (a *AbiCaller) Call(method string, block ethgo.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m := a.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}

	rawOutput, err := a.provider.Call(a.addr, data, block)
	if err != nil {
		return nil, err
	}

	resp, err := m.Decode(rawOutput)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
