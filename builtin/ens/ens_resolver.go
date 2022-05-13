package ens

import (
	"github.com/cloudwalk/ethgo"
	"github.com/cloudwalk/ethgo/contract"
	"github.com/cloudwalk/ethgo/jsonrpc"
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
	resolverAddr, err := e.e.Resolver(addrHash, block...)
	if err != nil {
		return
	}

	resolver := NewResolver(resolverAddr, contract.WithJsonRPC(e.provider))
	res, err = resolver.Addr(addrHash, block...)
	return
}
