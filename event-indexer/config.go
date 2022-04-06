package tracker

import (
	"io/ioutil"
	"log"

	blocktracker "github.com/umbracle/ethgo/block-tracker"
)

const (
	defaultMaxBlockBacklog = 10
	defaultBatchSize       = 100
)

// Config is the configuration of the tracker
type Config struct {
	Logger          *log.Logger
	BatchSize       uint64
	BlockTracker    *blocktracker.BlockTracker // move to interface
	EtherscanAPIKey string
	StartBlock      uint64
	Filter          *FilterConfig
	Entry           Entry
	MaxBacklog      uint64
}

type ConfigOption func(*Config)

func WithLogger(logger *log.Logger) ConfigOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

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

func WithStore(s Entry) ConfigOption {
	return func(c *Config) {
		c.Entry = s
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
		Entry:           NewInmemStore(),
		Filter:          &FilterConfig{},
		EtherscanAPIKey: "",
		MaxBacklog:      defaultMaxBlockBacklog,
		Logger:          log.New(ioutil.Discard, "", log.LstdFlags),
	}
}
