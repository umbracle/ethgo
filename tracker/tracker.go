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
	"github.com/umbracle/go-web3/blocktracker"
	"github.com/umbracle/go-web3/etherscan"
	"github.com/umbracle/go-web3/jsonrpc/codec"
	"github.com/umbracle/go-web3/tracker/store"
	"github.com/umbracle/go-web3/tracker/store/inmem"
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
	Start   uint64
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

// Config is the configuration of the tracker
type Config struct {
	BatchSize       uint64
	BlockTracker    *blocktracker.BlockTracker // move to interface
	EtherscanAPIKey string
	Filter          *FilterConfig
	Store           store.Store
}

type ConfigOption func(*Config)

func WithBatchSize(b uint64) ConfigOption {
	return func(c *Config) {
		c.BatchSize = b
	}
}

func WithBlockTracker(b *blocktracker.BlockTracker) ConfigOption {
	return func(c *Config) {
		c.BlockTracker = b
	}
}

func WithStore(s store.Store) ConfigOption {
	return func(c *Config) {
		c.Store = s
	}
}

func WithFilter(f *FilterConfig) ConfigOption {
	return func(c *Config) {
		c.Filter = f
	}
}

func WithEtherscan(k string) ConfigOption {
	return func(c *Config) {
		c.EtherscanAPIKey = k
	}
}

