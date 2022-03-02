package contract

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/jsonrpc"
)

// Contract is an Ethereum contract
type Contract struct {
	addr     ethgo.Address
	from     *ethgo.Address
	abi      *abi.ABI
	provider *jsonrpc.Client
}

// DeployContract deploys a contract
func DeployContract(provider *jsonrpc.Client, from ethgo.Address, abi *abi.ABI, bin []byte, args ...interface{}) *Txn {
	return &Txn{
		from:     from,
		provider: provider,
		method:   abi.Constructor,
		args:     args,
		bin:      bin,
	}
}

// NewContract creates a new contract instance
func NewContract(addr ethgo.Address, abi *abi.ABI, provider *jsonrpc.Client) *Contract {
	return &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
	}
}

// ABI returns the abi of the contract
func (c *Contract) ABI() *abi.ABI {
	return c.abi
}

// Addr returns the address of the contract
func (c *Contract) Addr() ethgo.Address {
	return c.addr
}

// SetFrom sets the origin of the calls
func (c *Contract) SetFrom(addr ethgo.Address) {
	c.from = &addr
}

// EstimateGas estimates the gas for a contract call
func (c *Contract) EstimateGas(method string, args ...interface{}) (uint64, error) {
	return c.Txn(method, args).EstimateGas()
}

// Call calls a method in the contract
func (c *Contract) Call(method string, block ethgo.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m := c.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}

	// Call function
	msg := &ethgo.CallMsg{
		To:   &c.addr,
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
	resp, err := m.Decode(raw)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Txn creates a new transaction object
func (c *Contract) Txn(method string, args ...interface{}) *Txn {
	m, ok := c.abi.Methods[method]
	if !ok {
		// TODO, return error
		panic(fmt.Errorf("method %s not found", method))
	}

	return &Txn{
		from:     *c.from,
		addr:     &c.addr,
		provider: c.provider,
		method:   m,
		args:     args,
	}
}

// Txn is a transaction object
type Txn struct {
	from     ethgo.Address
	addr     *ethgo.Address
	provider *jsonrpc.Client
	method   *abi.Method
	args     []interface{}
	data     []byte
	bin      []byte
	gasLimit uint64
	gasPrice uint64
	value    *big.Int
	hash     ethgo.Hash
	receipt  *ethgo.Receipt
}

func (t *Txn) isContractDeployment() bool {
	return t.bin != nil
}

// AddArgs is used to set the arguments of the transaction
func (t *Txn) AddArgs(args ...interface{}) *Txn {
	t.args = args
	return t
}

// SetValue sets the value for the txn
func (t *Txn) SetValue(v *big.Int) *Txn {
	t.value = new(big.Int).Set(v)
	return t
}

// EstimateGas estimates the gas for the call
func (t *Txn) EstimateGas() (uint64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	return t.estimateGas()
}

func (t *Txn) estimateGas() (uint64, error) {
	if t.isContractDeployment() {
		return t.provider.Eth().EstimateGasContract(t.data)
	}

	msg := &ethgo.CallMsg{
		From:  t.from,
		To:    t.addr,
		Data:  t.data,
		Value: t.value,
	}
	return t.provider.Eth().EstimateGas(msg)
}

// DoAndWait is a blocking query that combines
// both Do and Wait functions
func (t *Txn) DoAndWait() error {
	if err := t.Do(); err != nil {
		return err
	}
	if err := t.Wait(); err != nil {
		return err
	}
	return nil
}

// Do sends the transaction to the network
func (t *Txn) Do() error {
	err := t.Validate()
	if err != nil {
		return err
	}

	// estimate gas price
	if t.gasPrice == 0 {
		t.gasPrice, err = t.provider.Eth().GasPrice()
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
	txn := &ethgo.Transaction{
		From:     t.from,
		Input:    t.data,
		GasPrice: t.gasPrice,
		Gas:      t.gasLimit,
		Value:    t.value,
	}
	if t.addr != nil {
		txn.To = t.addr
	}
	t.hash, err = t.provider.Eth().SendTransaction(txn)
	if err != nil {
		return err
	}
	return nil
}

// Validate validates the arguments of the transaction
func (t *Txn) Validate() error {
	if t.data != nil {
		// Already validated
		return nil
	}
	if t.isContractDeployment() {
		t.data = append(t.data, t.bin...)
	}
	if t.method != nil {
		data, err := abi.Encode(t.args, t.method.Inputs)
		if err != nil {
			return fmt.Errorf("failed to encode arguments: %v", err)
		}
		if !t.isContractDeployment() {
			t.data = append(t.method.ID(), data...)
		} else {
			t.data = append(t.data, data...)
		}
	}
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
	if (t.hash == ethgo.Hash{}) {
		panic("transaction not executed")
	}

	var err error
	for {
		t.receipt, err = t.provider.Eth().GetTransactionReceipt(t.hash)
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
func (t *Txn) Receipt() *ethgo.Receipt {
	return t.receipt
}

// Event is a solidity event
type Event struct {
	event *abi.Event
}

// Encode encodes an event
func (e *Event) Encode() ethgo.Hash {
	return e.event.ID()
}

// ParseLog parses a log
func (e *Event) ParseLog(log *ethgo.Log) (map[string]interface{}, error) {
	return abi.ParseLog(e.event.Inputs, log)
}

// Event returns a specific event
func (c *Contract) Event(name string) (*Event, bool) {
	event, ok := c.abi.Events[name]
	if !ok {
		return nil, false
	}
	return &Event{event}, true
}
