package gaspricer

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/umbracle/go-web3/jsonrpc"
)

func NewNetworkGasPricer(logger *log.Logger, client *jsonrpc.Eth, interval ...time.Duration) (GasPricer, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	n := &NetworkGasPricer{
		client: client,
		logger: logger,
	}
	if len(interval) == 1 {
		n.interval = interval[0]
	} else {
		n.interval = 15 * time.Second
	}

	// try to fetch the gas price once
	if err := n.updateGasPrice(); err != nil {
		return nil, err
	}
	return n, nil
}

type NetworkGasPricer struct {
	logger   *log.Logger
	lock     sync.Mutex
	client   *jsonrpc.Eth
	interval time.Duration
	gasPrice uint64
}

func (n *NetworkGasPricer) updateGasPrice() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	gasPrice, err := n.client.GasPrice()
	if err != nil {
		return err
	}
	n.gasPrice = gasPrice
	return nil
}

func (n *NetworkGasPricer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(n.interval):
				if err := n.updateGasPrice(); err != nil {
					n.logger.Printf("[ERROR]: Failed to get gas price: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (n *NetworkGasPricer) GasPrice() uint64 {
	n.lock.Lock()
	defer n.lock.Unlock()

	return n.gasPrice
}
