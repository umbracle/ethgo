package blocktracker

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	var lastBlock *ethgo.Block
	count := uint64(0)
	recv := func() {
		count++

		select {
		case block := <-blocks:
			if lastBlock != nil {
				if lastBlock.Number+1 != block.Number {
					t.Fatalf("bad sequence %d %d", lastBlock.Number, block.Number)
				}
			}
			lastBlock = block

		case <-time.After(4 * time.Second):
			t.Fatal("timeout to receive block tracker block")
		}
	}

	server.ProcessBlock()
	recv()

	server.ProcessBlock()
	recv()
}

func TestBlockTracker_Listener_JsonRPC(t *testing.T) {
	t.Skip("Too brittle on CI, FIX")

	s := testutil.NewTestServer(t)

	c, _ := jsonrpc.NewClient(s.HTTPAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker := NewJSONBlockTracker(logger, c.Eth())
	tracker.PollInterval = 1 * time.Second
	testListener(t, s, tracker)
}

func TestBlockTracker_Listener_Websocket(t *testing.T) {
	t.Skip("Too brittle on CI, FIX")

	s := testutil.NewTestServer(t)

	c, _ := jsonrpc.NewClient(s.WSAddr())
	defer c.Close()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	tracker, err := NewSubscriptionBlockTracker(logger, c)
	if err != nil {
		t.Fatal(err)
	}
	testListener(t, s, tracker)
}

func TestBlockTracker_Lifecycle(t *testing.T) {
	t.Skip()
	s := testutil.NewTestServer(t)

	c, _ := jsonrpc.NewClient(s.HTTPAddr())
	tr := NewBlockTracker(c.Eth())
	assert.NoError(t, tr.Init())

	go tr.Start()

	// try to mine a block at least every 1 second
	go func() {
		for i := 0; i < 10; i++ {
			s.ProcessBlock()
			time.After(1 * time.Second)
		}
	}()

	sub := tr.Subscribe()
	for i := 0; i < 10; i++ {
		select {
		case <-sub:
		case <-time.After(2 * time.Second):
			t.Fatal("bad")
		}
	}
}

func TestBlockTracker_PopulateBlocks(t *testing.T) {
	// more than maxBackLog blocks
	{
		l := testutil.MockList{}
		l.Create(0, 15, func(b *testutil.MockBlock) {})

		m := &testutil.MockClient{}
		m.AddScenario(l)

		tt0 := NewBlockTracker(m)

		err := tt0.Init()
		if err != nil {
			t.Fatal(err)
		}
		if !testutil.CompareBlocks(l.ToBlocks()[5:], tt0.blocks) {
			t.Fatal("bad")
		}
	}
	// less than maxBackLog
	{
		l0 := testutil.MockList{}
		l0.Create(0, 5, func(b *testutil.MockBlock) {})

		m1 := &testutil.MockClient{}
		m1.AddScenario(l0)

		tt1 := NewBlockTracker(m1)
		tt1.provider = m1

		err := tt1.Init()
		if err != nil {
			panic(err)
		}
		if !testutil.CompareBlocks(l0.ToBlocks(), tt1.blocks) {
			t.Fatal("bad")
		}
	}
}

func TestBlockTracker_Events(t *testing.T) {

	type TestEvent struct {
		Added   testutil.MockList
		Removed testutil.MockList
	}

	type Reconcile struct {
		block *testutil.MockBlock
		event *TestEvent
	}

	cases := []struct {
		Name      string
		Scenario  testutil.MockList
		History   testutil.MockList
		Reconcile []Reconcile
		Expected  testutil.MockList
	}{
		{
			Name: "Empty history",
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x1),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x1),
						},
					},
				},
			},
			Expected: []*testutil.MockBlock{
				testutil.Mock(1),
			},
		},
		{
			Name: "Repeated header",
			History: []*testutil.MockBlock{
				testutil.Mock(0x1),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x1),
				},
			},
			Expected: []*testutil.MockBlock{
				testutil.Mock(0x1),
			},
		},
		{
			Name: "New head",
			History: testutil.MockList{
				testutil.Mock(0x1),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x2),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x2),
						},
					},
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
			},
		},
		{
			Name: "Ignore block already on history",
			History: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x2),
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
			},
		},
		{
			Name: "Multi Roll back",
			History: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
				testutil.Mock(0x4),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x30).Parent(0x2),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x30).Parent(0x2),
						},
						Removed: testutil.MockList{
							testutil.Mock(0x3),
							testutil.Mock(0x4),
						},
					},
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x30).Parent(0x2),
			},
		},
		{
			Name: "Backfills missing blocks",
			Scenario: testutil.MockList{
				testutil.Mock(0x3),
				testutil.Mock(0x4),
			},
			History: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x5),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x3),
							testutil.Mock(0x4),
							testutil.Mock(0x5),
						},
					},
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
				testutil.Mock(0x4),
				testutil.Mock(0x5),
			},
		},
		{
			Name: "Rolls back and backfills",
			Scenario: testutil.MockList{
				testutil.Mock(0x30).Parent(0x2).Num(3),
				testutil.Mock(0x40).Parent(0x30).Num(4),
			},
			History: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
				testutil.Mock(0x4),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x50).Parent(0x40).Num(5),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x30).Parent(0x2).Num(3),
							testutil.Mock(0x40).Parent(0x30).Num(4),
							testutil.Mock(0x50).Parent(0x40).Num(5),
						},
						Removed: testutil.MockList{
							testutil.Mock(0x3),
							testutil.Mock(0x4),
						},
					},
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1),
				testutil.Mock(0x2),
				testutil.Mock(0x30).Parent(0x2).Num(3),
				testutil.Mock(0x40).Parent(0x30).Num(4),
				testutil.Mock(0x50).Parent(0x40).Num(5),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// safe check for now, we ma need to restart the tracker and mock client for every reconcile scenario?
			if len(c.Reconcile) != 1 {
				t.Fatal("only one reconcile supported so far")
			}

			m := &testutil.MockClient{}

			// add the full scenario with the logs
			m.AddScenario(c.Scenario)

			tt := NewBlockTracker(m)

			// build past block history
			for _, b := range c.History.ToBlocks() {
				tt.AddBlockLocked(b)
			}

			sub := tt.Subscribe()
			for _, b := range c.Reconcile {
				if err := tt.HandleReconcile(b.block.Block()); err != nil {
					t.Fatal(err)
				}

				if b.event == nil {
					continue
				}

				var blockEvnt *BlockEvent
				select {
				case blockEvnt = <-sub:
				case <-time.After(1 * time.Second):
					t.Fatal("block event timeout")
				}

				// check blocks
				if !testutil.CompareBlocks(b.event.Added.ToBlocks(), blockEvnt.Added) {
					t.Fatal("err")
				}
				if !testutil.CompareBlocks(b.event.Removed.ToBlocks(), blockEvnt.Removed) {
					t.Fatal("err")
				}
			}
			if !testutil.CompareBlocks(tt.blocks, c.Expected.ToBlocks()) {
				t.Fatal("bad")
			}
		})
	}
}
