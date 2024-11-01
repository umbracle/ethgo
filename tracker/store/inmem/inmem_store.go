package inmem

import (
	"strings"
	"sync"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/tracker/store"
)

var _ store.Store = (*InmemStore)(nil)

// InmemStore implements the Store interface.
type InmemStore struct {
	l       sync.RWMutex
	entries map[string]*Entry
	kv      map[string]string
}

// NewInmemStore returns a new in-memory store.
func NewInmemStore() *InmemStore {
	return &InmemStore{
		entries: map[string]*Entry{},
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
func (i *InmemStore) GetEntry(hash string) (store.Entry, error) {
	i.l.Lock()
	defer i.l.Unlock()
	e, ok := i.entries[hash]
	if ok {
		return e, nil
	}
	e = &Entry{
		logs: []*ethgo.Log{},
	}
	i.entries[hash] = e
	return e, nil
}

// Entry is a store.Entry implementation
type Entry struct {
	l    sync.RWMutex
	logs []*ethgo.Log
}

// LastIndex implements the store interface
func (e *Entry) LastIndex() (uint64, error) {
	e.l.Lock()
	defer e.l.Unlock()
	return uint64(len(e.logs)), nil
}

// Logs returns the logs of the inmemory store
func (e *Entry) Logs() []*ethgo.Log {
	return e.logs
}

// StoreLogs implements the store interface
func (e *Entry) StoreLogs(logs []*ethgo.Log) error {
	e.l.Lock()
	defer e.l.Unlock()
	e.logs = append(e.logs, logs...)
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
func (e *Entry) GetLog(indx uint64, log *ethgo.Log) error {
	*log = *e.logs[indx]
	return nil
}
