package blocktracker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/umbracle/ethgo"
)

const (
	defaultMaxBlockBacklog = 10
)

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

// BlockProvider are the eth1x methods required by the block tracker
type BlockProvider interface {
	GetBlockByHash(hash ethgo.Hash, full bool) (*ethgo.Block, error)
	GetBlockByNumber(i ethgo.BlockNumber, full bool) (*ethgo.Block, error)
}

// BlockEvent is an event emitted when a new block is included
type BlockEvent struct {
	Added   []*ethgo.Block
	Removed []*ethgo.Block
}

func (b *BlockEvent) Header() *ethgo.Block {
	if len(b.Added) == 0 {
		return nil
	}
	return b.Added[len(b.Added)-1]
}

// BlockTracker is an interface to track new blocks on the chain
type BlockTracker struct {
	config *Config

	// blocks is the list of historical blocks
	blocks []*ethgo.Block

	// blocksLock is the lock to access 'blocks'
	lock sync.Mutex

	// stream handles a lock-free stream of blocks
	stream *blockStream

	// sub is the local subscription of the blocktracker
	sub *subscription

	// headTracker tracks the head of the chain
	headTracker BlockTrackerInterface

	// provider is a reference to the Ethereum API (JsonRPC)
	provider BlockProvider
}

func NewBlockTracker(provider BlockProvider, opts ...ConfigOption) (*BlockTracker, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	tracker := config.Tracker
	if tracker == nil {
		tracker = NewJSONHeadTracker(log.New(os.Stderr, "", log.LstdFlags), provider)
	}

	b := &BlockTracker{
		blocks:      []*ethgo.Block{},
		config:      config,
		headTracker: tracker,
		provider:    provider,
		stream:      newBlockStream(),
	}

	// add an initial block
	initial, err := provider.GetBlockByNumber(ethgo.Latest, false)
	if err != nil {
		return nil, err
	}
	if err := b.HandleReconcile(initial); err != nil {
		return nil, err
	}

	// create the subscription
	b.sub = b.subscribe()

	return b, nil
}

// Header returns the last block of the tracked chain
func (b *BlockTracker) Header() *ethgo.Block {
	b.lock.Lock()
	last := b.blocks[len(b.blocks)-1].Copy()
	b.lock.Unlock()
	return last
}

func (b *BlockTracker) Start(ctx context.Context) error {
	// start the polling
	err := b.headTracker.Track(ctx, func(block *ethgo.Block) error {
		return b.HandleReconcile(block)
	})
	if err != nil {
		return err
	}
	return err
}

func (t *BlockTracker) blockAtIndex(hash ethgo.Hash) int {
	for indx, b := range t.blocks {
		if b.Hash == hash {
			return indx
		}
	}
	return -1
}

func (t *BlockTracker) addBlocks(block *ethgo.Block) error {
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
		if err := t.addBlocks(block); err != nil {
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

	t.stream.push(blockEvnt)
	return nil
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

func (b *BlockTracker) subscribe() *subscription {
	return &subscription{last: b.stream.Head()}
}

func (b *BlockTracker) Flush() *BlockEvent {
	return b.sub.Flush()
}

func (b *BlockTracker) Next(ctx context.Context) (*BlockEvent, error) {
	return b.sub.Next(ctx)
}

type subscription struct {
	last *headElem
}

func (s *subscription) Flush() *BlockEvent {
	elem := s.last.flush()
	s.last = elem
	return elem.event
}

func (s *subscription) Next(ctx context.Context) (*BlockEvent, error) {
	elem, err := s.last.next(ctx)
	if err != nil {
		return nil, err
	}
	s.last = elem
	return elem.event, nil
}
