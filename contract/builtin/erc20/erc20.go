package erc20

import (
	"fmt"
	"math/big"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

var (
	_ = big.NewInt
)

// ERC20 is a solidity contract
type ERC20 struct {
	c *contract.Contract
}

// NewERC20 creates a new instance of the contract at a specific address
func NewERC20(addr web3.Address, provider *jsonrpc.Client) *ERC20 {
	return &ERC20{c: contract.NewContract(addr, abiERC20, provider)}
}

// Contract returns the contract object
func (a *ERC20) Contract() *contract.Contract {
	return a.c
}

// calls

// Allowance calls the allowance method in the solidity contract
func (a *ERC20) Allowance(owner web3.Address, spender web3.Address, block ...web3.BlockNumber) (retval0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("allowance", web3.EncodeBlock(block...), owner, spender)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// BalanceOf calls the balanceOf method in the solidity contract
func (a *ERC20) BalanceOf(owner web3.Address, block ...web3.BlockNumber) (retval0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("balanceOf", web3.EncodeBlock(block...), owner)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["balance"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Decimals calls the decimals method in the solidity contract
func (a *ERC20) Decimals(block ...web3.BlockNumber) (retval0 uint8, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("decimals", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(uint8)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Name calls the name method in the solidity contract
func (a *ERC20) Name(block ...web3.BlockNumber) (retval0 string, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("name", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(string)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Symbol calls the symbol method in the solidity contract
func (a *ERC20) Symbol(block ...web3.BlockNumber) (retval0 string, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("symbol", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(string)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// TotalSupply calls the totalSupply method in the solidity contract
func (a *ERC20) TotalSupply(block ...web3.BlockNumber) (retval0 *big.Int, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("totalSupply", web3.EncodeBlock(block...))
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// txns

// Approve sends a approve transaction in the solidity contract
func (a *ERC20) Approve(spender web3.Address, value *big.Int) *contract.Txn {
	return a.c.Txn("approve", spender, value)
}

// Transfer sends a transfer transaction in the solidity contract
func (a *ERC20) Transfer(to web3.Address, value *big.Int) *contract.Txn {
	return a.c.Txn("transfer", to, value)
}

// TransferFrom sends a transferFrom transaction in the solidity contract
func (a *ERC20) TransferFrom(from web3.Address, to web3.Address, value *big.Int) *contract.Txn {
	return a.c.Txn("transferFrom", from, to, value)
}

// events

func (a *ERC20) ApprovalEventSig() web3.Hash {
	return a.c.ABI().Events["Approval"].ID()
}

func (a *ERC20) TransferEventSig() web3.Hash {
	return a.c.ABI().Events["Transfer"].ID()
}
