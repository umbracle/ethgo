package tracker

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	blocktracker "github.com/umbracle/ethgo/block-tracker"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/jsonrpc/codec"
	"github.com/umbracle/ethgo/testutil"
)

func testConfig() ConfigOption {
	return func(c *Config) {
		c.BatchSize = 10
	}
}

func testFilter(t *testing.T, provider Provider, filterConfig *FilterConfig) []*ethgo.Log {
	tt, _ := NewTracker(provider, WithFilter(filterConfig))

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	if err := tt.Sync(ctx); err != nil {
		t.Fatal(err)
	}

	return tt.entry.(*inmemEntry).Logs()
}

func TestPolling(t *testing.T) {
	t.Skip()

	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	client, _ := jsonrpc.NewClient(s.HTTPAddr())

	c0 := &testutil.Contract{}
	c0.AddEvent(testutil.NewEvent("A").Add("uint256", true).Add("uint256", true))
	c0.EmitEvent("setA1", "A", "1", "2")

	_, addr0 := s.DeployContract(c0)

	// send 5 txns
	for i := 0; i < 5; i++ {
		s.TxnTo(addr0, "setA1")
	}

	tt, err := NewTracker(client.Eth())
	assert.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	err = tt.syncImpl(ctx)
	assert.NoError(t, err)

	/*
		// send another 5 transactions, we have to have another log each time
		for i := 0; i < 5; i++ {
			receipt := s.TxnTo(addr0, "setA1")

			select {
			case evnt := <-tt.EventCh:
				if !reflect.DeepEqual(evnt.Added, receipt.Logs) {
					t.Fatal("bad")
				}
			case <-time.After(2 * time.Second): // wait at least the polling interval
				t.Fatal("event expected")
			}
		}
	*/
}

func TestFilterIntegration(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	client, _ := jsonrpc.NewClient(s.HTTPAddr())

	c0 := &testutil.Contract{}
	c0.AddEvent(testutil.NewEvent("A").Add("uint256", true).Add("uint256", true))
	c0.EmitEvent("setA1", "A", "1", "2")

	_, addr0 := s.DeployContract(c0)
	_, addr1 := s.DeployContract(c0)

	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			s.TxnTo(addr0, "setA1")
		} else {
			s.TxnTo(addr1, "setA1")
		}
	}

	// sync all the logs
	logs := testFilter(t, client.Eth(), &FilterConfig{})
	if len(logs) != 20 {
		t.Fatal("bad")
	}

	// filter by address
	logs = testFilter(t, client.Eth(), &FilterConfig{Address: []ethgo.Address{addr0}})
	if len(logs) != 10 {
		t.Fatal("bad")
	}

	// filter by value
	typ, _ := abi.NewType("uint256")
	topic, _ := abi.EncodeTopic(typ, 1)

	logs = testFilter(t, client.Eth(), &FilterConfig{Topics: [][]*ethgo.Hash{{nil, &topic}}})
	if len(logs) != 20 {
		t.Fatal("bad")
	}
}

func TestFilterIntegrationEventHash(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	client, _ := jsonrpc.NewClient(s.HTTPAddr())

	c0 := &testutil.Contract{}
	c0.AddEvent(testutil.NewEvent("A").Add("uint256", true).Add("uint256", true))
	c0.EmitEvent("setA1", "A", "1", "2")

	c1 := &testutil.Contract{}
	c1.AddEvent(testutil.NewEvent("B").Add("uint256", true).Add("uint256", true))
	c1.EmitEvent("setB1", "B", "1", "2")

	artifacts0, addr0 := s.DeployContract(c0)
	_, addr1 := s.DeployContract(c0)

	abi0, _ := abi.NewABI(artifacts0.Abi)

	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			s.TxnTo(addr0, "setA1")
		} else {
			s.TxnTo(addr1, "setB1")
		}
	}

	eventTopicID := abi0.Events["A"].ID()
	logs := testFilter(t, client.Eth(), &FilterConfig{Topics: [][]*ethgo.Hash{{&eventTopicID}}})
	if len(logs) != 10 {
		t.Fatal("bad")
	}

	eventTopicID[1] = 1
	logs = testFilter(t, client.Eth(), &FilterConfig{Topics: [][]*ethgo.Hash{{&eventTopicID}}})
	if len(logs) != 0 {
		t.Fatal("bad")
	}
}

