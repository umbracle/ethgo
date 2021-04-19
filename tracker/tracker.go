package tracker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/etherscan"
	"github.com/umbracle/go-web3/jsonrpc/codec"
	"github.com/umbracle/go-web3/tracker/store"
)

var (
	dbGenesis   = "genesis"
	dbChainID   = "chainID"
	dbLastBlock = "lastBlock"
	dbFilter    = "filter"
)

const (
	defaultMaxBlockBacklog = 10
	defaultBatchSize       = 100
)

// FilterConfig is a tracker filter configuration
type FilterConfig struct {
	Address []web3.Address `json:"address"`
	Topics  []*web3.Hash   `json:"topics"`
	Hash    string
	Async   bool
}

func (f *FilterConfig) buildHash() {
	h := sha256.New()
	for _, i := range f.Address {
		h.Write([]byte(i.String()))
	}
	for _, i := range f.Topics {
		if i == nil {
			h.Write([]byte("empty"))
		} else {
			h.Write([]byte(i.String()))
		}
	}
	f.Hash = hex.EncodeToString(h.Sum(nil))
}

func (f *FilterConfig) getFilterSearch() *web3.LogFilter {
	filter := &web3.LogFilter{}
	if len(f.Address) != 0 {
		filter.Address = f.Address
	}
	if len(f.Topics) != 0 {
		filter.Topics = f.Topics
	}
	return filter
}

// Filter is a specific filter
type Filter struct {
	synced  int32
	config  *FilterConfig
	SyncCh  chan uint64
	EventCh chan *Event
	DoneCh  chan struct{}
	entry   store.Entry
	tracker *Tracker
}

func (f *Filter) Entry() store.Entry {
	return f.entry
}

// GetLastBlock returns the last block processed for this filter
func (f *Filter) GetLastBlock() (*web3.Block, error) {
	buf, err := f.tracker.store.Get(dbLastBlock + "_" + f.config.Hash)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, nil
	}
	raw, err := hex.DecodeString(buf)
	if err != nil {
		return nil, err
	}
	b := &web3.Block{}
	if err := b.UnmarshalJSON(raw); err != nil {
		return nil, err
	}
	return b, nil
}

func (f *Filter) storeLastBlock(b *web3.Block) error {
	if b.Difficulty == nil {
		b.Difficulty = big.NewInt(0)
	}
	buf, err := b.MarshalJSON()
	if err != nil {
		return err
	}
	raw := hex.EncodeToString(buf)
	return f.tracker.store.Set(dbLastBlock+"_"+f.config.Hash, raw)
}

// SyncAsync syncs the filter asynchronously
func (f *Filter) SyncAsync(ctx context.Context) {
	f.tracker.SyncAsync(ctx, f)
}

// Sync syncs the filter
func (f *Filter) Sync(ctx context.Context) error {
	return f.tracker.Sync(ctx, f)
}

func (f *Filter) emitEvent(evnt *Event) {
	if evnt == nil {
		return
	}
	if f.config.Async {
		select {
		case f.EventCh <- evnt:
		default:
		}
	} else {
		f.EventCh <- evnt
	}
}

// IsSynced returns true if the filter is synced to head
func (f *Filter) IsSynced() bool {
	return atomic.LoadInt32(&f.synced) != 0
}

// Wait waits the filter to finish
func (f *Filter) Wait() {
	f.WaitDuration(0)
}

// WaitDuration waits for the filter to finish up to duration
func (f *Filter) WaitDuration(dur time.Duration) error {
	if f.IsSynced() {
		return nil
	}

	var waitCh <-chan time.Time
	if dur == 0 {
		waitCh = time.After(dur)
	}
	select {
	case <-waitCh:
		return fmt.Errorf("timeout")
	case <-f.DoneCh:
	}
	return nil
}

// Config is the configuration of the tracker
type Config struct {
	BatchSize          uint64
	MaxBlockBacklog    uint64
	EtherscanFastTrack bool
	EtherscanAPIKey    string
}

