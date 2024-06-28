package ens

import (
	"encoding/hex"
	"fmt"

	"github.com/Ethernal-Tech/ethgo/abi"
)

var abiENS *abi.ABI

// ENSAbi returns the abi of the ENS contract
func ENSAbi() *abi.ABI {
	return abiENS
}

var binENS []byte

// ENSBin returns the bin of the ENS contract
func ENSBin() []byte {
	return binENS
}

func init() {
	var err error
	abiENS, err = abi.NewABI(abiENSStr)
	if err != nil {
		panic(fmt.Errorf("cannot parse ENS abi: %v", err))
	}
	if len(binENSStr) != 0 {
		binENS, err = hex.DecodeString(binENSStr[2:])
		if err != nil {
			panic(fmt.Errorf("cannot parse ENS bin: %v", err))
		}
	}
}

var binENSStr = "0x6060604052341561000f57600080fd5b60008080526020527fad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb58054600160a060020a033316600160a060020a0319909116179055610503806100626000396000f3006060604052600436106100825763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630178b8bf811461008757806302571be3146100b957806306ab5923146100cf57806314ab9038146100f657806316a25cbd146101195780631896f70a1461014c5780635b0fc9c31461016e575b600080fd5b341561009257600080fd5b61009d600435610190565b604051600160a060020a03909116815260200160405180910390f35b34156100c457600080fd5b61009d6004356101ae565b34156100da57600080fd5b6100f4600435602435600160a060020a03604435166101c9565b005b341561010157600080fd5b6100f460043567ffffffffffffffff6024351661028b565b341561012457600080fd5b61012f600435610357565b60405167ffffffffffffffff909116815260200160405180910390f35b341561015757600080fd5b6100f4600435600160a060020a036024351661038e565b341561017957600080fd5b6100f4600435600160a060020a0360243516610434565b600090815260208190526040902060010154600160a060020a031690565b600090815260208190526040902054600160a060020a031690565b600083815260208190526040812054849033600160a060020a039081169116146101f257600080fd5b8484604051918252602082015260409081019051908190039020915083857fce0457fe73731f824cc272376169235128c118b49d344817417c6d108d155e8285604051600160a060020a03909116815260200160405180910390a3506000908152602081905260409020805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03929092169190911790555050565b600082815260208190526040902054829033600160a060020a039081169116146102b457600080fd5b827f1d4f9bbfc9cab89d66e1a1562f2233ccbf1308cb4f63de2ead5787adddb8fa688360405167ffffffffffffffff909116815260200160405180910390a250600091825260208290526040909120600101805467ffffffffffffffff90921674010000000000000000000000000000000000000000027fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff909216919091179055565b60009081526020819052604090206001015474010000000000000000000000000000000000000000900467ffffffffffffffff1690565b600082815260208190526040902054829033600160a060020a039081169116146103b757600080fd5b827f335721b01866dc23fbee8b6b2c7b1e14d6f05c28cd35a2c934239f94095602a083604051600160a060020a03909116815260200160405180910390a250600091825260208290526040909120600101805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a03909216919091179055565b600082815260208190526040902054829033600160a060020a0390811691161461045d57600080fd5b827fd4735d920b0f87494915f556dd9b54c8f309026070caea5c737245152564d26683604051600160a060020a03909116815260200160405180910390a250600091825260208290526040909120805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a039092169190911790555600a165627a7a72305820f4c798d4c84c9912f389f64631e85e8d16c3e6644f8c2e1579936015c7d5f6660029"

var abiENSStr = `[{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"resolver","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"label","type":"bytes32"},{"name":"owner","type":"address"}],"name":"setSubnodeOwner","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"ttl","type":"uint64"}],"name":"setTTL","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"node","type":"bytes32"}],"name":"ttl","outputs":[{"name":"","type":"uint64"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"resolver","type":"address"}],"name":"setResolver","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"node","type":"bytes32"},{"name":"owner","type":"address"}],"name":"setOwner","outputs":[],"payable":false,"type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"owner","type":"address"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":true,"name":"label","type":"bytes32"},{"indexed":false,"name":"owner","type":"address"}],"name":"NewOwner","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"resolver","type":"address"}],"name":"NewResolver","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"node","type":"bytes32"},{"indexed":false,"name":"ttl","type":"uint64"}],"name":"NewTTL","type":"event"}]`
