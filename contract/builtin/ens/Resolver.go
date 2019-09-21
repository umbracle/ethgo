package ens

import (
	"fmt"
	"math/big"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

// Resolver is a solidity contract
type Resolver struct {
	c *contract.Contract
}

// NewResolver creates a new instance of the contract at a specific address
func NewResolver(addr string, provider *jsonrpc.Client) *Resolver{
	return &Resolver{c: contract.NewContract(addr, abiResolver, provider)}
}

// Contract returns the contract object
func (a* Resolver) Contract() *contract.Contract {
	return a.c
}

// calls

// ABI calls the ABI method in the solidity contract
func (a* Resolver) ABI(node [32]byte, contentTypes *big.Int, block ...web3.BlockNumber) (val0 *big.Int, val1 []byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("ABI", web3.EncodeBlock(block...), node, contentTypes)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["contentType"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	val1, ok = out["data"].([]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 1")
		return
	}
	
	return
}

// Addr calls the addr method in the solidity contract
func (a* Resolver) Addr(node [32]byte, block ...web3.BlockNumber) (val0 [20]byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("addr", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["ret"].([20]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Content calls the content method in the solidity contract
func (a* Resolver) Content(node [32]byte, block ...web3.BlockNumber) (val0 [32]byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("content", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["ret"].([32]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Name calls the name method in the solidity contract
func (a* Resolver) Name(node [32]byte, block ...web3.BlockNumber) (val0 string, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("name", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["ret"].(string)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Pubkey calls the pubkey method in the solidity contract
func (a* Resolver) Pubkey(node [32]byte, block ...web3.BlockNumber) (val0 [32]byte, val1 [32]byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("pubkey", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["x"].([32]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	val1, ok = out["y"].([32]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 1")
		return
	}
	
	return
}

// SupportsInterface calls the supportsInterface method in the solidity contract
func (a* Resolver) SupportsInterface(interfaceID [4]byte, block ...web3.BlockNumber) (val0 bool, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("supportsInterface", web3.EncodeBlock(block...), interfaceID)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(bool)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}


// txns

// SetABI sends a setABI transaction in the solidity contract
func (a* Resolver) SetABI(node [32]byte, contentType *big.Int, data []byte) *contract.Txn {
	return a.c.Txn("setABI", node, contentType, data)
}

// SetAddr sends a setAddr transaction in the solidity contract
func (a* Resolver) SetAddr(node [32]byte, addr [20]byte) *contract.Txn {
	return a.c.Txn("setAddr", node, addr)
}

// SetContent sends a setContent transaction in the solidity contract
func (a* Resolver) SetContent(node [32]byte, hash [32]byte) *contract.Txn {
	return a.c.Txn("setContent", node, hash)
}

// SetName sends a setName transaction in the solidity contract
func (a* Resolver) SetName(node [32]byte, name string) *contract.Txn {
	return a.c.Txn("setName", node, name)
}

// SetPubkey sends a setPubkey transaction in the solidity contract
func (a* Resolver) SetPubkey(node [32]byte, x [32]byte, y [32]byte) *contract.Txn {
	return a.c.Txn("setPubkey", node, x, y)
}


var abiResolver *abi.ABI

func init() {
	var err error
	abiResolver, err = abi.NewABI(abiResolverStr)
	if err != nil {
		panic(fmt.Errorf("cannot parse Resolver abi: %v", err))
	}
}

var binResolver = []byte{}

var abiResolverStr = `[{"constant":true,"inputs":[{"name":"interfaceID","type":"bytes4"}],"name":"supportsInterface","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"},{"name":"contentTypes","type":"uint256"}],"name":"ABI","outputs":[{"name":"contentType","type":"uint256"},{"name":"data","type":"bytes"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"x","type":"bytes32"},{"name":"y","type":"bytes32"}],"name":"setPubkey","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"content","outputs":[{"name":"ret","type":"bytes32"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"addr","outputs":[{"name":"ret","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"contentType","type":"uint256"},{"name":"data","type":"bytes"}],"name":"setABI","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"name","outputs":[{"name":"ret","type":"string"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"name","type":"string"}],"name":"setName","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"hash","type":"bytes32"}],"name":"setContent","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"pubkey","outputs":[{"name":"x","type":"bytes32"},{"name":"y","type":"bytes32"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"addr","type":"address"}],"name":"setAddr","outputs":[],"payable":false,"type":"function"},{"inputs":[{"name":"ensAddr","type":"address"}],"payable":false,"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"a","type":"address"}],"name":"AddrChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"hash","type":"bytes32"}],"name":"ContentChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"name","type":"string"}],"name":"NameChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":true,"name":"contentType","type":"uint256"}],"name":"ABIChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"x","type":"bytes32"},{"indexed":false,"name":"y","type":"bytes32"}],"name":"PubkeyChanged","type":"event"}]`