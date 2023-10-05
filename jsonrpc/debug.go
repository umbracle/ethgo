package jsonrpc

import "github.com/umbracle/ethgo"

type Debug struct {
	c *Client
}

// Debug returns the reference to the debug namespace
func (c *Client) Debug() *Debug {
	return c.endpoints.d
}

type TraceTransactionOptions struct {
	EnableMemory     bool
	DisableStack     bool
	DisableStorage   bool
	EnableReturnData bool
	Timeout          string
	Tracer           string
	TracerConfig     map[string]interface{}
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

func (d *Debug) TraceTransaction(hash ethgo.Hash, opts *TraceTransactionOptions) (*TransactionTrace, error) {
	var res *TransactionTrace
	err := d.c.Call("debug_traceTransaction", &res, hash, toTraceTransactionOpts(opts))
	return res, err
}

func toTraceTransactionOpts(opts *TraceTransactionOptions) map[string]interface{} {
	optsMap := make(map[string]interface{})

	if opts != nil {
		if opts.EnableMemory {
			optsMap["enableMemory"] = true
		}

		if opts.DisableStack {
			optsMap["disableStack"] = true
		}

		if opts.DisableStorage {
			optsMap["disableStorage"] = true
		}

		if opts.EnableReturnData {
			optsMap["enableReturnData"] = true
		}

		if opts.Timeout != "" {
			optsMap["timeout"] = opts.Timeout
		}

		if opts.Tracer != "" {
			optsMap["tracer"] = opts.Tracer

			if len(opts.TracerConfig) > 0 {
				optsMap["tracerConfig"] = opts.TracerConfig
			}
		}
	}

	return optsMap
}
