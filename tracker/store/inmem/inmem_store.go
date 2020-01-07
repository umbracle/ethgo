package inmem

import (
	"bytes"
	"sync"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/tracker/store"
)

var _ store.Store = (*InmemStore)(nil)

// InmemStore implements the Store interface.
type InmemStore struct {
	l       sync.RWMutex
	entries map[string]*Entry
	kv      map[string][]byte
}

// NewInmemStore returns a new in-memory store.
func NewInmemStore() *InmemStore {
	return &InmemStore{
		entries: map[string]*Entry{},
		kv:      map[string][]byte{},
	}
}

// Close implements the store interface
func (i *InmemStore) Close() error {
	return nil
}

// Get implements the store interface
func (i *InmemStore) Get(k []byte) ([]byte, error) {
	i.l.Lock()
	defer i.l.Unlock()
	return i.kv[string(k)], nil
}

// ListPrefix implements the store interface
func (i *InmemStore) ListPrefix(prefix []byte) ([][]byte, error) {
	i.l.Lock()
	defer i.l.Unlock()

	res := [][]byte{}
	for k, v := range i.kv {
		if bytes.HasPrefix([]byte(k), prefix) {
			res = append(res, v)
		}
	}
	return res, nil
}

// Set implements the store interface
func (i *InmemStore) Set(k, v []byte) error {
	i.l.Lock()
	defer i.l.Unlock()
	i.kv[string(k)] = v
	return nil
}

// GetEntry implements the store interface
func (i *InmemStore) GetEntry(hash string) (store.Entry, error) {
	i.l.Lock()
	defer i.l.Unlock()
	e, ok := i.entries[hash]
	if ok {
		return e, nil
	}
	e = &Entry{
		logs: []*web3.Log{},
	}
	i.entries[hash] = e
	return e, nil
}

// Entry is a store.Entry implementation
type Entry struct {
	l    sync.RWMutex
	logs []*web3.Log
}

// LastIndex implements the store interface
func (e *Entry) LastIndex() (uint64, error) {
	e.l.Lock()
	defer e.l.Unlock()
	return uint64(len(e.logs)), nil
}

// Logs returns the logs of the inmemory store
func (e *Entry) Logs() []*web3.Log {
	return e.logs
}

// StoreLogs implements the store interface
func (e *Entry) StoreLogs(logs []*web3.Log) error {
	e.l.Lock()
	defer e.l.Unlock()
	for _, log := range logs {
		e.logs = append(e.logs, log)
	}
	return nil
}

// RemoveLogs implements the store interface
func (e *Entry) RemoveLogs(indx uint64) error {
	e.l.Lock()
	defer e.l.Unlock()
	e.logs = e.logs[:indx]
	return nil
}

// GetLog implements the store interface
func (e *Entry) GetLog(indx uint64, log *web3.Log) error {
	*log = *e.logs[indx]
	return nil
}
