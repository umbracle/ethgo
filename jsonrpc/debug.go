package jsonrpc

import "github.com/Ethernal-Tech/ethgo"

type Debug struct {
	c *Client
}

// Debug returns the reference to the debug namespace
func (c *Client) Debug() *Debug {
	return c.endpoints.d
}

type TraceTransactionOptions struct {
	EnableMemory     bool                   `json:"enableMemory"`
	DisableStack     bool                   `json:"disableStack"`
	DisableStorage   bool                   `json:"disableStorage"`
	EnableReturnData bool                   `json:"enableReturnData"`
	Timeout          string                 `json:"timeout,omitempty"`
	Tracer           string                 `json:"tracer,omitempty"`
	TracerConfig     map[string]interface{} `json:"tracerConfig,omitempty"`
}

type TransactionTrace struct {
	Gas         uint64
	ReturnValue string
	StructLogs  []*StructLogs
}

type StructLogs struct {
	Depth   int
	Gas     int
	GasCost int
	Op      string
	Pc      int
	Memory  []string
	Stack   []string
	Storage map[string]string
}

func (d *Debug) TraceTransaction(hash ethgo.Hash, opts TraceTransactionOptions) (*TransactionTrace, error) {
	var res *TransactionTrace
	err := d.c.Call("debug_traceTransaction", &res, hash, opts)
	return res, err
}
