package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo/testutil"
)

func TestDebug_TraceTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)
	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("address", true))
	cc.EmitEvent("setA", "A", addr0.String())

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	r, err := s.TxnTo(addr, "setA2")
	require.NoError(t, err)

	trace, err := c.Debug().TraceTransaction(r.TransactionHash, nil)
	assert.NoError(t, err)
	assert.Greater(t, trace.Gas, uint64(20000))
	assert.NotEmpty(t, trace.StructLogs)
}

func Test_toTraceTransactionOpts(t *testing.T) {
	tests := []struct {
		name string
		opts *TraceTransactionOptions
		want map[string]interface{}
	}{
		{
			name: "nil options provided",
			opts: nil,
			want: map[string]interface{}{},
		},
		{
			name: "all fields are provided",
			opts: &TraceTransactionOptions{
				EnableMemory:     true,
				DisableStack:     true,
				DisableStorage:   true,
				EnableReturnData: true,
				Timeout:          "1s",
				Tracer:           "callTracer",
				TracerConfig: map[string]interface{}{
					"onlyTopCall": true,
					"withLog":     true,
				},
			},
			want: map[string]interface{}{
				"disableStack":     true,
				"disableStorage":   true,
				"enableMemory":     true,
				"enableReturnData": true,
				"timeout":          "1s",
				"tracer":           "callTracer",
				"tracerConfig": map[string]interface{}{
					"onlyTopCall": true,
					"withLog":     true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, toTraceTransactionOpts(tt.opts), "toTraceTransactionOpts(%v)", tt.opts)
		})
	}
}
