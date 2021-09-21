package tracker

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/blocktracker"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/jsonrpc/codec"
	"github.com/umbracle/go-web3/testutil"
	"github.com/umbracle/go-web3/tracker/store/inmem"
)

func testConfig() ConfigOption {
	return func(c *Config) {
		c.BatchSize = 10
	}
}

func testFilter(t *testing.T, provider Provider, filterConfig *FilterConfig) []*web3.Log {
	filterConfig.Async = true
	tt, _ := NewTracker(provider, WithFilter(filterConfig))

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	if err := tt.Sync(ctx); err != nil {
		t.Fatal(err)
	}

	return tt.entry.(*inmem.Entry).Logs()
}

func TestPolling(t *testing.T) {
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

	go func() {
		if err := tt.Sync(ctx); err != nil {
			panic(err)
		}
	}()

	// wait for the bulk sync to finish
	for {
		select {
		case <-tt.EventCh:
		case <-tt.DoneCh:
			goto EXIT
		case <-time.After(1 * time.Second):
			t.Fatal("timeout to sync")
		}
	}
EXIT:

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
	logs = testFilter(t, client.Eth(), &FilterConfig{Address: []web3.Address{addr0}})
	if len(logs) != 10 {
		t.Fatal("bad")
	}

	// filter by value
	typ, _ := abi.NewType("uint256")
	topic, _ := abi.EncodeTopic(typ, 1)

	logs = testFilter(t, client.Eth(), &FilterConfig{Topics: []*web3.Hash{nil, &topic}})
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
	logs := testFilter(t, client.Eth(), &FilterConfig{Topics: []*web3.Hash{&eventTopicID}})
	if len(logs) != 10 {
		t.Fatal("bad")
	}

	eventTopicID[1] = 1
	logs = testFilter(t, client.Eth(), &FilterConfig{Topics: []*web3.Hash{&eventTopicID}})
	if len(logs) != 0 {
		t.Fatal("bad")
	}
}

func TestPreflight(t *testing.T) {
	store := inmem.NewInmemStore()

	l := testutil.MockList{}
	l.Create(0, 100, func(b *testutil.MockBlock) {})

	m := &testutil.MockClient{}
	m.AddScenario(l)

	tt0, _ := NewTracker(m, testConfig(), WithStore(store))
	if err := tt0.preSyncCheckImpl(); err != nil {
		t.Fatal(err)
	}

	// change the genesis hash

	l0 := testutil.MockList{}
	l0.Create(0, 100, func(b *testutil.MockBlock) {
		b = b.Extra("1")
	})

	m.AddScenario(l0)

	tt1, _ := NewTracker(m, testConfig(), WithStore(store))
	if err := tt1.preSyncCheckImpl(); err == nil {
		t.Fatal("it should fail")
	}

	// change the chainID

	m.AddScenario(l)
	m.SetChainID(big.NewInt(1))

	tt2, _ := NewTracker(m, testConfig(), WithStore(store))
	if err := tt2.preSyncCheckImpl(); err == nil {
		t.Fatal("it should fail")
	}
}