// DefaultConfig returns the default tracker config
func DefaultConfig() *Config {
	return &Config{
		BatchSize:       defaultBatchSize,
		Store:           inmem.NewInmemStore(),
		Filter:          &FilterConfig{},
		EtherscanAPIKey: "",
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
	logger       *log.Logger
	provider     Provider
	config       *Config
	store        store.Store
	entry        store.Entry
	preSyncOnce  sync.Once
	blockTracker *blocktracker.BlockTracker
	synced       int32
	BlockCh      chan *blocktracker.BlockEvent
	ReadyCh      chan struct{}
	SyncCh       chan uint64
	EventCh      chan *Event
	DoneCh       chan struct{}
}

// NewTracker creates a new tracker
func NewTracker(provider Provider, opts ...ConfigOption) (*Tracker, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	t := &Tracker{
		provider:     provider,
		config:       config,
		BlockCh:      make(chan *blocktracker.BlockEvent, 1),
		logger:       log.New(ioutil.Discard, "", log.LstdFlags),
		ReadyCh:      make(chan struct{}),
		store:        config.Store,
		blockTracker: config.BlockTracker,
		DoneCh:       make(chan struct{}, 1),
		EventCh:      make(chan *Event),
		SyncCh:       make(chan uint64, 1),
		synced:       0,
	}
	if err := t.setupFilter(); err != nil {
		return nil, err
	}
	return t, nil
}

// NewFilter creates a new log filter
func (t *Tracker) setupFilter() error {
	if t.config.Filter == nil {
		// generic config
		t.config.Filter = &FilterConfig{}
	}

	// generate a random hash if not provided
	if t.config.Filter.Hash == "" {
		t.config.Filter.buildHash()
	}

	entry, err := t.store.GetEntry(t.config.Filter.Hash)
	if err != nil {
		return err
	}
	t.entry = entry

	// insert the filter config in the db
	filterKey := dbFilter + "_" + t.config.Filter.Hash
	data, err := t.store.Get(filterKey)
	if err != nil {
		return err
	}
	if data == "" {
		raw, err := json.Marshal(t.config.Filter)
		if err != nil {
			return err
		}
		rawStr := hex.EncodeToString(raw)
		if err := t.store.Set(filterKey, rawStr); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracker) Entry() store.Entry {
	return t.entry
}

// GetLastBlock returns the last block processed for this filter
func (t *Tracker) GetLastBlock() (*web3.Block, error) {
	buf, err := t.store.Get(dbLastBlock + "_" + t.config.Filter.Hash)
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

func (t *Tracker) storeLastBlock(b *web3.Block) error {
	if b.Difficulty == nil {
		b.Difficulty = big.NewInt(0)
	}
	buf, err := b.MarshalJSON()
	if err != nil {
		return err
	}
	raw := hex.EncodeToString(buf)
	return t.store.Set(dbLastBlock+"_"+t.config.Filter.Hash, raw)
}

func (t *Tracker) emitEvent(evnt *Event) {
	if evnt == nil {
		return
	}
	if t.config.Filter.Async {
		select {
		case t.EventCh <- evnt:
		default:
		}
	} else {
		t.EventCh <- evnt
	}
}

// IsSynced returns true if the filter is synced to head
func (t *Tracker) IsSynced() bool {
	return atomic.LoadInt32(&t.synced) != 0
}

// Wait waits the filter to finish
func (t *Tracker) Wait() {
	t.WaitDuration(0)
}

// WaitDuration waits for the filter to finish up to duration
func (t *Tracker) WaitDuration(dur time.Duration) error {
	if t.IsSynced() {
		return nil
	}

	var waitCh <-chan time.Time
	if dur == 0 {
		waitCh = time.After(dur)
	}
	select {
	case <-waitCh:
		return fmt.Errorf("timeout")
	case <-t.DoneCh:
	}
	return nil
}

func (t *Tracker) findAncestor(block, pivot *web3.Block) (uint64, error) {
	// block is part of a fork that is not the current head, find a common ancestor
	// both block and pivot are at the same height
	var err error

	for i := uint64(0); i < t.blockTracker.MaxBlockBacklog(); i++ {
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
	return 0, fmt.Errorf("the reorg is bigger than maxBlockBacklog %d", t.blockTracker.MaxBlockBacklog())
}

func (t *Tracker) emitLogs(typ EventType, logs []*web3.Log) {
	evnt := &Event{}
	if typ == EventAdd {
		evnt.Added = logs
	}
	if typ == EventDel {
		evnt.Removed = logs
	}
	t.emitEvent(evnt)
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

	if t.SyncCh != nil {
		select {
		case t.SyncCh <- dst:
		default:
		}
	}

	// add logs to the store
	if err := t.entry.StoreLogs(logs); err != nil {
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
	// Try to use first the user provided block if any
	if filterConfig.Start != 0 {
		bb, err := t.provider.GetBlockByNumber(web3.BlockNumber(filterConfig.Start), false)
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

func (t *Tracker) BatchSync(ctx context.Context) error {
	if err := t.preSyncCheck(); err != nil {
		return err
	}

	if t.blockTracker == nil {
		// run a specfic block tracker
		t.blockTracker = blocktracker.NewBlockTracker(t.provider)
		if err := t.blockTracker.Init(); err != nil {
			return err
		}
		go t.blockTracker.Start()
		go func() {
			// track our stop
			<-ctx.Done()
			t.blockTracker.Close()
		}()
	} else {
		// just try to init
		if err := t.blockTracker.Init(); err != nil {
			return err
		}
	}

	close(t.ReadyCh)

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

	// subscribe and sync
	sub := t.blockTracker.Subscribe()
	go func() {
		for {
			select {
			case evnt := <-sub:
				t.handleBlockEvnt(evnt)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (t *Tracker) syncImpl(ctx context.Context) error {
	if err := t.preSyncCheck(); err != nil {
		return err
	}

	lock := t.blockTracker.AcquireLock()
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
	if t.blockTracker.Len() == 0 {
		return nil
	}

	// get the current target
	target := t.blockTracker.LastBlocked()
	if target == nil {
		return nil
	}
	targetNum := target.Number

	last, err := t.GetLastBlock()
	if err != nil {
		return err
	}
	if last == nil {
		// Try to fast track to the valid block (if possible)
		last, err = t.fastTrack(t.config.Filter)
		if err != nil {
			return fmt.Errorf("failed to fasttrack: %v", err)
		}
		if last != nil {
			if err := t.storeLastBlock(last); err != nil {
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

	step := targetNum - origin + 1
	if step > t.blockTracker.MaxBlockBacklog() {
		// we are far (more than maxBackLog) from the target block
		// Do a bulk sync with the eth_getLogs endpoint and get closer
		// to the target block.

		for {
			if origin > targetNum {
				return fmt.Errorf("from (%d) higher than to (%d)", origin, targetNum)
			}
			if targetNum-origin+1 <= t.blockTracker.MaxBlockBacklog() {
				break
			}

			// release the lock
			lock.Unlock()

			limit := targetNum - t.blockTracker.MaxBlockBacklog()
			if err := t.syncBatch(ctx, origin, limit); err != nil {
				return err
			}

			origin = limit + 1

			// lock again to reset the target block
			lock.Lock()
			targetNum = t.blockTracker.LastBlocked().Number
		}
	}

	// we are still holding the lock on the blocksLock so that we are sure
	// that the targetNum has not changed
	trackerBlocks := t.blockTracker.BlocksBlocked()
	added := trackerBlocks[uint64(len(trackerBlocks))-1-(targetNum-origin):]

	evnt, err := t.doFilter(added, nil)
	if err != nil {
		return err
	}
	if evnt != nil {
		t.emitEvent(evnt)
	}

	// release the lock on the blocks
	lock.Unlock()
	return nil
}

func (t *Tracker) removeLogs(number uint64, hash *web3.Hash) ([]*web3.Log, error) {
	index, err := t.entry.LastIndex()
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
		if err := t.entry.GetLog(elemIndex, &log); err != nil {
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

	if err := t.entry.RemoveLogs(index); err != nil {
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

func (t *Tracker) handleBlockEvnt(blockEvnt *blocktracker.BlockEvent) error {
	if blockEvnt == nil {
		return nil
	}

	// emit the block event
	select {
	case t.BlockCh <- blockEvnt:
	default:
	}

	if t.IsSynced() {
		evnt, err := t.doFilter(blockEvnt.Added, blockEvnt.Removed)
		if err != nil {
			return err
		}
		if evnt != nil {
			t.emitEvent(evnt)
		}
	}
	return nil
}

func (t *Tracker) doFilter(added []*web3.Block, removed []*web3.Block) (*Event, error) {
	evnt := &Event{}
	if len(removed) != 0 {
		pivot := removed[0]
		logs, err := t.removeLogs(pivot.Number, &pivot.Hash)
		if err != nil {
			return nil, err
		}
		evnt.Removed = append(evnt.Removed, revertLogs(logs)...)
	}

	for _, block := range added {
		// check logs for this blocks
		query := t.config.Filter.getFilterSearch()
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
		if err := t.entry.StoreLogs(logs); err != nil {
			return nil, err
		}
		evnt.Added = append(evnt.Added, logs...)
	}

	// store the last block as the new index
	if err := t.storeLastBlock(added[len(added)-1]); err != nil {
		return nil, err
	}
	return evnt, nil
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