// DefaultConfig returns the default tracker config
func DefaultConfig() *Config {
	return &Config{
		BatchSize:          defaultBatchSize,
		MaxBlockBacklog:    defaultMaxBlockBacklog,
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
	logger      *log.Logger
	provider    Provider
	config      *Config
	store       store.Store
	preSyncOnce sync.Once

	blocks     []*web3.Block
	blocksLock sync.Mutex

	filterLock sync.Mutex
	filters    []*Filter

	blockTracker BlockTracker
	BlockCh      chan *BlockEvent

	ReadyCh chan struct{}
}

// NewTracker creates a new tracker
func NewTracker(provider Provider, config *Config) *Tracker {
	if config.BatchSize == 0 {
		config.BatchSize = defaultBatchSize
	}
	if config.MaxBlockBacklog == 0 {
		config.MaxBlockBacklog = defaultMaxBlockBacklog
	}
	return &Tracker{
		provider: provider,
		config:   config,
		blocks:   []*web3.Block{},
		filters:  []*Filter{},
		BlockCh:  make(chan *BlockEvent, 1),
		logger:   log.New(ioutil.Discard, "", log.LstdFlags),
		ReadyCh:  make(chan struct{}),
	}
}

// SetLogger sets a logger
func (t *Tracker) SetLogger(logger *log.Logger) {
	t.logger = logger
}

// SetStore sets the store
func (t *Tracker) SetStore(store store.Store) {
	t.store = store
}

// NewFilter creates a new log filter
func (t *Tracker) NewFilter(config *FilterConfig) (*Filter, error) {
	if config == nil {
		// generic config
		config = &FilterConfig{}
	}

	// generate a random hash if not provided
	config.buildHash()

	entry, err := t.store.GetEntry(config.Hash)
	if err != nil {
		return nil, err
	}

	f := &Filter{
		config:  config,
		DoneCh:  make(chan struct{}, 1),
		EventCh: make(chan *Event),
		SyncCh:  make(chan uint64, 1),
		entry:   entry,
		synced:  0,
		tracker: t,
	}

	// insert the filter config in the db
	filterKey := dbFilter + "_" + config.Hash
	data, err := t.store.Get(filterKey)
	if err != nil {
		return nil, err
	}
	if data == "" {
		raw, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		rawStr := hex.EncodeToString(raw)
		if err := t.store.Set(filterKey, rawStr); err != nil {
			return nil, err
		}
	}

	t.filterLock.Lock()
	t.filters = append(t.filters, f)
	t.filterLock.Unlock()

	return f, nil
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

func (f *Filter) emitLogs(typ EventType, logs []*web3.Log) {
	evnt := &Event{}
	if typ == EventAdd {
		evnt.Added = logs
	}
	if typ == EventDel {
		evnt.Removed = logs
	}
	f.emitEvent(evnt)
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

func (t *Tracker) syncBatch(ctx context.Context, filter *Filter, from, to uint64) error {
	query := filter.config.getFilterSearch()

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

	if filter.SyncCh != nil {
		select {
		case filter.SyncCh <- dst:
		default:
		}
	}

	// add logs to the store
	if err := filter.entry.StoreLogs(logs); err != nil {
		return err
	}
	filter.emitLogs(EventAdd, logs)

	// update the last block entry
	block, err := t.provider.GetBlockByNumber(web3.BlockNumber(dst), false)
	if err != nil {
		return err
	}
	if err := filter.storeLastBlock(block); err != nil {
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

func (t *Tracker) preSyncCheck() error {
	var err error
	t.preSyncOnce.Do(func() {
		err = t.preSyncCheckImpl()
	})
	return err
}

func (t *Tracker) preSyncCheckImpl() error {
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
		if genesis != rGenesis.Hash.String() {
			return fmt.Errorf("bad genesis")
		}
		if chainID != rChainID.String() {
			return fmt.Errorf("bad genesis")
		}
	} else {
		if err := t.store.Set(dbGenesis, rGenesis.Hash.String()); err != nil {
			return err
		}
		if err := t.store.Set(dbChainID, rChainID.String()); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracker) fastTrack(filterConfig *FilterConfig) (*web3.Block, error) {
	// Only possible if we filter addresses
	if len(filterConfig.Address) == 0 {
		return nil, nil
	}

	if t.config.EtherscanFastTrack {
		chainID, err := t.provider.ChainID()
		if err != nil {
			return nil, err
		}

		// get the etherscan instance for this chainID
		e, err := etherscan.NewEtherscanFromNetwork(web3.Network(chainID.Uint64()), t.config.EtherscanAPIKey)
		if err != nil {
			// there is no etherscan api for this specific chainid
			return nil, nil
		}

		getAddress := func(addr web3.Address) (uint64, error) {
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

		bb, err := t.provider.GetBlockByNumber(web3.BlockNumber(minBlock-1), false)
		if err != nil {
			return nil, err
		}
		return bb, nil
	}

	return nil, nil
}

func (t *Tracker) populateBlocks() ([]*web3.Block, error) {
	block, err := t.provider.GetBlockByNumber(web3.Latest, false)
	if err != nil {
		return nil, err
	}
	if block.Number == 0 {
		return []*web3.Block{}, nil
	}

	blocks := make([]*web3.Block, t.config.MaxBlockBacklog)

	var i uint64
	for i = 0; i < t.config.MaxBlockBacklog; i++ {
		blocks[t.config.MaxBlockBacklog-i-1] = block
		if block.Number == 0 {
			break
		}
		block, err = t.provider.GetBlockByHash(block.ParentHash, false)
		if err != nil {
			return nil, err
		}
	}

	if i != t.config.MaxBlockBacklog {
		// less than maxBacklog elements
		blocks = blocks[t.config.MaxBlockBacklog-i-1:]
	}
	return blocks, nil
}

// SyncAsync syncs a specific filter asynchronously
func (t *Tracker) SyncAsync(ctx context.Context, filter *Filter) {
	go t.Sync(ctx, filter)
}

// Sync syncs a specific filter
func (t *Tracker) Sync(ctx context.Context, filter *Filter) error {
	if err := t.syncImpl(ctx, filter); err != nil {
		return err
	}

	select {
	case filter.DoneCh <- struct{}{}:
	default:
	}

	atomic.StoreInt32(&filter.synced, 1)
	return nil
}

func (t *Tracker) syncImpl(ctx context.Context, filter *Filter) error {
	if err := t.preSyncCheck(); err != nil {
		return err
	}

	lock := lock{lock: &t.blocksLock}
	defer func() {
		if lock.Locked {
			lock.Unlock()
		}
	}()

	// We only hold the lock when we sync the head (last MaxBackLogs)
	// because we want to avoid changes in the head while we sync.
	// We will only release the lock if we do a bulk sync since it can
	// move the target block for the sync.

	lock.Lock()
	if len(t.blocks) == 0 {
		return nil
	}

	// get the current target
	target := t.blocks[len(t.blocks)-1]
	if target == nil {
		return nil
	}
	targetNum := target.Number

	last, err := filter.GetLastBlock()
	if err != nil {
		return err
	}
	if last == nil {
		// Try to fast track to the valid block (if possible)
		last, err = t.fastTrack(filter.config)
		if err != nil {
			return fmt.Errorf("failed to fasttrack: %v", err)
		}
		if last != nil {
			if err := filter.storeLastBlock(last); err != nil {
				return err
			}
		}
	} else {
		if last.Hash == target.Hash {
			return nil
		}
	}

	// There might been a reorg when we stopped syncing last time,
	// check that our 'beacon' block matches the one in the chain.
	// If that is not the case, we consider beacon-maxBackLog our
	// real origin point and remove any logs ahead of that point.

	var origin uint64
	if last != nil {
		if last.Number > targetNum {
			return fmt.Errorf("store is more advanced than the chain")
		}

		pivot, err := t.provider.GetBlockByNumber(web3.BlockNumber(last.Number), false)
		if err != nil {
			return err
		}

		if last.Number == targetNum {
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
			logs, err := t.removeLogs(filter, ancestor+1, nil)
			if err != nil {
				return err
			}
			filter.emitLogs(EventDel, logs)

			last, err = t.provider.GetBlockByNumber(web3.BlockNumber(ancestor), false)
			if err != nil {
				return err
			}
		}
	}

	step := targetNum - origin + 1
	if step > t.config.MaxBlockBacklog {
		// we are far (more than maxBackLog) from the target block
		// Do a bulk sync with the eth_getLogs endpoint and get closer
		// to the target block.

		for {
			if origin > targetNum {
				return fmt.Errorf("from (%d) higher than to (%d)", origin, targetNum)
			}
			if targetNum-origin+1 <= t.config.MaxBlockBacklog {
				break
			}

			// release the lock
			lock.Unlock()

			limit := targetNum - t.config.MaxBlockBacklog
			if err := t.syncBatch(ctx, filter, origin, limit); err != nil {
				return err
			}

			origin = limit + 1

			// lock again to reset the target block
			lock.Lock()
			targetNum = t.blocks[len(t.blocks)-1].Number
		}
	}

	// we are still holding the lock on the blocksLock so that we are sure
	// that the targetNum has not changed
	added := t.blocks[uint64(len(t.blocks))-1-(targetNum-origin):]

	evnt, err := t.doFilter(filter, added, nil)
	if err != nil {
		return err
	}
	if evnt != nil {
		filter.emitEvent(evnt)
	}

	// release the lock on the blocks
	lock.Unlock()
	return nil
}

// Start starts the syncing
func (t *Tracker) Start(ctx context.Context) error {
	if t.blockTracker == nil {
		t.blockTracker = NewJSONBlockTracker(t.logger, t.provider)
	}
	if err := t.preSyncCheck(); err != nil {
		return err
	}

	blocks, err := t.populateBlocks()
	if err != nil {
		return err
	}
	t.blocks = blocks

	close(t.ReadyCh)

	// start the polling
	err = t.blockTracker.Track(ctx, func(block *web3.Block) error {
		return t.handleReconcile(block)
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tracker) addBlockLocked(block *web3.Block) error {
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

func (t *Tracker) blockAtIndex(hash web3.Hash) int {
	for indx, b := range t.blocks {
		if b.Hash == hash {
			return indx
		}
	}
	return -1
}

func (t *Tracker) removeLogs(filter *Filter, number uint64, hash *web3.Hash) ([]*web3.Log, error) {
	index, err := filter.entry.LastIndex()
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
		if err := filter.entry.GetLog(elemIndex, &log); err != nil {
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

	if err := filter.entry.RemoveLogs(index); err != nil {
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

func (t *Tracker) handleBlockEvent(block *web3.Block) (*BlockEvent, error) {
	t.blocksLock.Lock()
	defer t.blocksLock.Unlock()

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
		if err := t.addBlockLocked(block); err != nil {
			return nil, err
		}
	}
	return blockEvnt, nil
}

func (t *Tracker) handleReconcile(block *web3.Block) error {
	blockEvnt, err := t.handleBlockEvent(block)
	if err != nil {
		return err
	}
	if blockEvnt == nil {
		return nil
	}

	// emit the block event
	select {
	case t.BlockCh <- blockEvnt:
	default:
	}

	t.filterLock.Lock()
	defer t.filterLock.Unlock()

	for _, filter := range t.filters {
		if filter.IsSynced() {
			evnt, err := t.doFilter(filter, blockEvnt.Added, blockEvnt.Removed)
			if err != nil {
				return err
			}
			if evnt != nil {
				filter.emitEvent(evnt)
			}
		}
	}

	return nil
}

func (t *Tracker) doFilter(filter *Filter, added []*web3.Block, removed []*web3.Block) (*Event, error) {
	evnt := &Event{}
	if len(removed) != 0 {
		pivot := removed[0]
		logs, err := t.removeLogs(filter, pivot.Number, &pivot.Hash)
		if err != nil {
			return nil, err
		}
		evnt.Removed = append(evnt.Removed, revertLogs(logs)...)
	}

	for _, block := range added {
		// check logs for this blocks
		query := filter.config.getFilterSearch()
		query.BlockHash = &block.Hash

		// We check the hash, we need to do a retry to let unsynced nodes get the block
		var logs []*web3.Log
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

		// add logs to the store
		if err := filter.entry.StoreLogs(logs); err != nil {
			return nil, err
		}
		evnt.Added = append(evnt.Added, logs...)
	}

	// store the last block as the new index
	if err := filter.storeLastBlock(added[len(added)-1]); err != nil {
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

	count := uint64(0)
	for {
		if count > t.config.MaxBlockBacklog {
			return nil, -1, fmt.Errorf("Cannot reconcile more than max backlog values")
		}
		count++

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

// GetSavedFilters returns the filters stored in the store
func (t *Tracker) GetSavedFilters() ([]*FilterConfig, error) {
	data, err := t.store.ListPrefix(dbFilter)
	if err != nil {
		return nil, err
	}

	config := []*FilterConfig{}
	for _, item := range data {
		raw, err := hex.DecodeString(item)
		if err != nil {
			return nil, err
		}
		var res *FilterConfig
		if err := json.Unmarshal(raw, &res); err != nil {
			return nil, err
		}
		config = append(config, res)
	}
	return config, nil
}

// EventType is the type of the event
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
	Added   []*web3.Log
	Removed []*web3.Log
}

// BlockEvent is an event emitted when a new block is included
type BlockEvent struct {
	Type    EventType
	Added   []*web3.Block
	Removed []*web3.Block
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

type lock struct {
	Locked bool
	lock   *sync.Mutex
}

func (l *lock) Lock() {
	l.Locked = true
	l.lock.Lock()
}

func (l *lock) Unlock() {
	l.Locked = false
	l.lock.Unlock()
}
