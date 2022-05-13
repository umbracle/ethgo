package jsonrpc

import (
	"testing"

	"github.com/cloudwalk/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDebug_TraceTransaction(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("address", true))
	cc.EmitEvent("setA", "A", addr0.String())

	_, addr := s.DeployContract(cc)
	r := s.TxnTo(addr, "setA2")

	trace, err := c.Debug().TraceTransaction(r.TransactionHash)
	assert.NoError(t, err)
	assert.Greater(t, trace.Gas, uint64(20000))
	assert.NotEmpty(t, trace.StructLogs)
}
