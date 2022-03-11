package tracker

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/umbracle/ethgo"
	blocktracker "github.com/umbracle/ethgo/block-tracker"
	"github.com/umbracle/ethgo/etherscan"
	"github.com/umbracle/ethgo/jsonrpc/codec"
)

// FilterConfig is a tracker filter configuration
type FilterConfig struct {
	Address []ethgo.Address
	Topics  [][]*ethgo.Hash
}

func (f *FilterConfig) getFilterSearch() *ethgo.LogFilter {
	filter := &ethgo.LogFilter{}
	if len(f.Address) != 0 {
		filter.Address = f.Address
	}
	if len(f.Topics) != 0 {
		filter.Topics = f.Topics
	}
	return filter
}

// Provider are the eth1x methods required by the tracker
type Provider interface {
	BlockNumber() (uint64, error)
	GetBlockByHash(hash ethgo.Hash, full bool) (*ethgo.Block, error)
	GetBlockByNumber(i ethgo.BlockNumber, full bool) (*ethgo.Block, error)
	GetLogs(filter *ethgo.LogFilter) ([]*ethgo.Log, error)
	ChainID() (*big.Int, error)
}

// Tracker is a contract event tracker
type Tracker struct {
	logger   *log.Logger
	provider Provider
	config   *Config
	entry    Entry
	synced   int32
	DoneCh   chan struct{}
}

// NewTracker creates a new tracker
func NewTracker(provider Provider, opts ...ConfigOption) (*Tracker, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if config.BlockTracker == nil {
		// create a block tracker with the provider
		config.BlockTracker = blocktracker.NewBlockTracker(provider, blocktracker.WithBlockMaxBacklog(100))
	}

	t := &Tracker{
		provider: provider,
		config:   config,
		logger:   config.Logger,
		DoneCh:   make(chan struct{}, 1),
		synced:   0,
		entry:    config.Entry,
	}
	return t, nil
}

// IsSynced returns true if the filter is synced to head
func (t *Tracker) IsSynced() bool {
	return atomic.LoadInt32(&t.synced) != 0
}

