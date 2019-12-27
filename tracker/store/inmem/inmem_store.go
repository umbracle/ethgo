package inmem

import (
	"sync"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/tracker/store"
)

var _ store.Store = (*InmemStore)(nil)

// InmemStore implements the Store interface.
type InmemStore struct {
	l    sync.RWMutex
	logs []*web3.Log
	kv   map[string][]byte
}

// NewInmemStore returns a new in-memory store.
func NewInmemStore() *InmemStore {
	return &InmemStore{
		logs: []*web3.Log{},
		kv:   map[string][]byte{},
	}
}

// Logs returns the logs of the inmemory store
func (i *InmemStore) Logs() []*web3.Log {
	return i.logs
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

// Set implements the store interface
func (i *InmemStore) Set(k, v []byte) error {
	i.l.Lock()
	defer i.l.Unlock()
	i.kv[string(k)] = v
	return nil
}

// LastIndex implements the store interface
func (i *InmemStore) LastIndex() (uint64, error) {
	i.l.Lock()
	defer i.l.Unlock()
	return uint64(len(i.logs)), nil
}

// StoreLogs implements the store interface
func (i *InmemStore) StoreLogs(logs []*web3.Log) error {
	i.l.Lock()
	defer i.l.Unlock()
	for _, log := range logs {
		i.logs = append(i.logs, log)
	}
	return nil
}

// RemoveLogs implements the store interface
func (i *InmemStore) RemoveLogs(indx uint64) error {
	i.l.Lock()
	defer i.l.Unlock()
	i.logs = i.logs[:indx]
	return nil
}

// GetLog implements the store interface
func (i *InmemStore) GetLog(indx uint64, log *web3.Log) error {
	*log = *i.logs[indx]
	return nil
}
