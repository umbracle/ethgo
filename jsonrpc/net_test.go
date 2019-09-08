package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3/testutil"
)

func TestNetVersion(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	_, err := c.Net().Version()
	assert.NoError(t, err)
}

func TestNetListening(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	ok, err := c.Net().Listening()
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestNetPeerCount(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c := NewClient(s.HTTPAddr())
	count, err := c.Net().PeerCount()
	assert.NoError(t, err)
	assert.Equal(t, count, uint64(0))
}
