package store

import web3 "github.com/umbracle/go-web3"

// Store is a datastore for the tracker
type Store interface {
	// LastIndex returns index of the last stored event
	LastIndex() (uint64, error)

	// StoreLogs stores the web3 logs of the event
	StoreLogs(logs []*web3.Log) error

	// RemoveLogs all the logs starting at index 'indx'
	RemoveLogs(indx uint64) error

	// GetLog returns the log at indx
	GetLog(indx uint64, log *web3.Log) error

	// Get gets a value
	Get(k []byte) ([]byte, error)

	// Set sets a value
	Set(k, v []byte) error

	// Close closes the store
	Close() error
}
