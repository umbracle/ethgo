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

	trace, err := c.Debug().TraceTransaction(r.TransactionHash)
	assert.NoError(t, err)
	assert.Greater(t, trace.Gas, uint64(20000))
	assert.NotEmpty(t, trace.StructLogs)
}
