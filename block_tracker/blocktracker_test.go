package blocktracker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/testutil"
)

func TestBlockTracker_Reconcile(t *testing.T) {
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
				testutil.Mock(0x1),
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

			tt := &BlockTracker{
				config:   DefaultConfig(),
				blocks:   []*ethgo.Block{},
				provider: m,
			}

			// build past block history
			for _, b := range c.History.ToBlocks() {
				tt.addBlocks(b)
			}

			for _, b := range c.Reconcile {
				blockEvnt, err := tt.HandleBlockEvent(b.block.Block())
				require.NoError(t, err)

				if b.event == nil {
					continue
				}

				// check blocks
				if !testutil.CompareBlocks(b.event.Added.ToBlocks(), blockEvnt.Added) {
					t.Fatal("err added")
				}
				if !testutil.CompareBlocks(b.event.Removed.ToBlocks(), blockEvnt.Removed) {
					t.Fatal("err removed")
				}
			}
			if !testutil.CompareBlocks(tt.blocks, c.Expected.ToBlocks()) {
				t.Fatal("bad blocks")
			}
		})
	}
}

func TestBlockTracker_Lifecycle(t *testing.T) {
	s := testutil.NewTestServer(t, func(c *testutil.TestServerConfig) {
		c.Period = 1
	})
	defer s.Close()

	c, _ := jsonrpc.NewClient(s.HTTPAddr())

	tr, err := NewBlockTracker(c.Eth())
	require.NoError(t, err)

	go tr.Start()

	sub := tr.subscribe()

	last, err := sub.Next(context.Background())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		evnt, err := sub.Next(context.Background())
		require.NoError(t, err)
		require.Equal(t, last.Added[0].Number+1, evnt.Added[0].Number)

		last = evnt
	}
}
