package relayer

import (
	"log"
	"os"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/relayer/gaspricer"
)

type Config struct {
	Logger    *log.Logger
	Endpoint  string
	GasPricer gaspricer.GasPricer
}

type RelayerOption func(*Config)

func WithGasPricer(pricer gaspricer.GasPricer) RelayerOption {
	return func(c *Config) {
		c.GasPricer = pricer
	}
}

func WithJSONRPCEndpoint(endpoint string) RelayerOption {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

func WithLogger(logger *log.Logger) RelayerOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

func DefaultConfig() *Config {
	return &Config{
		Logger:   log.New(os.Stdout, "", log.LstdFlags),
		Endpoint: "http://localhost:8545",
	}
}

type Relayer struct {
	config  *Config
	client  *jsonrpc.Client
	closeCh chan struct{}
}

func NewRelayer(configOpts ...RelayerOption) (*Relayer, error) {
	config := DefaultConfig()
	for _, opts := range configOpts {
		opts(config)
	}

	client, err := jsonrpc.NewClient(config.Endpoint)
	if err != nil {
		return nil, err
	}

	// if gas pricer is not set, use the network one
	if config.GasPricer == nil {
		pricer, err := gaspricer.NewNetworkGasPricer(config.Logger, client.Eth())
		if err != nil {
			return nil, err
		}
		config.GasPricer = pricer
	}

	r := &Relayer{
		config:  config,
		client:  client,
		closeCh: make(chan struct{}),
	}

	// fill in the pending queue with the pending transactions
	// from the storage

	go r.run()

	return r, nil
}

func (r *Relayer) SendTransaction(txn *web3.Transaction) (web3.Hash, error) {
	return web3.Hash{}, nil
}

func (r *Relayer) run() {
	for {
		// wait to see if anything is stucked
		// after it is done move to the next pending txn
		select {
		case <-r.closeCh:
			return
		}
	}
}