func TestTracker_Sync_Restart(t *testing.T) {
	// 10 blocks of backlog

	store := NewInmemStore()
	m := &testutil.MockClient{}
	l := testutil.MockList{}

	advance := func(first, last int, void ...bool) {
		if len(void) == 0 {
			l.Create(first, last, func(b *testutil.MockBlock) {
				if b.GetNum()%5 == 0 {
					b.Log("0x1")
				}
			})
			m.AddScenario(l)
		}

		tt, err := NewTracker(m,
			testConfig(),
			WithStore(store),
			WithFilter(&FilterConfig{}),
			WithMaxBacklog(10),
		)
		assert.NoError(t, err)

		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()

		err = tt.syncImpl(ctx)
		assert.NoError(t, err)

		if !testutil.CompareLogs(l.GetLogs(), tt.entry.(*inmemEntry).Logs()) {
			t.Fatal("bad")
		}
	}

	// initial range
	advance(0, 100)

	// dont advance
	advance(0, 100, true)

	// advance less than backlog
	advance(100, 105)

	// advance more than backlog
	advance(105, 150)
}

func testSyncerReconcile(t *testing.T, iniLen, forkNum, endLen int) {
	// test that the syncer can reconcile if there is a fork in the saved state
	l := testutil.MockList{}
	l.Create(0, iniLen, func(b *testutil.MockBlock) {
		b.Log("0x01")
	})

	m := &testutil.MockClient{}
	m.AddScenario(l)

	store := NewInmemStore()

	tt0, err := NewTracker(m,
		testConfig(),
		WithStore(store),
		WithFilter(&FilterConfig{}),
	)
	assert.NoError(t, err)

	err = tt0.syncImpl(context.TODO())
	assert.NoError(t, err)

	// create a fork at 'forkNum' and continue to 'endLen'
	l1 := testutil.MockList{}
	l1.Create(0, endLen, func(b *testutil.MockBlock) {
		if b.GetNum() < forkNum {
			b.Log("0x01") // old fork
		} else {
			if b.GetNum() == forkNum {
				b = b.Log("0x02")
			} else {
				b = b.Log("0x03")
			}
			b.Extra("123") // used to set the new fork
		}
	})

	m1 := &testutil.MockClient{}
	m1.AddScenario(l)
	m1.AddScenario(l1)

	tt1, _ := NewTracker(m1,
		testConfig(),
		WithStore(store),
		WithFilter(&FilterConfig{}),
	)

	err = tt1.syncImpl(context.Background())
	assert.NoError(t, err)

	logs := tt1.entry.(*inmemEntry).Logs()

	if !testutil.CompareLogs(l1.GetLogs(), logs) {
		t.Fatal("bad")
	}

	// check the content of the logs

	// first half
	for i := 0; i < forkNum; i++ {
		if logs[i].Data[0] != 0x1 {
			t.Fatal("bad")
		}
	}
	// fork point
	if logs[forkNum].Data[0] != 0x2 {
		t.Fatal("bad")
	}
	// second half
	for i := forkNum + 1; i < endLen; i++ {
		if logs[i].Data[0] != 0x3 {
			t.Fatal("bad")
		}
	}
}

func TestTracker_Sync_Reconcile(t *testing.T) {
	t.Run("Backlog", func(t *testing.T) {
		testSyncerReconcile(t, 50, 45, 55)
	})
	t.Run("Historical", func(t *testing.T) {
		testSyncerReconcile(t, 50, 45, 100)
	})
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func testTrackerSyncerRandom(t *testing.T, n int, backlog uint64) {
	m := &testutil.MockClient{}
	c := 0 // current block
	f := 0 // current fork

	store := NewInmemStore()

	for i := 0; i < n; i++ {
		// create the new batch of blocks
		var forkSize int
		if randomInt(0, 10) < 3 && c > 10 {
			// add a fork, go back at most maxBacklogSize
			forkSize = randomInt(1, int(backlog))
			c = c - forkSize
			f++
		}

		forkID := strconv.Itoa(f)

		// add new blocks
		l := testutil.MockList{}

		// we have to create at least the blocks removed by the fork, otherwise
		// we may end up going backwards if the forks remove more data than the
		// advance includes

		start := forkSize
		if start == 0 && i == 0 {
			start = 1 // at least advance one block on the first iteration
		}
		num := randomInt(start, 20)
		count := 0

		for j := c; j < c+num; j++ {
			bb := testutil.Mock(j).Extra(forkID)
			if j != 0 {
				count++
				bb = bb.Log(forkID)
			}
			l = append(l, bb)
		}

		m.AddScenario(l)

		// use a custom block tracker to add specific backlog
		tracker := blocktracker.NewBlockTracker(m, blocktracker.WithBlockMaxBacklog(backlog))

		tt, _ := NewTracker(m,
			testConfig(),
			WithStore(store),
			WithBlockTracker(tracker),
		)

		err := tt.syncImpl(context.Background())
		assert.NoError(t, err)

		// validate logs
		if logs := m.GetAllLogs(); !testutil.CompareLogs(tt.entry.(*inmemEntry).Logs(), logs) {
			t.Fatal("bad logs")
		}

		c += num
	}
}

func TestTracker_Sync_Random(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			testTrackerSyncerRandom(t, 100, uint64(randomInt(2, 10)))
		})
	}
}

