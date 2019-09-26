package contract

import (
	"encoding/hex"
	"fmt"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/jsonrpc"
)

// Contract is an Ethereum contract
type Contract struct {
	addr     web3.Address
	from     *web3.Address
	abi      *abi.ABI
	provider *jsonrpc.Client
}

// NewContract creates a new contract instance
func NewContract(addr web3.Address, abi *abi.ABI, provider *jsonrpc.Client) *Contract {
	return &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
	}
}

// Addr returns the address of the contract
func (c *Contract) Addr() web3.Address {
	return c.addr
}

// SetFrom sets the origin of the calls
func (c *Contract) SetFrom(addr web3.Address) {
	c.from = &addr
}

// EstimateGas estimates the gas for a contract call
func (c *Contract) EstimateGas(method string, args ...interface{}) (uint64, error) {
	return c.Txn(method, args).EstimateGas()
}

const emptyAddr = "0x0000000000000000000000000000000000000000"

// Call calls a method in the contract
func (c *Contract) Call(method string, block web3.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m, ok := c.abi.Methods[method]
	if !ok {
		return nil, fmt.Errorf("method %s not found", method)
	}

	// Encode input
	data, err := abi.Encode(args, m.Inputs.Type())
	if err != nil {
		return nil, err
	}
	data = append(m.ID(), data...)

	// Call function
	msg := &web3.CallMsg{
		To:   c.addr,
		Data: data,
	}
	if c.from != nil {
		msg.From = *c.from
	}

	rawStr, err := c.provider.Eth().Call(msg, block)
	if err != nil {
		return nil, err
	}

	// Decode output
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	respInterface, err := abi.Decode(m.Outputs.Type(), raw)
	if err != nil {
		return nil, err
	}

	resp := respInterface.(map[string]interface{})
	return resp, nil
}

// Txn creates a new transaction object
func (c *Contract) Txn(method string, args ...interface{}) *Txn {
	return &Txn{
		contract: c,
		method:   method,
		args:     args,
	}
}

// Txn is a transaction object
type Txn struct {
	contract *Contract
	method   string
	args     []interface{}
	data     []byte
	gasLimit uint64
	gasPrice uint64
	hash     web3.Hash
	receipt  *web3.Receipt
}

// EstimateGas estimates the gas for the call
func (t *Txn) EstimateGas() (uint64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	return t.estimateGas()
}

func (t *Txn) estimateGas() (uint64, error) {
	msg := &web3.CallMsg{
		From: *t.contract.from,
		To:   t.contract.addr,
		Data: t.data,
	}
	return t.contract.provider.Eth().EstimateGas(msg)
}

// Do sends the transaction to the network
func (t *Txn) Do() error {
	err := t.Validate()
	if err != nil {
		return err
	}

	// estimate gas price
	if t.gasPrice == 0 {
		t.gasPrice, err = t.contract.provider.Eth().GasPrice()
		if err != nil {
			return err
		}
	}
	// estimate gas limit
	if t.gasLimit == 0 {
		t.gasLimit, err = t.estimateGas()
		if err != nil {
			return err
		}
	}

	// send transaction
	txn := &web3.Transaction{
		From:     *t.contract.from,
		To:       t.contract.addr.String(),
		Input:    t.data,
		GasPrice: t.gasPrice,
		Gas:      t.gasLimit,
	}
	t.hash, err = t.contract.provider.Eth().SendTransaction(txn)
	if err != nil {
		return err
	}
	return nil
}

// Validate validates the arguments of the transaction
func (t *Txn) Validate() error {
	if t.data == nil {
		// Already validated
		return nil
	}
	m, ok := t.contract.abi.Methods[t.method]
	if !ok {
		return fmt.Errorf("method %s not found", t.method)
	}
	data, err := abi.Encode(t.args, m.Inputs.Type())
	if err != nil {
		return fmt.Errorf("failed to encode arguments: %v", err)
	}
	t.data = append(m.ID(), data...)
	return nil
}

// SetGasPrice sets the gas price of the transaction
func (t *Txn) SetGasPrice(gasPrice uint64) *Txn {
	t.gasPrice = gasPrice
	return t
}

// SetGasLimit sets the gas limit of the transaction
func (t *Txn) SetGasLimit(gasLimit uint64) *Txn {
	t.gasLimit = gasLimit
	return t
}

// Wait waits till the transaction is mined
func (t *Txn) Wait() error {
	if (t.hash == web3.Hash{}) {
		panic("transaction not executed")
	}

	var err error
	for {
		t.receipt, err = t.contract.provider.Eth().GetTransactionReceipt(t.hash)
		if err != nil {
			if err.Error() != "not found" {
				return err
			}
		}
		if t.receipt != nil {
			break
		}
	}
	return nil
}

// Receipt returns the receipt of the transaction after wait
func (t *Txn) Receipt() *web3.Receipt {
	return t.receipt
}

// Event parses a specific event
func (c *Contract) Event(name string, log *web3.Log) (map[string]interface{}, error) {
	event, ok := c.abi.Events[name]
	if !ok {
		return nil, fmt.Errorf("event %s not found", name)
	}
	return abi.ParseLog(event.Inputs, log)
}
