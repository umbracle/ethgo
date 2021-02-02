package tracker

import (
	"context"
	"fmt"
	"log"
	"time"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

// BlockTracker is an interface to track new blocks on the chain
type BlockTracker interface {
	Track(context.Context, func(block *web3.Block) error) error
}

const (
	defaultPollInterval = 5 * time.Second
)

// JSONBlockTracker implements the BlockTracker interface using
// the http jsonrpc endpoint
type JSONBlockTracker struct {
	logger       *log.Logger
	PollInterval time.Duration
	provider     Provider
}

// NewJSONBlockTracker creates a new json block tracker
func NewJSONBlockTracker(logger *log.Logger, provider Provider) *JSONBlockTracker {
	return &JSONBlockTracker{
		logger:       logger,
		provider:     provider,
		PollInterval: defaultPollInterval,
	}
}

// Track implements the BlockTracker interface
func (k *JSONBlockTracker) Track(ctx context.Context, handle func(block *web3.Block) error) error {
	go func() {
		var lastBlock *web3.Block

		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(k.PollInterval):
				block, err := k.provider.GetBlockByNumber(web3.Latest, false)
				if err != nil {
					k.logger.Printf("[ERR]: Tracker failed to get last block: %v", err)
					continue
				}

				if lastBlock != nil && lastBlock.Hash == block.Hash {
					continue
				}

				if err := handle(block); err != nil {
					k.logger.Printf("[ERROR]: blocktracker: Failed to handle block: %v", err)
				} else {
					lastBlock = block
				}
			}
		}
	}()

	return nil
}

// SubscriptionBlockTracker is an interface to track new blocks using
// the newHeads subscription endpoint
type SubscriptionBlockTracker struct {
	logger *log.Logger
	client *jsonrpc.Client
}

// NewSubscriptionBlockTracker creates a new block tracker using the subscription endpoint
func NewSubscriptionBlockTracker(logger *log.Logger, client *jsonrpc.Client) (*SubscriptionBlockTracker, error) {
	if !client.SubscriptionEnabled() {
		return nil, fmt.Errorf("subscription is not enabled")
	}
	s := &SubscriptionBlockTracker{
		logger: logger,
		client: client,
	}
	return s, nil
}

// Track implements the BlockTracker interface
func (s *SubscriptionBlockTracker) Track(ctx context.Context, handle func(block *web3.Block) error) error {
	data := make(chan []byte)
	cancel, err := s.client.Subscribe("newHeads", func(b []byte) {
		data <- b
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case buf := <-data:
				var block web3.Block
				if err := block.UnmarshalJSON(buf); err != nil {
					s.logger.Printf("[ERR]: Tracker failed to parse web3.Block: %v", err)
				} else {
					handle(&block)
				}

			case <-ctx.Done():
				cancel()
			}
		}
	}()

	return nil
}