func TestTracker_Reconcile(t *testing.T) {
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
		/*
			{
				Name: "Empty history",
				Reconcile: []Reconcile{
					{
						block: testutil.Mock(0x1).Log("0x1"),
						event: &TestEvent{
							Added: testutil.MockList{
								testutil.Mock(0x1).Log("0x1"),
							},
						},
					},
				},
				Expected: []*testutil.MockBlock{
					testutil.Mock(1).Log("0x1"),
				},
			},
		*/
		/*
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
		*/
		/*
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
		*/
		/*
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
						event: &TestEvent{},
					},
				},
				Expected: testutil.MockList{
					testutil.Mock(0x1),
					testutil.Mock(0x2),
					testutil.Mock(0x3),
				},
			},
		*/
		/*
			{
				Name: "Multi Roll back",
				History: testutil.MockList{
					testutil.Mock(0x1),
					testutil.Mock(0x2),
					testutil.Mock(0x3).Log("0x3"),
					testutil.Mock(0x4).Log("0x4"),
				},
				Reconcile: []Reconcile{
					{
						block: testutil.Mock(0x30).Parent(0x2).Log("0x30"),
						event: &TestEvent{
							Added: testutil.MockList{
								testutil.Mock(0x30).Parent(0x2).Log("0x30"),
							},
							Removed: testutil.MockList{
								testutil.Mock(0x3).Log("0x3"),
								testutil.Mock(0x4).Log("0x4"),
							},
						},
					},
				},
				Expected: testutil.MockList{
					testutil.Mock(0x1),
					testutil.Mock(0x2),
					testutil.Mock(0x30).Parent(0x2).Log("0x30"),
				},
			},
		*/

		{
			Name: "Backfills missing blocks",
			Scenario: testutil.MockList{
				testutil.Mock(0x3),
				testutil.Mock(0x4).Log("0x2"),
			},
			History: testutil.MockList{
				testutil.Mock(0x1).Log("0x1"),
				testutil.Mock(0x2),
			},
			Reconcile: []Reconcile{
				{
					block: testutil.Mock(0x5).Log("0x3"),
					event: &TestEvent{
						Added: testutil.MockList{
							testutil.Mock(0x3),
							testutil.Mock(0x4).Log("0x2"),
							testutil.Mock(0x5).Log("0x3"),
						},
					},
				},
			},
			Expected: testutil.MockList{
				testutil.Mock(0x1).Log("0x1"),
				testutil.Mock(0x2),
				testutil.Mock(0x3),
				testutil.Mock(0x4).Log("0x2"),
				testutil.Mock(0x5).Log("0x3"),
			},
		},
		/*
			{
				Name: "Rolls back and backfills",
				Scenario: testutil.MockList{
					testutil.Mock(0x30).Parent(0x2).Num(3).Log("0x5"),
					testutil.Mock(0x40).Parent(0x30).Num(4),
				},
				History: testutil.MockList{
					testutil.Mock(0x1),
					testutil.Mock(0x2).Log("0x3"),
					testutil.Mock(0x3).Log("0x2"),
					testutil.Mock(0x4).Log("0x1"),
				},
				Reconcile: []Reconcile{
					{
						block: testutil.Mock(0x50).Parent(0x40).Num(5),
						event: &TestEvent{
							Added: testutil.MockList{
								testutil.Mock(0x30).Parent(0x2).Num(3).Log("0x5"),
								testutil.Mock(0x40).Parent(0x30).Num(4),
								testutil.Mock(0x50).Parent(0x40).Num(5),
							},
							Removed: testutil.MockList{
								testutil.Mock(0x3).Log("0x2"),
								testutil.Mock(0x4).Log("0x1"),
							},
						},
					},
				},
				Expected: testutil.MockList{
					testutil.Mock(0x1),
					testutil.Mock(0x2).Log("0x3"),
					testutil.Mock(0x30).Parent(0x2).Num(3).Log("0x5"),
					testutil.Mock(0x40).Parent(0x30).Num(4),
					testutil.Mock(0x50).Parent(0x40).Num(5),
				},
			},
		*/
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

			// add the logs of the reconcile block because those are also unknown for the tracker
			m.AddLogs(c.Reconcile[0].block.GetLogs())

			store := NewInmemStore().(*inmemEntry)

			btracker := blocktracker.NewBlockTracker(m)

			tt, err := NewTracker(m, WithStore(store), WithBlockTracker(btracker))
			assert.NoError(t, err)

			for _, b := range c.History {
				// add all the history blocks to the tracker
				err = btracker.AddBlocksLocked(b.Block())
				assert.NoError(t, err)

				// add all the logs to the store
				store.storeLogs(b.GetLogs())
			}

			// compute the reconcile and check the result
			for _, b := range c.Reconcile {
				bEvent, err := btracker.HandleBlockEvent(b.block.Block())
				assert.NoError(t, err)

				event, err := tt.handleBlockEvent(bEvent)
				assert.NoError(t, err)

				if event == nil {
					continue
				}

				if !testutil.CompareLogs(b.event.Added.GetLogs(), event.Added) {
					t.Fatal("incorrect added logs")
				}
				if !testutil.CompareLogs(b.event.Removed.GetLogs(), event.Removed) {

					fmt.Println(b.event.Removed.GetLogs())
					fmt.Println(event.Removed)

					t.Fatal("incorrect removed logs")
				}
			}

			/*
				// important to set a buffer here, otherwise everything is blocked
				tt.EventCh = make(chan *Event, 1)

				// set the filter as synced since we only want to
				// try reconciliation
				tt.synced = 1

				// build past block history
				for _, b := range c.History.ToBlocks() {
					tt.blockTracker.AddBlockLocked(b)
				}
				// add the history to the store
				for _, b := range c.History {
					tt.entry.StoreLogs(b.GetLogs())
				}

				for _, b := range c.Reconcile {
					aux, err := tt.blockTracker.HandleBlockEvent(b.block.Block())
					if err != nil {
						t.Fatal(err)
					}
					if aux == nil {
						continue
					}
					if err := tt.handleBlockEvnt(aux); err != nil {
						t.Fatal(err)
					}

					var evnt *Event
					select {
					case evnt = <-tt.EventCh:
					case <-time.After(1 * time.Second):
						t.Fatal("log event timeout")
					}

					// check logs
					if !testutil.CompareLogs(b.event.Added.GetLogs(), evnt.Added) {
						t.Fatal("err")
					}
					if !testutil.CompareLogs(b.event.Removed.GetLogs(), evnt.Removed) {
						t.Fatal("err")
					}

					var blockEvnt *blocktracker.BlockEvent
					select {
					case blockEvnt = <-tt.BlockCh:
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

				// check the post state (logs and blocks) after all the reconcile events
				if !testutil.CompareLogs(tt.entry.(*inmemEntry).Logs(), c.Expected.GetLogs()) {
					t.Fatal("bad3")
				}
				if !testutil.CompareBlocks(tt.blockTracker.BlocksBlocked(), c.Expected.ToBlocks()) {
					t.Fatal("bad")
				}
			*/
		})
	}
}

