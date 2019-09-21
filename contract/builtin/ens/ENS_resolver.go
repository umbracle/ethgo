package ens

import (
	"encoding/hex"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

type ENSResolver struct {
	e        *ENS
	provider *jsonrpc.Client
}

func NewENSResolver(addr string, provider *jsonrpc.Client) *ENSResolver {
	return &ENSResolver{NewENS(addr, provider), provider}
}

func (e *ENSResolver) Resolve(addr string, block ...web3.BlockNumber) (res [20]byte, err error) {
	aux := NameHash(addr)
	addrHash := [32]byte{}
	copy(addrHash[:], aux)

	resolverAddr, err := e.e.Resolver(addrHash, block...)
	if err != nil {
		return
	}

	resolver := NewResolver("0x"+hex.EncodeToString(resolverAddr[:]), e.provider)
	res, err = resolver.Addr(addrHash, block...)
	return 
}
