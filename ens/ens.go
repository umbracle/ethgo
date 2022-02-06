package ens

import (
	"fmt"
	"log"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/contract/builtin/ens"
	"github.com/umbracle/go-web3/jsonrpc"
)

type EnsConfig struct {
	Logger   *log.Logger
	Client   *jsonrpc.Client
	Addr     string
	Resolver web3.Address
}

type EnsOption func(*EnsConfig)

func WithResolver(resolver web3.Address) EnsOption {
	return func(c *EnsConfig) {
		c.Resolver = resolver
	}
}

func WithLogger(logger *log.Logger) EnsOption {
	return func(c *EnsConfig) {
		c.Logger = logger
	}
}

func WithAddress(addr string) EnsOption {
	return func(c *EnsConfig) {
		c.Addr = addr
	}
}

func WithClient(client *jsonrpc.Client) EnsOption {
	return func(c *EnsConfig) {
		c.Client = client
	}
}

type ENS struct {
	config *EnsConfig
}

func NewENS(opts ...EnsOption) (*ENS, error) {
	config := &EnsConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.Client == nil {
		// addr must be set
		if config.Addr == "" {
			return nil, fmt.Errorf("jsonrpc addr is empty")
		}
		client, err := jsonrpc.NewClient(config.Addr)
		if err != nil {
			return nil, err
		}
		config.Client = client
	}

	if config.Resolver == web3.ZeroAddress {
		// try to get the resolver address from the builtin list
		chainID, err := config.Client.Eth().ChainID()
		if err != nil {
			return nil, err
		}
		addr, ok := builtinEnsAddr[chainID.Uint64()]
		if !ok {
			return nil, fmt.Errorf("no builtin Ens resolver found for chain %s", chainID)
		}
		config.Resolver = addr
	}
	ens := &ENS{
		config: config,
	}
	return ens, nil
}

func (e *ENS) Resolve(name string) (web3.Address, error) {
	resolver := ens.NewENSResolver(e.config.Resolver, e.config.Client)
	return resolver.Resolve(name)
}
