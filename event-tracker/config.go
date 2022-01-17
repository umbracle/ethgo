package tracker

import (
	blocktracker "github.com/umbracle/go-web3/block-tracker"
)

const (
	defaultMaxBlockBacklog = 10
	defaultBatchSize       = 100
)

// Config is the configuration of the tracker
type Config struct {
	BatchSize       uint64
	BlockTracker    *blocktracker.BlockTracker // move to interface
	EtherscanAPIKey string
	StartBlock      uint64
	Filter          *FilterConfig
	Store           Store
	MaxBacklog      uint64
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

func WithStore(s Store) ConfigOption {
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

func WithStartBlock(block uint64) ConfigOption {
	return func(c *Config) {
		c.StartBlock = block
	}
}

func WithMaxBacklog(backLog uint64) ConfigOption {
	return func(c *Config) {
		c.MaxBacklog = backLog
	}
}

// DefaultConfig returns the default tracker config
func DefaultConfig() *Config {
	return &Config{
		BatchSize:       defaultBatchSize,
		Store:           NewInmemStore(),
		Filter:          &FilterConfig{},
		EtherscanAPIKey: "",
		MaxBacklog:      defaultMaxBlockBacklog,
	}
}
