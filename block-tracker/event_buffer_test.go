package blocktracker

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

func TestEventBufferFuzz(t *testing.T) {
	nReaders := 1000
	nMessages := 1000

	b := newEventBuffer(1000)

	// Start a write goroutine that will publish 10000 messages with sequential
	// indexes and some jitter in timing (to allow clients to "catch up" and block
	// waiting for updates).
	go func() {
		seed := time.Now().UnixNano()
		t.Logf("Using seed %d", seed)
		// z is a Zipfian distribution that gives us a number of milliseconds to
		// sleep which are mostly low - near zero but occasionally spike up to near
		// 100.
		z := rand.NewZipf(rand.New(rand.NewSource(seed)), 1.5, 1.5, 50)

		for i := 0; i < nMessages; i++ {
			// Event content is arbitrary and not valid for our use of buffers in
			// streaming - here we only care about the semantics of the buffer.
			block := &ethgo.Block{
				Number: uint64(i),
			}
			b.Append(&BlockEvent{Added: []*ethgo.Block{block}})
			// Sleep sometimes for a while to let some subscribers catch up
			wait := time.Duration(z.Uint64()) * time.Millisecond
			time.Sleep(wait)
		}
	}()

	// Run n subscribers following and verifying
	errCh := make(chan error, nReaders)

	// Load head here so all subscribers start from the same point or they might
	// not run until several appends have already happened.
	head := b.Head()

	for i := 0; i < nReaders; i++ {
		go func(i int) {
			expect := uint64(0)
			item := head
			var err error
			for {
				item, err = item.Next(context.Background(), nil)
				if err != nil {
					errCh <- fmt.Errorf("subscriber %05d failed getting next %d: %s", i,
						expect, err)
					return
				}
				if item.Events.Added[0].Number != expect {
					errCh <- fmt.Errorf("subscriber %05d got bad event want=%d, got=%d", i,
						expect, item.Events.Added[0].Number)
					return
				}
				expect++
				if expect == uint64(nMessages) {
					// Succeeded
					errCh <- nil
					return
				}
			}
		}(i)
	}

	// Wait for all readers to finish one way or other
	for i := 0; i < nReaders; i++ {
		err := <-errCh
		assert.NoError(t, err)
	}
}

func TestEventBuffer_Slow_Reader(t *testing.T) {
	b := newEventBuffer(10)

	for i := 1; i < 11; i++ {
		block := &ethgo.Block{
			Number: uint64(i),
		}
		b.Append(&BlockEvent{Added: []*ethgo.Block{block}})
	}

	require.Equal(t, 10, b.Len())

	head := b.Head()

	for i := 10; i < 15; i++ {
		block := &ethgo.Block{
			Number: uint64(i),
		}
		b.Append(&BlockEvent{Added: []*ethgo.Block{block}})
	}

	// Ensure the slow reader errors to handle dropped events and
	// fetch latest head
	ev, err := head.Next(context.Background(), nil)
	require.Error(t, err)
	require.Nil(t, ev)

	newHead := b.Head()
	require.Equal(t, 5, int(newHead.Events.Added[0].Number))
}

func TestEventBuffer_MaxSize(t *testing.T) {
	b := newEventBuffer(10)

	for i := 0; i < 100; i++ {
		block := &ethgo.Block{
			Number: uint64(i),
		}
		b.Append(&BlockEvent{Added: []*ethgo.Block{block}})
	}

	require.Equal(t, 10, b.Len())
}

// TestEventBuffer_Emptying_Buffer tests the behavior when all items
// are removed, the event buffer should advance its head down to the last message
// and insert a placeholder sentinel value.
func TestEventBuffer_Emptying_Buffer(t *testing.T) {
	b := newEventBuffer(10)

	for i := 0; i < 10; i++ {
		block := &ethgo.Block{
			Number: uint64(i),
		}
		b.Append(&BlockEvent{Added: []*ethgo.Block{block}})
	}

	require.Equal(t, 10, int(b.Len()))

	// empty the buffer, which will bring the event buffer down
	// to a single sentinel value
	for i := 0; i < 16; i++ {
		b.advanceHead()
	}

	// head and tail are now a sentinel value
	head := b.Head()
	tail := b.Tail()
	require.Equal(t, 0, b.Len())
	require.Equal(t, head, tail)

	block := &ethgo.Block{
		Number: 100,
	}
	b.Append(&BlockEvent{Added: []*ethgo.Block{block}})

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
	defer cancel()

	next, err := head.Next(ctx, make(chan struct{}))
	require.NoError(t, err)
	require.NotNil(t, next)
	require.Equal(t, uint64(100), next.Events.Added[0].Number)
}