func TestTrackerSyncerRestarts(t *testing.T) {
	store := inmem.NewInmemStore()
	m := &testutil.MockClient{}
	l := testutil.MockList{}

	advance := func(first, last int, void ...bool) {
		if len(void) == 0 {
			l.Create(first, last, func(b *testutil.MockBlock) {
				if b.GetNum()%5 == 0 {
					b = b.Log("0x1")
				}
			})
			m.AddScenario(l)
		}

		tt, err := NewTracker(m,
			testConfig(),
			WithStore(store),
			WithFilter(&FilterConfig{Async: true}),
		)
		assert.NoError(t, err)

		go func() {
			if err := tt.Sync(context.Background()); err != nil {
				panic(err)
			}
		}()

		if err := tt.WaitDuration(2 * time.Second); err != nil {
			t.Fatal(err)
		}

		if tt.blockTracker.BlocksBlocked()[0].Number != uint64(last-10) {
			t.Fatal("bad")
		}
		if tt.blockTracker.BlocksBlocked()[9].Number != uint64(last-1) {
			t.Fatal("bad")
		}
		if !testutil.CompareLogs(l.GetLogs(), tt.entry.(*inmem.Entry).Logs()) {
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
		b = b.Log("0x01")
	})

	m := &testutil.MockClient{}
	m.AddScenario(l)

	store := inmem.NewInmemStore()

	tt0, err := NewTracker(m,
		testConfig(),
		WithStore(store),
		WithFilter(&FilterConfig{Async: true}),
	)
	assert.NoError(t, err)

	go func() {
		if err := tt0.Sync(context.Background()); err != nil {
			panic(err)
		}
	}()
	tt0.WaitDuration(2 * time.Second)

	// create a fork at 'forkNum' and continue to 'endLen'
	l1 := testutil.MockList{}
	l1.Create(0, endLen, func(b *testutil.MockBlock) {
		if b.GetNum() < forkNum {
			b = b.Log("0x01") // old fork
		} else {
			if b.GetNum() == forkNum {
				b = b.Log("0x02")
			} else {
				b = b.Log("0x03")
			}
			b = b.Extra("123") // used to set the new fork
		}
	})

	m1 := &testutil.MockClient{}
	m1.AddScenario(l)
	m1.AddScenario(l1)

	tt1, _ := NewTracker(m1,
		testConfig(),
		WithStore(store),
		WithFilter(&FilterConfig{Async: true}),
	)
	go func() {
		if err := tt1.Sync(context.Background()); err != nil {
			panic(err)
		}
	}()
	tt1.WaitDuration(2 * time.Second)

	logs := tt1.entry.(*inmem.Entry).Logs()

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

func TestTrackerSyncerReconcile(t *testing.T) {
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

	store := inmem.NewInmemStore()

	for i := 0; i < n; i++ {
		// fmt.Println("########################################")

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

		go func() {
			if err := tt.Sync(context.Background()); err != nil {
				panic(err)
			}
		}()

		var added, removed []*web3.Log
		for {
			select {
			case evnt := <-tt.EventCh:
				added = append(added, evnt.Added...)
				removed = append(removed, evnt.Removed...)

			case <-tt.DoneCh:
				// no more events to read
				goto EXIT
			}
		}
	EXIT:

		// validate the included logs
		if len(added) != count {
			t.Fatal("bad added logs")
		}
		// validate the removed logs
		if len(removed) != forkSize {
			t.Fatal("bad removed logs")
		}

		// validate blocks
		if blocks := m.GetLastBlocks(backlog); !testutil.CompareBlocks(tt.blockTracker.BlocksBlocked(), blocks) {
			// tracker does not consider block 0 but getLastBlocks does return it, this is only a problem
			// with syncs on chains lower than maxBacklog
			if !testutil.CompareBlocks(blocks[1:], tt.blockTracker.BlocksBlocked()) {
				t.Fatal("bad blocks")
			}
		}
		// validate logs
		if logs := m.GetAllLogs(); !testutil.CompareLogs(tt.entry.(*inmem.Entry).Logs(), logs) {
			t.Fatal("bad logs")
		}

		c += num
	}
}

func TestTrackerSyncerRandom(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			testTrackerSyncerRandom(t, 100, uint64(randomInt(2, 10)))
		})
	}
}

func TestTrackerReconcile(t *testing.T) {
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

			store := inmem.NewInmemStore()

			btracker := blocktracker.NewBlockTracker(m)

			tt, err := NewTracker(m, WithStore(store), WithBlockTracker(btracker))
			if err != nil {
				t.Fatal(err)
			}

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
			if !testutil.CompareLogs(tt.entry.(*inmem.Entry).Logs(), c.Expected.GetLogs()) {
				t.Fatal("bad3")
			}
			if !testutil.CompareBlocks(tt.blockTracker.BlocksBlocked(), c.Expected.ToBlocks()) {
				t.Fatal("bad")
			}
		})
	}
}

type mockClientWithLimit struct {
	limit uint64
	testutil.MockClient
}

func (m *mockClientWithLimit) GetLogs(filter *web3.LogFilter) ([]*web3.Log, error) {
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
		MockClient: *m,
	}

	config := DefaultConfig()
	config.BatchSize = 11

	tt, _ := NewTracker(mm,
		WithFilter(&FilterConfig{Async: true}),
	)
	if err := tt.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if count != len(tt.entry.(*inmem.Entry).Logs()) {
		t.Fatal("not the same count")
	}
}
