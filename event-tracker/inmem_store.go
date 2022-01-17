package tracker

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"sync"

	web3 "github.com/umbracle/go-web3"
)

var (
	dbGenesis = "genesis"
	//dbChainID   = "chainID"
	dbLastBlock = "lastBlock"
	//dbFilter    = "filter"
)

var _ Store = (*InmemStore)(nil)

// InmemStore implements the Store interface.
type InmemStore struct {
	l       sync.RWMutex
	entries map[string]*inmemEntry
	kv      map[string]string
}

// NewInmemStore returns a new in-memory store.
func NewInmemStore() *InmemStore {
	return &InmemStore{
		entries: map[string]*inmemEntry{},
		kv:      map[string]string{},
	}
}

// Close implements the store interface
func (i *InmemStore) Close() error {
	return nil
}

// Get implements the store interface
func (i *InmemStore) Get(k string) (string, error) {
	i.l.Lock()
	defer i.l.Unlock()
	return i.kv[string(k)], nil
}

// ListPrefix implements the store interface
func (i *InmemStore) ListPrefix(prefix string) ([]string, error) {
	i.l.Lock()
	defer i.l.Unlock()

	res := []string{}
	for k, v := range i.kv {
		if strings.HasPrefix(k, prefix) {
			res = append(res, v)
		}
	}
	return res, nil
}

// Set implements the store interface
func (i *InmemStore) Set(k, v string) error {
	i.l.Lock()
	defer i.l.Unlock()
	i.kv[string(k)] = v
	return nil
}

// GetEntry implements the store interface
func (i *InmemStore) GetEntry(filter *FilterConfig) (Entry, error) {
	i.l.Lock()
	defer i.l.Unlock()
	hash := filter.Hash
	e, ok := i.entries[hash]
	if ok {
		return e, nil
	}
	e = &inmemEntry{
		store: i,
		hash:  hash,
		logs:  []*web3.Log{},
	}
	i.entries[hash] = e
	return e, nil
}

// Entry is a store.Entry implementation
type inmemEntry struct {
	store *InmemStore
	hash  string
	l     sync.RWMutex
	logs  []*web3.Log
}

func (e *inmemEntry) Close() error {
	return nil
}

func (e *inmemEntry) UpsertGenesis(g *Genesis) error {
	raw, err := e.store.Get(dbGenesis)
	if err != nil {
		return err
	}
	if len(raw) != 0 {
		var found *Genesis
		if err := json.Unmarshal([]byte(raw), &found); err != nil {
			return err
		}
		return found.Equal(g)
	} else {
		data, err := json.Marshal(g)
		if err != nil {
			return err
		}
		if err := e.store.Set(dbGenesis, string(data)); err != nil {
			return err
		}
	}
	return nil
}

// LastIndex implements the store interface
func (e *inmemEntry) LastIndex() (uint64, error) {
	e.l.Lock()
	defer e.l.Unlock()
	return uint64(len(e.logs)), nil
}

// Logs returns the logs of the inmemory store
func (e *inmemEntry) Logs() []*web3.Log {
	return e.logs
}

func (e *inmemEntry) GetLastBlock() (*web3.Block, error) {
	buf, err := e.store.Get(dbLastBlock + "_" + e.hash)
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

func (e *inmemEntry) StoreEvent(evnt *Event) error {
	if evnt.Indx >= 0 {
		if err := e.RemoveLogs(uint64(evnt.Indx)); err != nil {
			return err
		}
	}
	if err := e.StoreLogs(evnt.Added); err != nil {
		return err
	}
	if evnt.Block != nil {
		b := evnt.Block

		if b.Difficulty == nil {
			b.Difficulty = big.NewInt(0)
		}
		buf, err := b.MarshalJSON()
		if err != nil {
			return err
		}
		raw := hex.EncodeToString(buf)
		return e.store.Set(dbLastBlock+"_"+e.hash, raw)
	}
	return nil
}

// StoreLogs implements the store interface
func (e *inmemEntry) StoreLogs(logs []*web3.Log) error {
	e.l.Lock()
	defer e.l.Unlock()
	e.logs = append(e.logs, logs...)
	return nil
}

// RemoveLogs implements the store interface
func (e *inmemEntry) RemoveLogs(indx uint64) error {
	e.l.Lock()
	defer e.l.Unlock()
	e.logs = e.logs[:indx]
	return nil
}

// GetLog implements the store interface
func (e *inmemEntry) GetLog(indx uint64, log *web3.Log) error {
	*log = *e.logs[indx]
	return nil
}