type mockClientWithLimit struct {
	limit uint64
	*testutil.MockClient
}

func (m *mockClientWithLimit) GetLogs(filter *ethgo.LogFilter) ([]*ethgo.Log, error) {
	if filter.BlockHash != nil {
		return m.MockClient.GetLogs(filter)
	}
	from, to := uint64(*filter.From), uint64(*filter.To)
	if from > to {
		return nil, fmt.Errorf("from higher than to")
	}
	if to-from > m.limit {
		return nil, &codec.ErrorObject{Message: "query returned more than 10000 results"}
	}
	// fallback to the client
	return m.MockClient.GetLogs(filter)
}

func TestTooMuchDataRequested(t *testing.T) {
	count := 0

	// create 100 blocks with 2 (even) or 5 (odd) logs each
	l := testutil.MockList{}
	l.Create(0, 100, func(b *testutil.MockBlock) {
		var numLogs int
		if b.GetNum()%2 == 0 {
			numLogs = 2
		} else {
			numLogs = 5
		}
		for i := 0; i < numLogs; i++ {
			count++
			b.Log("0x1")
		}
	})

	m := &testutil.MockClient{}
	m.AddScenario(l)

	mm := &mockClientWithLimit{
		limit:      3,
		MockClient: m,
	}

	config := DefaultConfig()
	config.BatchSize = 11

	tt, _ := NewTracker(mm,
		WithFilter(&FilterConfig{}),
	)
	if err := tt.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if count != len(tt.entry.(*inmemEntry).Logs()) {
		t.Fatal("not the same count")
	}
}
