package blocktracker

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/testutil"
)

func testListener(t *testing.T, server *testutil.TestServer, tracker BlockTrackerInterface) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	blocks := make(chan *ethgo.Block)
	err := tracker.Track(ctx, func(block *ethgo.Block) error {
		blocks <- block
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	count := uint64(0)
	recv := func() {
		count++

		select {
		case block := <-blocks:
			if block.Number != count {
				t.Fatal("bad number")
			}
		case <-time.After(4 * time.Second):
			t.Fatal("timeout")
		}
	}

	server.ProcessBlock()
	recv()

	server.ProcessBlock()
	recv()
}

func TestBlockTracker_Listener_JsonRPC(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c, _ := jsonrpc.NewClient(s.HTTPAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker := NewJSONHeadTracker(logger, c.Eth())
	tracker.PollInterval = 1 * time.Second
	testListener(t, s, tracker)
}

func TestBlockTracker_Listener_Websocket(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	c, _ := jsonrpc.NewClient(s.WSAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker, err := NewSubscriptionHeadTracker(logger, c)
	if err != nil {
		t.Fatal(err)
	}
	testListener(t, s, tracker)
}
