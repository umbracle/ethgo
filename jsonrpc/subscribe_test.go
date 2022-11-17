package jsonrpc

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/testutil"
)

func TestSubscribeNewHead(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
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

		var lastBlock *ethgo.Block
		recv := func(ok bool) {
			select {
			case buf := <-data:
				if !ok {
					t.Fatal("unexpected value")
				}

				var block ethgo.Block
				if err := block.UnmarshalJSON(buf); err != nil {
					t.Fatal(err)
				}
				if lastBlock != nil {
					if lastBlock.Number+1 != block.Number {
						t.Fatal("bad sequence")
					}
				}
				lastBlock = &block

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
