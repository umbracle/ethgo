package tracker

import web3 "github.com/umbracle/go-web3"

// Store is a datastore for the tracker
type Store interface {
	// GetEntry returns a specific entry
	GetEntry(*FilterConfig) (Entry, error)
}

// Entry is a filter entry in the store
type Entry interface {
	// LastIndex returns index of the last stored event
	LastIndex() (uint64, error)

	// Closes closes the connection with the entry
	Close() error

	// UpsertGenesis check that genesis is correct
	UpsertGenesis(g *Genesis) error

	// StoreEvent stores an event that include added and removed logs
	StoreEvent(evnt *Event) error

	// GetLog returns the valid log at indx
	GetLog(indx uint64, log *web3.Log) error

	// GetLastBlock returns the last block processed
	GetLastBlock() (*web3.Block, error)
}
