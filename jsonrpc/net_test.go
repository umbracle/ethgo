package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetVersion(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	_, err := s.Client().Net().Version()
	assert.NoError(t, err)
}

func TestNetListening(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	ok, err := s.Client().Net().Listening()
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestNetPeerCount(t *testing.T) {
	s := NewTestServer(t, nil)
	defer s.Close()

	count, err := s.Client().Net().PeerCount()
	assert.NoError(t, err)
	assert.Equal(t, count, uint64(0))
}
