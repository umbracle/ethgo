package tracker

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/etherscan"
	"github.com/umbracle/go-web3/jsonrpc/codec"
)

const (
	maxReconcileBlocks = 10
)

// Config is the configuration of the tracker
type Config struct {
	BatchSize          uint64
	MaxBlockBacklog    uint64
	PollInterval       time.Duration
	EtherscanFastTrack bool
}

// DefaultConfig returns the default tracker config
func DefaultConfig() *Config {
	return &Config{
		BatchSize:          100,
		MaxBlockBacklog:    10,
		PollInterval:       5 * time.Second,
		EtherscanFastTrack: false,
	}
}

// Provider are the eth1x methods required by the tracker
type Provider interface {
	BlockNumber() (uint64, error)
	GetBlockByHash(hash web3.Hash, full bool) (*web3.Block, error)
	GetBlockByNumber(i web3.BlockNumber, full bool) (*web3.Block, error)
	GetLogs(filter *web3.LogFilter) ([]*web3.Log, error)
	ChainID() (*big.Int, error)
}

// Tracker is a contract event tracker
type Tracker struct {
	provider Provider
	config   *Config
	store    Store
	EventCh  chan *Event
	blocks   []*web3.Block

	// SyncCh is used to notify of sync events
	SyncCh chan uint64

	// filter
	topics  []*web3.Hash
	address *web3.Address
}

// NewTracker creates a new tracker
func NewTracker(provider Provider, config *Config) *Tracker {
	return &Tracker{
		provider: provider,
		config:   config,
		blocks:   []*web3.Block{},
	}
}

// SetStore sets the store
func (t *Tracker) SetStore(store Store) {
	t.store = store
}

// SetFilterAddress sets the filter address for the tracker
func (t *Tracker) SetFilterAddress(addr web3.Address) {
	t.address = &addr
}

// SetFilterTopics sets the filter topics for the tracker
func (t *Tracker) SetFilterTopics(topics []*web3.Hash) {
	t.topics = topics
}

// SyncAsync syncs asyncronously
func (t *Tracker) SyncAsync(ctx context.Context) chan error {
	err := make(chan error, 1)
	go func() {
		err <- t.Sync(ctx)
	}()
	return err
}

