package ens

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

// ENS is a solidity contract
type ENS struct {
	c *contract.Contract
}

// DeployENS deploys a new ENS contract
func DeployENS(provider *jsonrpc.Client, from web3.Address, args ...interface{}) *contract.Txn {
	return contract.DeployContract(provider, from, abiENS, binENS, args...)
}

// NewENS creates a new instance of the contract at a specific address
func NewENS(addr web3.Address, provider *jsonrpc.Client) *ENS {
	return &ENS{c: contract.NewContract(addr, abiENS, provider)}
}

// Contract returns the contract object
func (a *ENS) Contract() *contract.Contract {
	return a.c
}

// calls

// Owner calls the owner method in the solidity contract
func (a *ENS) Owner(node [32]byte, block ...web3.BlockNumber) (retval0 web3.Address, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("owner", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(web3.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}

	return
}

// Resolver calls the resolver method in the solidity contract
func (a *ENS) Resolver(node [32]byte, block ...web3.BlockNumber) (retval0 web3.Address, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("resolver", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(web3.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}

	return
}

// Ttl calls the ttl method in the solidity contract
func (a *ENS) Ttl(node [32]byte, block ...web3.BlockNumber) (retval0 uint64, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("ttl", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(uint64)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}

	return
}

// txns

// SetOwner sends a setOwner transaction in the solidity contract
func (a *ENS) SetOwner(node [32]byte, owner web3.Address) *contract.Txn {
	return a.c.Txn("setOwner", node, owner)
}

// SetResolver sends a setResolver transaction in the solidity contract
func (a *ENS) SetResolver(node [32]byte, resolver web3.Address) *contract.Txn {
	return a.c.Txn("setResolver", node, resolver)
}

// SetSubnodeOwner sends a setSubnodeOwner transaction in the solidity contract
func (a *ENS) SetSubnodeOwner(node [32]byte, label [32]byte, owner web3.Address) *contract.Txn {
	return a.c.Txn("setSubnodeOwner", node, label, owner)
}

// SetTTL sends a setTTL transaction in the solidity contract
func (a *ENS) SetTTL(node [32]byte, ttl uint64) *contract.Txn {
	return a.c.Txn("setTTL", node, ttl)
}

// events

func (a *ENS) NewOwnerEventSig() web3.Hash {
	return a.c.ABI().Events["NewOwner"].ID()
}

func (a *ENS) NewResolverEventSig() web3.Hash {
	return a.c.ABI().Events["NewResolver"].ID()
}

func (a *ENS) NewTTLEventSig() web3.Hash {
	return a.c.ABI().Events["NewTTL"].ID()
}

func (a *ENS) TransferEventSig() web3.Hash {
	return a.c.ABI().Events["Transfer"].ID()
}
