package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo/testutil"
	"golang.org/x/crypto/sha3"
)

func TestWeb3ClientVersion(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		_, err := c.Web3().ClientVersion()
		assert.NoError(t, err)
	})
}

func TestWeb3Sha3(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		src := []byte{0x1, 0x2, 0x3}

		found, err := c.Web3().Sha3(src)
		assert.NoError(t, err)

		k := sha3.NewLegacyKeccak256()
		k.Write(src)
		expected := k.Sum(nil)

		assert.Equal(t, expected, found)
	})
}
