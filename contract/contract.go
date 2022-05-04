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

// Provider handles the interactions with the Ethereum 1x node
type Provider interface {
	Call(ethgo.Address, []byte, *CallOpts) ([]byte, error)
	Txn(ethgo.Address, ethgo.Key, []byte, *TxnOpts) (Txn, error)
}

type jsonRPCNodeProvider struct {
	client *jsonrpc.Eth
}

func (j *jsonRPCNodeProvider) Call(addr ethgo.Address, input []byte, opts *CallOpts) ([]byte, error) {
	msg := &ethgo.CallMsg{
		To:   &addr,
		Data: input,
	}
	if opts.From != ethgo.ZeroAddress {
		msg.From = opts.From
	}
	rawStr, err := j.client.Call(msg, opts.Block)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (j *jsonRPCNodeProvider) Txn(addr ethgo.Address, key ethgo.Key, input []byte, opts *TxnOpts) (Txn, error) {
	var err error

	from := key.Address()

	// estimate gas price
	if opts.GasPrice == 0 {
		opts.GasPrice, err = j.client.GasPrice()
		if err != nil {
			return nil, err
		}
	}
	// estimate gas limit
	if opts.GasLimit == 0 {
		msg := &ethgo.CallMsg{
			From:     from,
			To:       nil,
			Data:     input,
			Value:    opts.Value,
			GasPrice: opts.GasPrice,
		}
		if addr != ethgo.ZeroAddress {
			msg.To = &addr
		}
		opts.GasLimit, err = j.client.EstimateGas(msg)
		if err != nil {
			return nil, err
		}
	}
	// calculate the nonce
	if opts.Nonce == 0 {
		opts.Nonce, err = j.client.GetNonce(from, ethgo.Latest)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate nonce: %v", err)
		}
	}

	chainID, err := j.client.ChainID()
	if err != nil {
		return nil, err
	}

	// send transaction
	rawTxn := &ethgo.Transaction{
		From:     from,
		Input:    input,
		GasPrice: opts.GasPrice,
		Gas:      opts.GasLimit,
		Value:    opts.Value,
		Nonce:    opts.Nonce,
	}
	if addr != ethgo.ZeroAddress {
		rawTxn.To = &addr
	}

	signer := wallet.NewEIP155Signer(chainID.Uint64())
	signedTxn, err := signer.SignTx(rawTxn, key)
	if err != nil {
		return nil, err
	}
	txnRaw, err := signedTxn.MarshalRLPTo(nil)
	if err != nil {
		return nil, err
	}

	txn := &jsonrpcTransaction{
		txn:    signedTxn,
		txnRaw: txnRaw,
		client: j.client,
	}
	return txn, nil
}

type jsonrpcTransaction struct {
	hash   ethgo.Hash
	client *jsonrpc.Eth
	txn    *ethgo.Transaction
	txnRaw []byte
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
	hash, err := j.client.SendRawTransaction(j.txnRaw)
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
		receipt, err := j.client.GetTransactionReceipt(j.hash)
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

// Txn is the transaction object returned
type Txn interface {
	Hash() ethgo.Hash
	EstimatedGas() uint64
	GasPrice() uint64
	Do() error
	Wait() (*ethgo.Receipt, error)
}

type Opts struct {
	JsonRPCEndpoint string
	JsonRPCClient   *jsonrpc.Eth
	Provider        Provider
	Sender          ethgo.Key
}

type ContractOption func(*Opts)

func WithJsonRPCEndpoint(endpoint string) ContractOption {
	return func(o *Opts) {
		o.JsonRPCEndpoint = endpoint
	}
}

func WithJsonRPC(client *jsonrpc.Eth) ContractOption {
	return func(o *Opts) {
		o.JsonRPCClient = client
	}
}

func WithProvider(provider Provider) ContractOption {
	return func(o *Opts) {
		o.Provider = provider
	}
}

func WithSender(sender ethgo.Key) ContractOption {
	return func(o *Opts) {
		o.Sender = sender
	}
}

func DeployContract(abi *abi.ABI, bin []byte, args []interface{}, opts ...ContractOption) (Txn, error) {
	a := NewContract(ethgo.Address{}, abi, opts...)
	a.bin = bin
	return a.Txn("constructor", args...)
}

func NewContract(addr ethgo.Address, abi *abi.ABI, opts ...ContractOption) *Contract {
	opt := &Opts{
		JsonRPCEndpoint: "http://localhost:8545",
	}
	for _, c := range opts {
		c(opt)
	}

	var provider Provider
	if opt.Provider != nil {
		provider = opt.Provider
	} else if opt.JsonRPCClient != nil {
		provider = &jsonRPCNodeProvider{client: opt.JsonRPCClient}
	} else {
		client, _ := jsonrpc.NewClient(opt.JsonRPCEndpoint)
		provider = &jsonRPCNodeProvider{client: client.Eth()}
	}

	a := &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
		key:      opt.Sender,
	}

	return a
}

// Contract is a wrapper to make abi calls to contract with a state provider
type Contract struct {
	addr     ethgo.Address
	abi      *abi.ABI
	bin      []byte
	provider Provider
	key      ethgo.Key
}

func (a *Contract) GetABI() *abi.ABI {
	return a.abi
}

type TxnOpts struct {
	Value    *big.Int
	GasPrice uint64
	GasLimit uint64
	Nonce    uint64
}

func (a *Contract) Txn(method string, args ...interface{}) (Txn, error) {
	if a.key == nil {
		return nil, fmt.Errorf("no key selected")
	}

	isContractDeployment := method == "constructor"

	var input []byte
	if isContractDeployment {
		input = append(input, a.bin...)
	}

	var abiMethod *abi.Method
	if isContractDeployment {
		if a.abi.Constructor != nil {
			abiMethod = a.abi.Constructor
		}
	} else {
		if abiMethod = a.abi.GetMethod(method); abiMethod == nil {
			return nil, fmt.Errorf("method %s not found", method)
		}
	}
	if abiMethod != nil {
		data, err := abi.Encode(args, abiMethod.Inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode arguments: %v", err)
		}
		if isContractDeployment {
			input = append(input, data...)
		} else {
			input = append(abiMethod.ID(), data...)
		}
	}

	txn, err := a.provider.Txn(a.addr, a.key, input, &TxnOpts{})
	if err != nil {
		return nil, err
	}
	return txn, nil
}

type CallOpts struct {
	Block ethgo.BlockNumber
	From  ethgo.Address
}

func (a *Contract) Call(method string, block ethgo.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m := a.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}

	opts := &CallOpts{
		Block: block,
	}
	if a.key != nil {
		opts.From = a.key.Address()
	}
	rawOutput, err := a.provider.Call(a.addr, data, opts)
	if err != nil {
		return nil, err
	}

	resp, err := m.Decode(rawOutput)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
