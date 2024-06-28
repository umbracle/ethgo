package ens

import (
	"strings"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/contract"
	"github.com/Ethernal-Tech/ethgo/jsonrpc"
)

type ENSResolver struct {
	e        *ENS
	provider *jsonrpc.Eth
}

func NewENSResolver(addr ethgo.Address, provider *jsonrpc.Client) *ENSResolver {
	return &ENSResolver{NewENS(addr, contract.WithJsonRPC(provider.Eth())), provider.Eth()}
}

func (e *ENSResolver) Resolve(addr string, block ...ethgo.BlockNumber) (res ethgo.Address, err error) {
	addrHash := NameHash(addr)
	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Addr(addrHash, block...)
	return
}

func addressToReverseDomain(addr ethgo.Address) string {
	return strings.ToLower(strings.TrimPrefix(addr.String(), "0x")) + ".addr.reverse"
}

func (e *ENSResolver) ReverseResolve(addr ethgo.Address, block ...ethgo.BlockNumber) (res string, err error) {
	addrHash := NameHash(addressToReverseDomain(addr))

	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Name(addrHash, block...)
	return
}

func (e *ENSResolver) prepareResolver(inputHash ethgo.Hash, block ...ethgo.BlockNumber) (*Resolver, error) {
	resolverAddr, err := e.e.Resolver(inputHash, block...)
	if err != nil {
		return nil, err
	}

	resolver := NewResolver(resolverAddr, contract.WithJsonRPC(e.provider))
	return resolver, nil
}
