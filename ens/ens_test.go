package ens

import (
	"testing"

	"github.com/cloudwalk/ethgo"
	"github.com/cloudwalk/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

func TestENS_Resolve(t *testing.T) {
	ens, err := NewENS(WithAddress(testutil.TestInfuraEndpoint(t)))
	assert.NoError(t, err)

	addr, err := ens.Resolve("arachnid.eth")
	assert.NoError(t, err)
	assert.Equal(t, ethgo.HexToAddress("0xb8c2C29ee19D8307cb7255e1Cd9CbDE883A267d5"), addr)
}