func (t *Tracker) findAncestor(block, pivot *web3.Block) (uint64, error) {
	// block is part of a fork that is not the current head, find a common ancestor
	// both block and pivot are at the same height
	var err error

	for i := uint64(0); i < t.config.MaxBlockBacklog; i++ {
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
	return 0, fmt.Errorf("the reorg is bigger than maxBlockBacklog %d", t.config.MaxBlockBacklog)
}

func (t *Tracker) fillBlocksCache(last *web3.Block, num uint64) error {
	parent := last
	num = min(num, last.Number)

	t.blocks = make([]*web3.Block, num)
	t.blocks[num-1] = parent

	// fill the rest of the values
	for i := uint64(2); i < num+1; i++ {
		block, err := t.provider.GetBlockByHash(parent.ParentHash, false)
		if err != nil {
			return err
		}

		t.blocks[num-i] = block
		parent = block
	}
	return nil
}

func (t *Tracker) emitLogs(typ EventType, logs []*web3.Log) {
	evnt := &Event{}
	if typ == EventAdd {
		evnt.AddedLogs = logs
	}
	if typ == EventDel {
		evnt.RemovedLogs = logs
	}

	select {
	case t.EventCh <- evnt:
	default:
	}
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
	filter := t.getFilter()

	batchSize := t.config.BatchSize
	additiveFactor := uint64(float64(batchSize) * 0.10)

	i := from

START:
	dst := min(to, i+batchSize)

	filter.SetFromUint64(i)
	filter.SetToUint64(dst)

	logs, err := t.provider.GetLogs(filter)
	if err != nil {
		if tooMuchDataRequestedError(err) {
			// multiplicative decrease
			batchSize = batchSize / 2
			goto START
		}
		return err
	}

	if t.SyncCh != nil {
		select {
		case t.SyncCh <- dst:
		default:
		}
	}

	// add logs to the store
	if err := t.store.StoreLogs(logs); err != nil {
		return err
	}
	t.emitLogs(EventAdd, logs)

	// update the last block entry
	block, err := t.provider.GetBlockByNumber(web3.BlockNumber(dst), false)
	if err != nil {
		return err
	}
	if err := t.storeLastBlock(block); err != nil {
		return err
	}

	// check if the execution is over after each query batch
	if err := ctx.Err(); err != nil {
		return err
	}

	// update the batchSize with additive increase
	if batchSize < t.config.BatchSize {
		batchSize = min(t.config.BatchSize, batchSize+additiveFactor)
	}

	i += batchSize + 1
	if i <= to {
		goto START
	}
	return nil
}

func (t *Tracker) syncBulk(ctx context.Context, from, to uint64) (uint64, uint64, error) {
	var err error
	for {
		if from > to {
			return 0, 0, fmt.Errorf("from (%d) higher than to (%d)", from, to)
		}

		if to-from+1 <= t.config.MaxBlockBacklog {
			return from, to, nil
		}

		limit := to - t.config.MaxBlockBacklog
		if err := t.syncBatch(ctx, from, limit); err != nil {
			return 0, 0, err
		}

		from = limit + 1
		if to, err = t.provider.BlockNumber(); err != nil {
			return 0, 0, err
		}
	}
}

func (t *Tracker) preSyncCheck() error {
	rGenesis, err := t.provider.GetBlockByNumber(0, false)
	if err != nil {
		return err
	}
	rChainID, err := t.provider.ChainID()
	if err != nil {
		return err
	}

	genesis, err := t.store.Get(dbGenesis)
	if err != nil {
		return err
	}
	chainID, err := t.store.Get(dbChainID)
	if err != nil {
		return err
	}
	if len(genesis) != 0 {
		if !bytes.Equal(genesis, rGenesis.Hash[:]) {
			return fmt.Errorf("bad genesis")
		}
		if !bytes.Equal(chainID, rChainID.Bytes()) {
			return fmt.Errorf("bad genesis")
		}
	} else {
		if err := t.store.Set(dbGenesis, rGenesis.Hash[:]); err != nil {
			return err
		}
		if err := t.store.Set(dbChainID, rChainID.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// getFilter returns the filter query
func (t *Tracker) getFilter() *web3.LogFilter {
	filter := &web3.LogFilter{}
	if t.address != nil {
		filter.Address = []web3.Address{*t.address}
	}
	if t.topics != nil {
		filter.Topics = t.topics
	}
	return filter
}

func (t *Tracker) fastTrack() (*web3.Block, error) {
	// Only possible if we filter addresses
	if t.address == nil {
		return nil, nil
	}

	if t.config.EtherscanFastTrack {
		chainID, err := t.provider.ChainID()
		if err != nil {
			return nil, err
		}

		// get the etherscan instance for this chainID
		e, err := etherscan.NewEtherscanFromNetwork(web3.Network(chainID.Uint64()))
		if err != nil {
			// there is no etherscan api for this specific chainid
			return nil, nil
		}

		params := map[string]string{
			"address":   t.address.String(),
			"fromBlock": "0",
			"toBlock":   "latest",
		}
		var out []map[string]interface{}
		if err := e.Query("logs", "getLogs", &out, params); err != nil {
			return nil, err
		}
		if len(out) == 0 {
			return nil, nil
		}

		cc, ok := out[0]["blockNumber"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to cast blocknumber")
		}

		num, err := parseUint64orHex(cc)
		if err != nil {
			return nil, err
		}
		bb, err := t.provider.GetBlockByNumber(web3.BlockNumber(num-1), false)
		if err != nil {
			return nil, err
		}
		return bb, nil
	}

	return nil, nil
}

// Sync syncs the historical data
func (t *Tracker) Sync(ctx context.Context) error {
	// do some preflight checks
	if err := t.preSyncCheck(); err != nil {
		return err
	}

	// take the last number and search those values
	// After its done, check if the number has changed something
	// If thats the case just parse it again to the new values

	target, err := t.provider.BlockNumber()
	if err != nil {
		return err
	}
	if target == 0 {
		return nil
	}
	last, err := t.GetLastBlock()
	if err != nil {
		return err
	}

	if last == nil {
		// Try to fast track to the valid block (if possible)
		last, err = t.fastTrack()
		if err != nil {
			return fmt.Errorf("failed to fasttrack: %v", err)
		}
		if last != nil {
			if err := t.storeLastBlock(last); err != nil {
				return err
			}
		}
	}

	// There might been a reorg when we stopped syncing last time,
	// check that our 'beacon' block matches the one in the chain.
	// If that is not the case, we consider beacon-maxBackLog our
	// real origin point and remove any logs ahead of that point.

	var origin uint64
	if last != nil {
		if last.Number > target {
			return fmt.Errorf("store is more advanced than the chain")
		}

		pivot, err := t.provider.GetBlockByNumber(web3.BlockNumber(last.Number), false)
		if err != nil {
			return err
		}

		if last.Number == target {
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
			logs, err := t.removeLogs(ancestor+1, nil)
			if err != nil {
				return err
			}
			t.emitLogs(EventDel, logs)

			last, err = t.provider.GetBlockByNumber(web3.BlockNumber(ancestor), false)
			if err != nil {
				return err
			}
		}
	}

	step := target - origin + 1
	if step > t.config.MaxBlockBacklog {
		// bulk sync the logs
		origin, target, err = t.syncBulk(ctx, origin, target)
		if err != nil {
			return err
		}
	} else if step < t.config.MaxBlockBacklog {
		// how sure are we that this only works with the last value set??
		if last != nil {
			num := uint64(int(t.config.MaxBlockBacklog) - int(target-last.Number))
			if err := t.fillBlocksCache(last, num); err != nil {
				return err
			}
		}
	}

	// prev has the last batch block, advance one for the correct position

	// Start the specific sync of the last maxBacklog values.
	for i := origin; i <= target; i++ {
		block, err := t.provider.GetBlockByNumber(web3.BlockNumber(i), false)
		if err != nil {
			return err
		}

		evnt, err := t.handleReconcile(block)
		if err != nil {
			return err
		}

		if evnt != nil {
			select {
			case t.EventCh <- evnt:
			default:
			}
		}

		err = ctx.Err()
		if i == target || err != nil {
			// Update the store last block entry
			if err := t.storeLastBlock(block); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Polling starts the polling of the chain, it should be run after sync
// to populate block cache and avoid duplicate values
func (t *Tracker) Polling(ctx context.Context) {
	go func() {
		var lastBlock *web3.Block

		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(t.config.PollInterval):
				block, err := t.provider.GetBlockByNumber(web3.Latest, false)
				if err != nil {
					continue
				}

				if lastBlock != nil && lastBlock.Hash == block.Hash {
					continue
				}

				lastBlock = block
				evnt, err := t.handleReconcile(block)
				if err != nil {
					// log
					continue
				}

				if evnt != nil {
					select {
					case t.EventCh <- evnt:
					default:
					}
				}
			}
		}
	}()
}

var (
	lastBlock = []byte("lastBlock")
)

// GetLastBlock returns the last block processed by the tracker
func (t *Tracker) GetLastBlock() (*web3.Block, error) {
	buf, err := t.store.Get(lastBlock)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, nil
	}
	b := &web3.Block{}
	if err := b.UnmarshalJSON(buf); err != nil {
		return nil, err
	}
	return b, nil
}

func (t *Tracker) storeLastBlock(b *web3.Block) error {
	if b.Difficulty == nil {
		b.Difficulty = big.NewInt(0)
	}
	buf, err := b.MarshalJSON()
	if err != nil {
		return err
	}
	return t.store.Set(lastBlock, buf)
}

func (t *Tracker) addBlock(block *web3.Block) error {
	if len(t.blocks) == maxReconcileBlocks {
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

func (t *Tracker) blockAtIndex(hash web3.Hash) int {
	for indx, b := range t.blocks {
		if b.Hash == hash {
			return indx
		}
	}
	return -1
}

func (t *Tracker) removeLogs(number uint64, hash *web3.Hash) ([]*web3.Log, error) {
	index, err := t.store.LastIndex()
	if err != nil {
		return nil, err
	}
	if index == 0 {
		return nil, nil
	}

	var remove []*web3.Log
	for {
		elemIndex := index - 1

		var log web3.Log
		if err := t.store.GetLog(elemIndex, &log); err != nil {
			return nil, err
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

	if err := t.store.RemoveLogs(index); err != nil {
		return nil, err
	}
	return remove, nil
}

func revertLogs(in []*web3.Log) (out []*web3.Log) {
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return
}

func (t *Tracker) handleReconcile(block *web3.Block) (*Event, error) {
	blocks, indx, err := t.handleReconcileImpl(block)
	if err != nil {
		return nil, err
	}
	// nothing to do
	if len(blocks) == 0 {
		return nil, nil
	}

	evnt := &Event{}
	if indx != -1 {
		// there are some blocks to remove
		for i := indx + 1; i < len(t.blocks); i++ {
			evnt.Removed = append(evnt.Removed, t.blocks[i])
		}

		if len(evnt.Removed) != 0 {
			pivot := evnt.Removed[0]
			logs, err := t.removeLogs(pivot.Number, &pivot.Hash)
			if err != nil {
				return nil, err
			}
			evnt.removeLogs(revertLogs(logs))
		}

		// check if we have to remove any logs from here
		t.blocks = t.blocks[:indx+1]
	}

	for _, block := range blocks {
		evnt.addBlock(block)
		if err := t.addBlock(block); err != nil {
			return nil, err
		}

		// check logs for this blocks
		filter := t.getFilter()
		filter.BlockHash = &block.Hash

		logs, err := t.provider.GetLogs(filter)
		if err != nil {
			return nil, err
		}

		// add logs to the store
		if err := t.store.StoreLogs(logs); err != nil {
			return nil, err
		}
		// add logs to the event
		evnt.addLogs(logs)
	}

	// store the last block as the new index
	if err := t.storeLastBlock(blocks[len(blocks)-1]); err != nil {
		return nil, err
	}
	return evnt, nil
}

func (t *Tracker) handleReconcileImpl(block *web3.Block) ([]*web3.Block, int, error) {
	// The block already exists
	if t.blockAtIndex(block.Hash) != -1 {
		return nil, -1, nil
	}

	// The state is empty
	if len(t.blocks) == 0 {
		return []*web3.Block{block}, -1, nil
	}

	// Append to the head of the chain
	if t.blocks[len(t.blocks)-1].Hash == block.ParentHash {
		return []*web3.Block{block}, -1, nil
	}

	// Fork in the middle of the chain
	if indx := t.blockAtIndex(block.ParentHash); indx != -1 {
		return []*web3.Block{block}, indx, nil
	}

	// Backfill. We dont know the parent of the block.
	// Need to query the chain untill we find a known block

	added := []*web3.Block{block}
	var indx int

	for {
		// TODO: Add a counter to avoid big backfills.
		parent, err := t.provider.GetBlockByHash(block.ParentHash, false)
		if err != nil {
			return nil, -1, fmt.Errorf("Parent with hash %s not found", block.ParentHash)
		}

		added = append(added, parent)
		if indx = t.blockAtIndex(parent.ParentHash); indx != -1 {
			break
		}
		block = parent
	}

	// need the blocks in reverse order
	blocks := []*web3.Block{}
	for i := len(added) - 1; i >= 0; i-- {
		blocks = append(blocks, added[i])
	}
	return blocks, indx, nil
}

// EventType is the type of the event
type EventType int

const (
	// EventAdd happens when a new event is included in the chain
	EventAdd EventType = iota
	// EventDel may happen when there is a reorg and a past event is deleted
	EventDel
)

// Event is a log event
type Event struct {
	Type      EventType
	Added     []*web3.Block
	AddedLogs []*web3.Log

	Removed     []*web3.Block
	RemovedLogs []*web3.Log // TODO, build better
}

func (e *Event) removeBlock(b *web3.Block) {
	e.Removed = append(e.Removed, b)
}

func (e *Event) addLogs(logs []*web3.Log) {
	e.AddedLogs = append(e.AddedLogs, logs...)
}

func (e *Event) addBlock(b *web3.Block) {
	e.Added = append(e.Added, b)
}

func (e *Event) removeLogs(logs []*web3.Log) {
	e.RemovedLogs = append(e.RemovedLogs, logs...)
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
