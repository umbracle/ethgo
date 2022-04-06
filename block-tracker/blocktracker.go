package blocktracker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
)

// BlockProvider are the eth1x methods required by the block tracker
type BlockProvider interface {
	GetBlockByHash(hash ethgo.Hash, full bool) (*ethgo.Block, error)
	GetBlockByNumber(i ethgo.BlockNumber, full bool) (*ethgo.Block, error)
}

const (
	defaultMaxBlockBacklog = 10
)

// BlockTracker is an interface to track new blocks on the chain
type BlockTracker struct {
	config *Config

	blocks   []*ethgo.Block
	lock     sync.Mutex
	tracker  BlockTrackerInterface
	provider BlockProvider

	eventBroker *EventBroker
	closeCh     chan struct{}
}

type Config struct {
	Tracker         BlockTrackerInterface
	MaxBlockBacklog uint64
}

func DefaultConfig() *Config {
	return &Config{
		MaxBlockBacklog: defaultMaxBlockBacklog,
	}
}

type ConfigOption func(*Config)

func WithBlockMaxBacklog(b uint64) ConfigOption {
	return func(c *Config) {
		c.MaxBlockBacklog = b
	}
}

func WithTracker(b BlockTrackerInterface) ConfigOption {
	return func(c *Config) {
		c.Tracker = b
	}
}

func NewBlockTracker(provider BlockProvider, opts ...ConfigOption) *BlockTracker {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	tracker := config.Tracker
	if tracker == nil {
		tracker = NewJSONBlockTracker(log.New(os.Stderr, "", log.LstdFlags), provider)
	}

	broker, err := NewEventBroker(context.Background(), EventBrokerCfg{})
	if err != nil {
		panic(err)
	}

	initial, err := provider.GetBlockByNumber(ethgo.Latest, false)
	if err != nil {
		panic(err)
	}

	b := &BlockTracker{
		blocks:      []*ethgo.Block{},
		config:      config,
		tracker:     tracker,
		provider:    provider,
		eventBroker: broker,
		closeCh:     make(chan struct{}),
	}

	// add an initial block
	if err := b.HandleReconcile(initial); err != nil {
		panic(err)
	}
	return b
}

// Header returns the last block of the tracked chain
func (b *BlockTracker) Header() *ethgo.Block {
	b.lock.Lock()
	last := b.blocks[len(b.blocks)-1].Copy()
	b.lock.Unlock()
	return last
}

func (b *BlockTracker) Subscribe() *Subscription {
	return b.eventBroker.Subscribe()
}

func (b *BlockTracker) Close() error {
	close(b.closeCh)
	return nil
}

func (b *BlockTracker) Start() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-b.closeCh
		cancelFn()
	}()
	// start the polling
	err := b.tracker.Track(ctx, func(block *ethgo.Block) error {
		return b.HandleReconcile(block)
	})
	if err != nil {
		return err
	}
	return err
}

func (t *BlockTracker) AddBlocks(block *ethgo.Block) error {
	if uint64(len(t.blocks)) == t.config.MaxBlockBacklog {
		// remove past blocks if there are more than maxReconcileBlocks
		t.blocks = t.blocks[1:]
	}
	if len(t.blocks) != 0 {
		lastNum := t.blocks[len(t.blocks)-1].Number
		if lastNum+1 != block.Number {
			return fmt.Errorf("bad number sequence. %d and %d", lastNum, block.Number)
		}
	}
	t.blocks = append(t.blocks, block)
	return nil
}

func (t *BlockTracker) blockAtIndex(hash ethgo.Hash) int {
	for indx, b := range t.blocks {
		if b.Hash == hash {
			return indx
		}
	}
	return -1
}

func (t *BlockTracker) handleReconcileImpl(block *ethgo.Block) ([]*ethgo.Block, int, error) {
	// The block already exists
	if t.blockAtIndex(block.Hash) != -1 {
		return nil, -1, nil
	}

	// The state is empty
	if len(t.blocks) == 0 {
		return []*ethgo.Block{block}, -1, nil
	}

	// Append to the head of the chain
	if t.blocks[len(t.blocks)-1].Hash == block.ParentHash {
		return []*ethgo.Block{block}, -1, nil
	}

	// Fork in the middle of the chain
	if indx := t.blockAtIndex(block.ParentHash); indx != -1 {
		return []*ethgo.Block{block}, indx, nil
	}

	// Backfill. We dont know the parent of the block.
	// Need to query the chain untill we find a known block

	added := []*ethgo.Block{block}
	var indx int

	count := uint64(0)
	for {
		if count > t.config.MaxBlockBacklog {
			return nil, -1, fmt.Errorf("cannot reconcile more than '%d' max backlog values", t.config.MaxBlockBacklog)
		}
		count++

		parent, err := t.provider.GetBlockByHash(block.ParentHash, false)
		if err != nil {
			return nil, -1, fmt.Errorf("parent with hash %s not found", block.ParentHash)
		}

		added = append(added, parent)
		if indx = t.blockAtIndex(parent.ParentHash); indx != -1 {
			break
		}
		block = parent
	}

	// need the blocks in reverse order
	blocks := []*ethgo.Block{}
	for i := len(added) - 1; i >= 0; i-- {
		blocks = append(blocks, added[i])
	}
	return blocks, indx, nil
}

func (t *BlockTracker) HandleBlockEvent(block *ethgo.Block) (*BlockEvent, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	blocks, indx, err := t.handleReconcileImpl(block)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, nil
	}

	blockEvnt := &BlockEvent{}

	// there are some blocks to remove
	if indx != -1 {
		for i := indx + 1; i < len(t.blocks); i++ {
			blockEvnt.Removed = append(blockEvnt.Removed, t.blocks[i])
		}
		t.blocks = t.blocks[:indx+1]
	}

	// include the new blocks
	for _, block := range blocks {
		blockEvnt.Added = append(blockEvnt.Added, block)
		if err := t.AddBlocks(block); err != nil {
			return nil, err
		}
	}
	return blockEvnt, nil
}

func (t *BlockTracker) HandleReconcile(block *ethgo.Block) error {
	blockEvnt, err := t.HandleBlockEvent(block)
	if err != nil {
		return err
	}
	if blockEvnt == nil {
		return nil
	}

	t.eventBroker.Publish(blockEvnt)
	return nil
}

type BlockTrackerInterface interface {
	Track(context.Context, func(block *ethgo.Block) error) error
}

const (
	defaultPollInterval = 1 * time.Second
)

// JSONBlockTracker implements the BlockTracker interface using
// the http jsonrpc endpoint
type JSONBlockTracker struct {
	logger       *log.Logger
	PollInterval time.Duration
	provider     BlockProvider
}

// NewJSONBlockTracker creates a new json block tracker
func NewJSONBlockTracker(logger *log.Logger, provider BlockProvider) *JSONBlockTracker {
	return &JSONBlockTracker{
		logger:       logger,
		provider:     provider,
		PollInterval: defaultPollInterval,
	}
}

// Track implements the BlockTracker interface
func (k *JSONBlockTracker) Track(ctx context.Context, handle func(block *ethgo.Block) error) error {
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
func (s *SubscriptionBlockTracker) Track(ctx context.Context, handle func(block *ethgo.Block) error) error {
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

// BlockEvent is an event emitted when a new block is included
type BlockEvent struct {
	Added   []*ethgo.Block
	Removed []*ethgo.Block
}

func (b *BlockEvent) Header() *ethgo.Block {
	return b.Added[len(b.Added)-1]
}
