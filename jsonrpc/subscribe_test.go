package jsonrpc

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/panyanyany/go-web3"
	"github.com/panyanyany/go-web3/testutil"
)

func TestSubscribeNewHead(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	count := uint64(0)
	testutil.MultiAddr(t, nil, func(s *testutil.TestServer, addr string) {
		if strings.HasPrefix(addr, "http") {
			return
		}

		c, _ := NewClient(addr)
		defer c.Close()

		data := make(chan []byte)
		cancel, err := c.Subscribe("newHeads", func(b []byte) {
			data <- b
		})
		if err != nil {
			t.Fatal(err)
		}

		recv := func(ok bool) {
			count++

			select {
			case buf := <-data:
				if !ok {
					t.Fatal("unexpected value")
				}

				var block web3.Block
				if err := block.UnmarshalJSON(buf); err != nil {
					t.Fatal(err)
				}
				if block.Number != count {
					t.Fatal("bad")
				}

			case <-time.After(1 * time.Second):
				if ok {
					t.Fatal("timeout")
				}
			}
		}

		s.ProcessBlock()
		recv(true)

		s.ProcessBlock()
		recv(true)

		assert.NoError(t, cancel())

		s.ProcessBlock()
		recv(false)

		// subscription already closed
		assert.Error(t, cancel())
	})
}
