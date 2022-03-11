package tracker

import "github.com/umbracle/ethgo"

// Entry is a filter entry in the store
type Entry interface {
	// LastIndex returns index of the last stored event
	LastIndex() (uint64, error)

	// Closes closes the connection with the entry
	Close() error

	// StoreEvent stores an event that include added and removed logs
	StoreEvent(evnt *Event) error

	// GetLog returns the valid log at indx
	GetLog(indx uint64, log *ethgo.Log) error

	// GetLastBlock returns the last block processed
	GetLastBlock() (*ethgo.Block, error)
}
