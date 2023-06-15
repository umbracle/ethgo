package jsonrpc

import (
	"context"
	"strings"
	"testing"
	"time"

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

		ctx, cancel := context.WithCancel(context.Background())
		data := make(chan []byte)
		err := c.Subscribe(ctx, "newHeads", func(b []byte) {
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
						t.Fatalf("bad sequence %d %d", lastBlock.Number, block.Number)
					}
				}
				lastBlock = &block

			case <-time.After(1 * time.Second):
				if ok {
					t.Fatal("timeout for new head")
				}
			}
		}

		s.ProcessBlock()
		recv(true)

		s.ProcessBlock()
		recv(true)

		s.ProcessBlock()
		recv(false)

		cancel()
		close(data)
	})
}
