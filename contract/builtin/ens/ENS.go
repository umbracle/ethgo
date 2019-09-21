package ens

import (
	"fmt"
	
	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/contract"
	"github.com/umbracle/go-web3/jsonrpc"
)

// ENS is a solidity contract
type ENS struct {
	c *contract.Contract
}

// NewENS creates a new instance of the contract at a specific address
func NewENS(addr string, provider *jsonrpc.Client) *ENS{
	return &ENS{c: contract.NewContract(addr, abiENS, provider)}
}

// Contract returns the contract object
func (a* ENS) Contract() *contract.Contract {
	return a.c
}

// calls

// Owner calls the owner method in the solidity contract
func (a* ENS) Owner(node [32]byte, block ...web3.BlockNumber) (val0 [20]byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("owner", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].([20]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Resolver calls the resolver method in the solidity contract
func (a* ENS) Resolver(node [32]byte, block ...web3.BlockNumber) (val0 [20]byte, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("resolver", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].([20]byte)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Ttl calls the ttl method in the solidity contract
func (a* ENS) Ttl(node [32]byte, block ...web3.BlockNumber) (val0 uint64, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = a.c.Call("ttl", web3.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	val0, ok = out["0"].(uint64)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}


// txns

// SetOwner sends a setOwner transaction in the solidity contract
func (a* ENS) SetOwner(node [32]byte, owner [20]byte) *contract.Txn {
	return a.c.Txn("setOwner", node, owner)
}

// SetResolver sends a setResolver transaction in the solidity contract
func (a* ENS) SetResolver(node [32]byte, resolver [20]byte) *contract.Txn {
	return a.c.Txn("setResolver", node, resolver)
}

// SetSubnodeOwner sends a setSubnodeOwner transaction in the solidity contract
func (a* ENS) SetSubnodeOwner(node [32]byte, label [32]byte, owner [20]byte) *contract.Txn {
	return a.c.Txn("setSubnodeOwner", node, label, owner)
}

// SetTTL sends a setTTL transaction in the solidity contract
func (a* ENS) SetTTL(node [32]byte, ttl uint64) *contract.Txn {
	return a.c.Txn("setTTL", node, ttl)
}


var abiENS *abi.ABI

func init() {
	var err error
	abiENS, err = abi.NewABI(abiENSStr)
	if err != nil {
		panic(fmt.Errorf("cannot parse ENS abi: %v", err))
	}
}

var binENS = []byte{}

var abiENSStr = `[{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"resolver","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"label","type":"bytes32"},{"name":"owner","type":"address"}],"name":"setSubnodeOwner","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"ttl","type":"uint64"}],"name":"setTTL","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"ttl","outputs":[{"name":"","type":"uint64"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"resolver","type":"address"}],"name":"setResolver","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"owner","type":"address"}],"name":"setOwner","outputs":[],"payable":false,"type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"owner","type":"address"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":true,"name":"label","type":"bytes32"},{"indexed":false,"name":"owner","type":"address"}],"name":"NewOwner","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"resolver","type":"address"}],"name":"NewResolver","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"ttl","type":"uint64"}],"name":"NewTTL","type":"event"}]`