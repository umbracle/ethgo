package ens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/testutil"
)

func TestENS_Resolve(t *testing.T) {
	ens, err := NewENS(WithAddress(testutil.TestInfuraEndpoint(t)))
	assert.NoError(t, err)

	addr, err := ens.Resolve("nick.eth")
	assert.NoError(t, err)
	assert.Equal(t, ethgo.HexToAddress("0xb8c2C29ee19D8307cb7255e1Cd9CbDE883A267d5"), addr)

	name, err := ens.ReverseResolve(ethgo.HexToAddress("0xb8c2C29ee19D8307cb7255e1Cd9CbDE883A267d5"))
	assert.NoError(t, err)
	assert.Equal(t, "nick.eth", name)
}
