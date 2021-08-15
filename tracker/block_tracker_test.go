package tracker

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/testutil"
)

func testTracker(t *testing.T, server *testutil.TestServer, tracker BlockTracker) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	blocks := make(chan *web3.Block)
	err := tracker.Track(ctx, func(block *web3.Block) error {
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
		case <-time.After(2 * time.Second):
			t.Fatal("timeout")
		}
	}

	server.ProcessBlock()
	recv()

	server.ProcessBlock()
	recv()
}

func TestBlockTracker_JSONRPC(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	//fmt.Println("aaa")
	//fmt.Println(s.HTTPAddr())

	c, _ := jsonrpc.NewClient(s.HTTPAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker := NewJSONBlockTracker(logger, c.Eth())
	tracker.PollInterval = 1 * time.Second
	testTracker(t, s, tracker)
}

func TestBlockTracker_Subscription(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	//fmt.Println("bbb")
	//fmt.Println(s.WSAddr())

	c, _ := jsonrpc.NewClient(s.WSAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker, err := NewSubscriptionBlockTracker(logger, c)
	if err != nil {
		t.Fatal(err)
	}
	testTracker(t, s, tracker)
}
