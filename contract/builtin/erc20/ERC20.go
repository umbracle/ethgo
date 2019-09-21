package erc20

import (
	"fmt"
	"math/big"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

// ERC20 is a solidity contract
type ERC20 struct {
	c *contract.Contract
}

// NewERC20 creates a new instance of the contract at a specific address
func NewERC20(addr string, provider *jsonrpc.Client) *ERC20{
	return &ERC20{c: contract.NewContract(addr, abiERC20, provider)}
}

// Contract returns the contract object
func (a* ERC20) Contract() *contract.Contract {
	return a.c
}

// calls

// Allowance calls the allowance method in the solidity contract
func (a* ERC20) Allowance(owner [20]byte, spender [20]byte, block ...web3.BlockNumber) (val0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("allowance", web3.EncodeBlock(block...), owner, spender)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// BalanceOf calls the balanceOf method in the solidity contract
func (a* ERC20) BalanceOf(owner [20]byte, block ...web3.BlockNumber) (val0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("balanceOf", web3.EncodeBlock(block...), owner)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["balance"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Decimals calls the decimals method in the solidity contract
func (a* ERC20) Decimals(block ...web3.BlockNumber) (val0 uint8, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("decimals", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(uint8)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Name calls the name method in the solidity contract
func (a* ERC20) Name(block ...web3.BlockNumber) (val0 string, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("name", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(string)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Symbol calls the symbol method in the solidity contract
func (a* ERC20) Symbol(block ...web3.BlockNumber) (val0 string, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("symbol", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(string)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// TotalSupply calls the totalSupply method in the solidity contract
func (a* ERC20) TotalSupply(block ...web3.BlockNumber) (val0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("totalSupply", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}


// txns

// Approve sends a approve transaction in the solidity contract
func (a* ERC20) Approve(spender [20]byte, value *big.Int) *contract.Txn {
	return a.c.Txn("approve", spender, value)
}

// Transfer sends a transfer transaction in the solidity contract
func (a* ERC20) Transfer(to [20]byte, value *big.Int) *contract.Txn {
	return a.c.Txn("transfer", to, value)
}

// TransferFrom sends a transferFrom transaction in the solidity contract
func (a* ERC20) TransferFrom(from [20]byte, to [20]byte, value *big.Int) *contract.Txn {
	return a.c.Txn("transferFrom", from, to, value)
}


var abiERC20 *abi.ABI

func init() {
	var err error
	abiERC20, err = abi.NewABI(abiERC20Str)
	if err != nil {
		panic(fmt.Errorf("cannot parse ERC20 abi: %v", err))
	}
}

var binERC20 = []byte{}

var abiERC20Str = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`