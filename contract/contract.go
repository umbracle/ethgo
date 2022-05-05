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
	Txn(ethgo.Address, ethgo.Key, []byte) (Txn, error)
}

type jsonRPCNodeProvider struct {
	client  *jsonrpc.Eth
	eip1559 bool
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

func (j *jsonRPCNodeProvider) Txn(addr ethgo.Address, key ethgo.Key, input []byte) (Txn, error) {
	txn := &jsonrpcTransaction{
		opts:    &TxnOpts{},
		input:   input,
		client:  j.client,
		key:     key,
		to:      addr,
		eip1559: j.eip1559,
	}
	return txn, nil
}

type jsonrpcTransaction struct {
	to      ethgo.Address
	input   []byte
	hash    ethgo.Hash
	opts    *TxnOpts
	key     ethgo.Key
	client  *jsonrpc.Eth
	txn     *ethgo.Transaction
	txnRaw  []byte
	eip1559 bool
}

func (j *jsonrpcTransaction) Hash() ethgo.Hash {
	return j.hash
}

func (j *jsonrpcTransaction) WithOpts(opts *TxnOpts) {
	j.opts = opts
}

func (j *jsonrpcTransaction) Build() error {
	var err error
	from := j.key.Address()

	// estimate gas price
	if j.opts.GasPrice == 0 && !j.eip1559 {
		j.opts.GasPrice, err = j.client.GasPrice()
		if err != nil {
			return err
		}
	}
	// estimate gas limit
	if j.opts.GasLimit == 0 {
		msg := &ethgo.CallMsg{
			From:     from,
			To:       nil,
			Data:     j.input,
			Value:    j.opts.Value,
			GasPrice: j.opts.GasPrice,
		}
		if j.to != ethgo.ZeroAddress {
			msg.To = &j.to
		}
		j.opts.GasLimit, err = j.client.EstimateGas(msg)
		if err != nil {
			return err
		}
	}
	// calculate the nonce
	if j.opts.Nonce == 0 {
		j.opts.Nonce, err = j.client.GetNonce(from, ethgo.Latest)
		if err != nil {
			return fmt.Errorf("failed to calculate nonce: %v", err)
		}
	}

	chainID, err := j.client.ChainID()
	if err != nil {
		return err
	}

	// send transaction
	rawTxn := &ethgo.Transaction{
		From:     from,
		Input:    j.input,
		GasPrice: j.opts.GasPrice,
		Gas:      j.opts.GasLimit,
		Value:    j.opts.Value,
		Nonce:    j.opts.Nonce,
		ChainID:  chainID,
	}
	if j.to != ethgo.ZeroAddress {
		rawTxn.To = &j.to
	}

	if j.eip1559 {
		rawTxn.Type = ethgo.TransactionDynamicFee

		// use gas price as fee data
		gasPrice, err := j.client.GasPrice()
		if err != nil {
			return err
		}
		rawTxn.MaxFeePerGas = new(big.Int).SetUint64(gasPrice)
		rawTxn.MaxPriorityFeePerGas = new(big.Int).SetUint64(gasPrice)
	}

	j.txn = rawTxn
	return nil
}

func (j *jsonrpcTransaction) Do() error {
	if j.txn == nil {
		if err := j.Build(); err != nil {
			return err
		}
	}

	signer := wallet.NewEIP155Signer(j.txn.ChainID.Uint64())
	signedTxn, err := signer.SignTx(j.txn, j.key)
	if err != nil {
		return err
	}
	txnRaw, err := signedTxn.MarshalRLPTo(nil)
	if err != nil {
		return err
	}

	j.txnRaw = txnRaw
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
	WithOpts(opts *TxnOpts)
	Do() error
	Wait() (*ethgo.Receipt, error)
}

type Opts struct {
	JsonRPCEndpoint string
	JsonRPCClient   *jsonrpc.Eth
	Provider        Provider
	Sender          ethgo.Key
	EIP1559         bool
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

func WithEIP1559() ContractOption {
	return func(o *Opts) {
		o.EIP1559 = true
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
		provider = &jsonRPCNodeProvider{client: opt.JsonRPCClient, eip1559: opt.EIP1559}
	} else {
		client, _ := jsonrpc.NewClient(opt.JsonRPCEndpoint)
		provider = &jsonRPCNodeProvider{client: client.Eth(), eip1559: opt.EIP1559}
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

	txn, err := a.provider.Txn(a.addr, a.key, input)
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
