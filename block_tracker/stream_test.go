package blocktracker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

func mockBlock(num uint64) *ethgo.Block {
	return &ethgo.Block{Number: num}
}

func TestStream_NonSubscribedHead(t *testing.T) {
	// Old events should be discarded from the stream if no
	// subscription is enabled.
	s := newBlockStream()
	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(1)}})
	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(2)}})

	cur := s.Head()
	require.Equal(t, cur.event.Added[0].Number, uint64(2))
}

func TestStream_SubscribeAndConsume(t *testing.T) {
	s := newBlockStream()
	cur := s.Head()

	for i := uint64(0); i < 10; i++ {
		s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(i)}})
	}

	elem := cur
	var err error

	for i := uint64(0); i < 10; i++ {
		elem, err = elem.next(context.Background())
		require.NoError(t, err)
		require.Equal(t, elem.event.Added[0].Number, i)
	}
}

func TestStream_Timeout(t *testing.T) {
	s := newBlockStream()
	cur := s.Head()

	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()

	_, err := cur.next(ctx)
	require.Error(t, err)
}

func TestStream_WaitNext(t *testing.T) {
	s := newBlockStream()
	cur := s.Head()

	doneCh := make(chan uint64)
	go func() {
		elem, _ := cur.next(context.Background())
		doneCh <- elem.event.Added[0].Number
	}()

	select {
	case <-doneCh:
		t.Fatal("it should not consume a message")
	case <-time.After(1 * time.Second):
	}

	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(10)}})

	select {
	case num := <-doneCh:
		require.Equal(t, num, uint64(10))
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

func TestStream_Flush(t *testing.T) {
	s := newBlockStream()
	cur := s.Head()

	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(10)}})
	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(11)}})
	s.push(&BlockEvent{Added: []*ethgo.Block{mockBlock(12)}})

	cur, _ = cur.next(context.Background())
	require.Equal(t, cur.event.Added[0].Number, uint64(10))

	cur = cur.flush()
	require.Equal(t, cur.event.Added[0].Number, uint64(12))
}
