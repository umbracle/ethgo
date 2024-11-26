package relayer

import (
	"fmt"
	"log"
	"os"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/relayer/gaspricer"
)

var _ contract.Provider = (*Relayer)(nil)

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

	go r.run()

	return r, nil
}

type stateTxn struct {
	To    ethgo.Address
	key   ethgo.Key
	input []byte
	opts  *contract.TxnOpts
}

func (r *Relayer) Txn(to ethgo.Address, key ethgo.Key, input []byte, opts *contract.TxnOpts) (contract.Txn, error) {
	txn := &stateTxn{}
	fmt.Println(txn)
	return nil, nil
}

func (r *Relayer) SendTransaction(txn *ethgo.Transaction) (ethgo.Hash, error) {
	return ethgo.Hash{}, nil
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

func (r *Relayer) Call(to ethgo.Address, input []byte, opts *contract.CallOpts) ([]byte, error) {
	panic("relayer does not make calls")
}
