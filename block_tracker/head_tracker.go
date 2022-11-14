package blocktracker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
)

type BlockTrackerInterface interface {
	Track(context.Context, func(block *ethgo.Block) error) error
}

const (
	defaultPollInterval = 1 * time.Second
)

// JSONHeadTracker implements the BlockTracker interface using
// the http jsonrpc endpoint
type JSONHeadTracker struct {
	logger       *log.Logger
	PollInterval time.Duration
	provider     BlockProvider
}

// NewJSONHeadTracker creates a new json block tracker
func NewJSONHeadTracker(logger *log.Logger, provider BlockProvider) *JSONHeadTracker {
	return &JSONHeadTracker{
		logger:       logger,
		provider:     provider,
		PollInterval: defaultPollInterval,
	}
}

// Track implements the BlockTracker interface
func (k *JSONHeadTracker) Track(ctx context.Context, handle func(block *ethgo.Block) error) error {
	go func() {
		var lastBlock *ethgo.Block

		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(k.PollInterval):
				block, err := k.provider.GetBlockByNumber(ethgo.Latest, false)
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

// SubscriptionHeadTracker is an interface to track new blocks using
// the newHeads subscription endpoint
type SubscriptionHeadTracker struct {
	logger *log.Logger
	client *jsonrpc.Client
}

// NewSubscriptionHeadTracker creates a new block tracker using the subscription endpoint
func NewSubscriptionHeadTracker(logger *log.Logger, client *jsonrpc.Client) (*SubscriptionHeadTracker, error) {
	if !client.SubscriptionEnabled() {
		return nil, fmt.Errorf("subscription is not enabled")
	}
	s := &SubscriptionHeadTracker{
		logger: logger,
		client: client,
	}
	return s, nil
}

// Track implements the BlockTracker interface
func (s *SubscriptionHeadTracker) Track(ctx context.Context, handle func(block *ethgo.Block) error) error {
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
				var block ethgo.Block
				if err := block.UnmarshalJSON(buf); err != nil {
					s.logger.Printf("[ERR]: Tracker failed to parse ethgo.Block: %v", err)
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
