package ens

import (
	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

type ENSResolver struct {
	e        *ENS
	provider *jsonrpc.Client
}

func NewENSResolver(addr web3.Address, provider *jsonrpc.Client) *ENSResolver {
	return &ENSResolver{NewENS(addr, provider), provider}
}

func (e *ENSResolver) Resolve(addr string, block ...web3.BlockNumber) (res web3.Address, err error) {
	addrHash := NameHash(addr)
	resolverAddr, err := e.e.Resolver(addrHash, block...)
	if err != nil {
		return
	}

	resolver := NewResolver(resolverAddr, e.provider)
	res, err = resolver.Addr(addrHash, block...)
	return
}
