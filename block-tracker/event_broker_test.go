package blocktracker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

func TestEventBroker_PublishAndSubscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	publisher, err := NewEventBroker(ctx, EventBrokerCfg{EventBufferSize: 100})
	require.NoError(t, err)

	sub := publisher.Subscribe()
	eventCh := consumeSubscription(ctx, sub)

	// Now subscriber should block waiting for updates
	assertNoResult(t, eventCh)

	block := &ethgo.Block{
		Number: uint64(100),
	}
	publisher.Publish(&BlockEvent{Added: []*ethgo.Block{block}})

	// Subscriber should see the published event
	result := nextResult(t, eventCh)
	require.NoError(t, result.Err)
	require.Equal(t, uint64(100), result.Events.Added[0].Number)

	// Now subscriber should block waiting for updates
	assertNoResult(t, eventCh)

	// Publish a second event
	block = &ethgo.Block{
		Number: uint64(200),
	}
	publisher.Publish(&BlockEvent{Added: []*ethgo.Block{block}})

	result = nextResult(t, eventCh)
	require.NoError(t, result.Err)
	require.Equal(t, uint64(200), result.Events.Added[0].Number)
}

func TestEventBroker_ShutdownClosesSubscriptions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	publisher, err := NewEventBroker(ctx, EventBrokerCfg{})
	require.NoError(t, err)

	sub1 := publisher.Subscribe()
	defer sub1.Unsubscribe()

	sub2 := publisher.Subscribe()
	defer sub2.Unsubscribe()

	cancel() // Shutdown

	err = consumeSub(context.Background(), sub1)
	require.Equal(t, err, ErrSubscriptionClosed)

	_, err = sub2.Next(context.Background())
	require.Equal(t, err, ErrSubscriptionClosed)
}

func consumeSubscription(ctx context.Context, sub *Subscription) <-chan subNextResult {
	eventCh := make(chan subNextResult, 1)
	go func() {
		for {
			es, err := sub.Next(ctx)
			eventCh <- subNextResult{
				Events: &es,
				Err:    err,
			}
			if err != nil {
				return
			}
		}
	}()
	return eventCh
}

type subNextResult struct {
	Events *BlockEvent
	Err    error
}

func nextResult(t *testing.T, eventCh <-chan subNextResult) subNextResult {
	t.Helper()
	select {
	case next := <-eventCh:
		return next
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("no event after 100ms")
	}
	return subNextResult{}
}

func assertNoResult(t *testing.T, eventCh <-chan subNextResult) {
	t.Helper()
	select {
	case next := <-eventCh:
		require.NoError(t, next.Err)
		//require.Len(t, next.Events, 1)
		t.Fatalf("received unexpected event: %#v", next.Events)
	case <-time.After(100 * time.Millisecond):
	}
}

func consumeSub(ctx context.Context, sub *Subscription) error {
	for {
		_, err := sub.Next(ctx)
		if err != nil {
			return err
		}
	}
}