func (t *Tracker) findAncestor(block, pivot *ethgo.Block) (uint64, error) {
	// block is part of a fork that is not the current head, find a common ancestor
	// both block and pivot are at the same height
	var err error

	for i := uint64(0); i < t.config.MaxBacklog; i++ {
		if block.Number != pivot.Number {
			return 0, fmt.Errorf("block numbers do not match")
		}
		if block.Hash == pivot.Hash {
			// this is the common ancestor in both
			return block.Number, nil
		}
		block, err = t.provider.GetBlockByHash(block.ParentHash, false)
		if err != nil {
			return 0, err
		}
		pivot, err = t.provider.GetBlockByHash(pivot.ParentHash, false)
		if err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("the reorg is bigger than maxBlockBacklog %d", t.config.MaxBacklog)
}

func tooMuchDataRequestedError(err error) bool {
	obj, ok := err.(*codec.ErrorObject)
	if !ok {
		return false
	}
	if obj.Message == "query returned more than 10000 results" {
		return true
	}
	return false
}

func (t *Tracker) syncBatch(ctx context.Context, from, to uint64) error {
	if to < from {
		panic(fmt.Sprintf("BUG sync batch: (%d, %d)", from, to))
	}

	query := t.config.Filter.getFilterSearch()

	batchSize := t.config.BatchSize
	additiveFactor := uint64(float64(batchSize) * 0.10)

	i := from

START:
	dst := min(to, i+batchSize)

	query.SetFromUint64(i)
	query.SetToUint64(dst)

	logs, err := t.provider.GetLogs(query)
	if err != nil {
		if tooMuchDataRequestedError(err) {
			// multiplicative decrease
			batchSize = batchSize / 2
			goto START
		}
		return err
	}

	// update the last block entry
	block, err := t.provider.GetBlockByNumber(ethgo.BlockNumber(dst), false)
	if err != nil {
		return err
	}

	// add logs to the store
	evnt := &Event{Added: logs, Indx: -1}
	evnt.Block = block

	if err := t.entry.StoreEvent(evnt); err != nil {
		return err
	}

	// check if the execution is over after each query batch
	if err := ctx.Err(); err != nil {
		return err
	}

	i += batchSize + 1

	// update the batchSize with additive increase
	if batchSize < t.config.BatchSize {
		batchSize = min(t.config.BatchSize, batchSize+additiveFactor)
	}

	if i <= to {
		goto START
	}
	return nil
}

func (t *Tracker) fastTrack(filterConfig *FilterConfig) (*ethgo.Block, error) {
	// Try to use first the user provided block if any
	if t.config.StartBlock != 0 {
		bb, err := t.provider.GetBlockByNumber(ethgo.BlockNumber(t.config.StartBlock), false)
		if err != nil {
			return nil, err
		}
		return bb, nil
	}

	// Only possible if we filter addresses
	if len(filterConfig.Address) == 0 {
		return nil, nil
	}

	if t.config.EtherscanAPIKey != "" {
		chainID, err := t.provider.ChainID()
		if err != nil {
			return nil, err
		}

		// get the etherscan instance for this chainID
		e, err := etherscan.NewEtherscanFromNetwork(ethgo.Network(chainID.Uint64()), t.config.EtherscanAPIKey)
		if err != nil {
			// there is no etherscan api for this specific chainid
			return nil, nil
		}

		getAddress := func(addr ethgo.Address) (uint64, error) {
			params := map[string]string{
				"address":   addr.String(),
				"fromBlock": "0",
				"toBlock":   "latest",
			}
			var out []map[string]interface{}
			if err := e.Query("logs", "getLogs", &out, params); err != nil {
				return 0, err
			}
			if len(out) == 0 {
				return 0, nil
			}

			cc, ok := out[0]["blockNumber"].(string)
			if !ok {
				return 0, fmt.Errorf("failed to cast blocknumber")
			}

			num, err := parseUint64orHex(cc)
			if err != nil {
				return 0, err
			}
			return num, nil
		}

		minBlock := ^uint64(0) // max uint64
		for _, addr := range filterConfig.Address {
			num, err := getAddress(addr)
			if err != nil {
				return nil, err
			}
			if num < minBlock {
				minBlock = num
			}
		}

		bb, err := t.provider.GetBlockByNumber(ethgo.BlockNumber(minBlock-1), false)
		if err != nil {
			return nil, err
		}
		return bb, nil
	}

	return nil, nil
}

func (t *Tracker) BatchSync(ctx context.Context) error {
	if t.config.BlockTracker == nil {
		// run a specfic block tracker
		t.config.BlockTracker = blocktracker.NewBlockTracker(t.provider, blocktracker.WithBlockMaxBacklog(t.config.MaxBacklog))
		go t.config.BlockTracker.Start()

		go func() {
			// track our stop
			<-ctx.Done()
			t.config.BlockTracker.Close()
		}()
	}

	if err := t.syncImpl(ctx); err != nil {
		return err
	}

	select {
	case t.DoneCh <- struct{}{}:
	default:
	}

	atomic.StoreInt32(&t.synced, 1)
	return nil
}

// Sync syncs a specific filter
func (t *Tracker) Sync(ctx context.Context) error {
	if err := t.BatchSync(ctx); err != nil {
		return err
	}

	/*
		// subscribe and sync
		ch := t.blockSub.GetEventCh()

		go func() {
			for {
				select {
				case evnt := <-ch:
					t.handleBlockEvnt(evnt)
				case <-ctx.Done():
					return
				}
			}
		}()
	*/

	return nil
}

func (t *Tracker) syncImpl(ctx context.Context) error {
	// get the current target
	headBlock := t.config.BlockTracker.Header()
	if headBlock == nil {
		return nil
	}
	headNum := headBlock.Number

	last, err := t.entry.GetLastBlock()
	if err != nil {
		return err
	}
	if last == nil {
		// Fast track to an initial block (if possible)
		last, err = t.fastTrack(t.config.Filter)
		if err != nil {
			return fmt.Errorf("failed to fast track initial block: %v", err)
		}
	} else {
		if last.Hash == headBlock.Hash {
			return nil
		}
	}

	// First it needs to figure out if there was a reorg just at the
	// stopping point of the last execution (if any). Check that our
	// last processed block ('beacon') hash matches the canonical one
	// in the chain. Otherwise, figure out the common ancestor up to
	// 'beacon' - maxBackLog, set that as our real origin and remove
	// any logs from the store.

	var origin uint64
	if last != nil {
		if last.Number > headNum {
			return fmt.Errorf("store '%d' is more advanced than the head chain block '%d'", last.Number, headNum)
		}

		pivot, err := t.provider.GetBlockByNumber(ethgo.BlockNumber(last.Number), false)
		if err != nil {
			return err
		}

		if last.Number == headNum {
			origin = last.Number
		} else {
			origin = last.Number + 1
		}

		if pivot.Hash != last.Hash {
			ancestor, err := t.findAncestor(last, pivot)
			if err != nil {
				return err
			}

			origin = ancestor + 1
			_, indx, err := t.removeLogs(ancestor+1, nil)
			if err != nil {
				return err
			}
			if err := t.entry.StoreEvent(&Event{Indx: int64(indx)}); err != nil {
				return err
			}
		}
	}

	if headNum-origin+1 > t.config.MaxBacklog {
		// The tracker is far (more than maxBackLog) from the canonical head.
		// Do a bulk sync with the eth_getLogs endpoint and get closer to the target.

		for {
			if origin > headNum {
				return fmt.Errorf("from (%d) higher than to (%d)", origin, headNum)
			}
			if headNum-origin+1 <= t.config.MaxBacklog {
				// Already in reorg range
				break
			}

			target := headNum - t.config.MaxBacklog
			if err := t.syncBatch(ctx, origin, target); err != nil {
				return err
			}

			origin = target + 1

			// Reset the canonical head since it could have moved during the batch logs
			headNum = t.config.BlockTracker.Header().Number
		}
	}

	// At this point we are either:
	// 1. At 'canonical head' - maxBackLog if batch sync was done.
	// 2. Inside maxBackLog range if our last processed block was close to the head.
	// In both cases, the variable 'origin' indicates the last block processed.
	// Now we fill the rest of the blocks till the block head using as a reference
	// the block tracker subscription. After that, we can use the same subscription
	// reference to start the watch.
	// It is important to fill these blocks using block hashes and the block chain
	// parent hash references since we are in reorgs range.

	sub := t.config.BlockTracker.Subscribe()

	evnt, err := sub.Next(context.Background())
	if err != nil {
		panic(err)
	}

	// we include the first header from the subscription too.
	// TODO: HOW DOES THE SUBSCRIPTION WORKS NOW? TEST IT.
	header := evnt.Header()
	added := []*ethgo.Block{header}

	for header.Number != origin {
		header, err = t.provider.GetBlockByHash(header.ParentHash, false)
		if err != nil {
			return err
		}
		added = append(added, header)
	}

	if len(added) == 0 {
		return nil
	}

	// we need to reverse the blocks since they were included in descending order
	// and we need to process them in ascending order.
	added = reverseBlocks(added)

	if _, err := t.handleBlockEvent(&blocktracker.BlockEvent{Added: added}); err != nil {
		return err
	}

	return nil
}

func (t *Tracker) removeLogs(number uint64, hash *ethgo.Hash) ([]*ethgo.Log, uint64, error) {
	index, err := t.entry.LastIndex()
	if err != nil {
		return nil, 0, err
	}
	if index == 0 {
		return nil, 0, nil
	}

	var remove []*ethgo.Log
	for {
		elemIndex := index - 1

		var log ethgo.Log
		if err := t.entry.GetLog(elemIndex, &log); err != nil {
			return nil, 0, err
		}
		if log.BlockNumber == number {
			if hash != nil && log.BlockHash != *hash {
				break
			}
		}
		if log.BlockNumber < number {
			break
		}
		remove = append(remove, &log)
		if elemIndex == 0 {
			index = 0
			break
		}
		index = elemIndex
	}

	return remove, index, nil
}

func reverseBlocks(in []*ethgo.Block) (out []*ethgo.Block) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

func reverseLogs(in []*ethgo.Log) (out []*ethgo.Log) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

func (t *Tracker) handleBlockEvent(blockEvnt *blocktracker.BlockEvent) (*Event, error) {
	evnt := &Event{
		Indx:    -1,
		Added:   []*ethgo.Log{},
		Removed: []*ethgo.Log{},
	}
	if len(blockEvnt.Removed) != 0 {
		pivot := blockEvnt.Removed[0]
		logs, index, err := t.removeLogs(pivot.Number, &pivot.Hash)
		if err != nil {
			return nil, err
		}
		evnt.Indx = int64(index)
		evnt.Removed = append(evnt.Removed, reverseLogs(logs)...)
	}

	for _, block := range blockEvnt.Added {
		// check logs for this blocks
		query := t.config.Filter.getFilterSearch()
		query.BlockHash = &block.Hash

		// We check the hash, we need to do a retry to let unsynced nodes get the block
		var logs []*ethgo.Log
		var err error

		for i := 0; i < 5; i++ {
			logs, err = t.provider.GetLogs(query)
			if err == nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		if err != nil {
			return nil, err
		}
		evnt.Added = append(evnt.Added, logs...)
	}

	evnt.Block = blockEvnt.Added[len(blockEvnt.Added)-1]

	// store the event in the store
	if err := t.entry.StoreEvent(evnt); err != nil {
		return nil, err
	}
	return evnt, nil
}

// EventType is the type of the event (TODO: REMOVE)
type EventType int

const (
	// EventAdd happens when a new event is included in the chain
	EventAdd EventType = iota
	// EventDel may happen when there is a reorg and a past event is deleted
	EventDel
)

// Event is an event emitted when a new log is included
type Event struct {
	Type    EventType
	Added   []*ethgo.Log
	Removed []*ethgo.Log
	Indx    int64
	Block   *ethgo.Block
}

// BlockEvent is an event emitted when a new block is included
type BlockEvent struct {
	Type    EventType
	Added   []*ethgo.Block
	Removed []*ethgo.Block
}

func min(i, j uint64) uint64 {
	if i < j {
		return i
	}
	return j
}

func parseUint64orHex(str string) (uint64, error) {
	base := 10
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
		base = 16
	}
	return strconv.ParseUint(str, base, 64)
}
