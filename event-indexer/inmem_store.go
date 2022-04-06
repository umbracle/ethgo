package indexer

import (
	"math/big"
	"sync"

	"github.com/umbracle/ethgo"
)

var _ Entry = (*inmemEntry)(nil)

// NewInmemStore returns a new in-memory store.
func NewInmemStore() Entry {
	e := &inmemEntry{
		logs: []*ethgo.Log{},
	}
	return e
}

// Entry is a store.Entry implementation
type inmemEntry struct {
	l    sync.RWMutex
	logs []*ethgo.Log
	last *ethgo.Block
}

func (e *inmemEntry) Close() error {
	return nil
}

// LastIndex implements the store interface
func (e *inmemEntry) LastIndex() (uint64, error) {
	e.l.Lock()
	defer e.l.Unlock()
	return uint64(len(e.logs)), nil
}

// Logs returns the logs of the inmemory store
func (e *inmemEntry) Logs() []*ethgo.Log {
	return e.logs
}

func (e *inmemEntry) GetLastBlock() (*ethgo.Block, error) {
	e.l.Lock()
	defer e.l.Unlock()

	last := e.last
	if last == nil {
		return nil, nil
	}
	last = last.Copy()
	return last, nil
}

func (e *inmemEntry) StoreEvent(evnt *Event) error {
	e.l.Lock()
	defer e.l.Unlock()

	if evnt.Indx >= 0 {
		// remove logs
		e.logs = e.logs[:evnt.Indx]
	}
	// append new logs
	e.storeLogs(evnt.Added)

	if evnt.Block != nil {
		b := evnt.Block

		if b.Difficulty == nil {
			b.Difficulty = big.NewInt(0)
		}
		e.last = b.Copy()
	}
	return nil
}

func (e *inmemEntry) storeLogs(logs []*ethgo.Log) {
	e.logs = append(e.logs, logs...)
}

// GetLog implements the store interface
func (e *inmemEntry) GetLog(indx uint64, log *ethgo.Log) error {
	e.l.Lock()
	defer e.l.Unlock()

	*log = *e.logs[indx]
	return nil
}
