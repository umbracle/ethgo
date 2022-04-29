package blocktracker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type EventBrokerCfg struct {
	EventBufferSize int64
}

type EventBroker struct {
	// mu protects subscriptions
	mu            sync.Mutex
	subscriptions map[string]*Subscription

	// eventBuf stores a configurable amount of events in memory
	eventBuf *eventBuffer

	// publishCh is used to send messages from an active txn to a goroutine which
	// publishes events, so that publishing can happen asynchronously from
	// the Commit call in the FSM hot path.
	publishCh chan *BlockEvent
}

// NewEventBroker returns an EventBroker for publishing change events.
// A goroutine is run in the background to publish events to an event buffer.
// Cancelling the context will shutdown the goroutine to free resources, and stop
// all publishing.
func NewEventBroker(ctx context.Context, cfg EventBrokerCfg) (*EventBroker, error) {
	//if cfg.Logger == nil {
	//	cfg.Logger = hclog.NewNullLogger()
	//}

	// Set the event buffer size to a minimum
	if cfg.EventBufferSize == 0 {
		cfg.EventBufferSize = 100
	}

	buffer := newEventBuffer(cfg.EventBufferSize)
	e := &EventBroker{
		//logger:    cfg.Logger.Named("event_broker"),
		eventBuf:      buffer,
		publishCh:     make(chan *BlockEvent, 64),
		subscriptions: make(map[string]*Subscription),
	}

	go e.handleUpdates(ctx)

	return e, nil
}

// Returns the current length of the event buffer
func (e *EventBroker) Len() int {
	return e.eventBuf.Len()
}

// Publish events to all subscribers of the event Topic.
func (e *EventBroker) Publish(events *BlockEvent) {
	if len(events.Added) == 0 && len(events.Removed) == 0 {
		return
	}

	e.publishCh <- events
}

// Subscribe returns a new Subscription for a given request. A Subscription
// will receive an initial empty currentItem value which points to the first item
// in the buffer. This allows the new subscription to call Next() without first checking
// for the current Item.
//
// A Subscription will start at the requested index, or as close as possible to
// the requested index if it is no longer in the buffer. If StartExactlyAtIndex is
// set and the index is no longer in the buffer or not yet in the buffer an error
// will be returned.
//
// When a caller is finished with the subscription it must call Subscription.Unsubscribe
// to free ACL tracking resources.
func (e *EventBroker) Subscribe() *Subscription {
	e.mu.Lock()
	defer e.mu.Unlock()

	head := e.eventBuf.Head()

	// Empty head so that calling Next on sub
	start := newBufferItem(&BlockEvent{})
	start.link.next.Store(head)
	close(start.link.nextCh)

	// create subscription
	id := fmt.Sprintf("%d", len(e.subscriptions))
	sub := newSubscription(start, e.unsubscribeFn(id))

	e.subscriptions[id] = sub
	return sub
}

// CloseAll closes all subscriptions
func (e *EventBroker) CloseAll() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, sub := range e.subscriptions {
		sub.forceClose()
	}
}

func (e *EventBroker) handleUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			e.CloseAll()
			return
		case update := <-e.publishCh:
			e.eventBuf.Append(update)
		}
	}
}

func (s *Subscription) forceClose() {
	if atomic.CompareAndSwapUint32(&s.state, subscriptionStateOpen, subscriptionStateClosed) {
		close(s.forceClosed)
	}
}

// unsubscribeFn returns a function that the subscription will call to remove
// itself from the subsByToken.
// This function is returned as a closure so that the caller doesn't need to keep
// track of the SubscriptionRequest, and can not accidentally call unsubscribeFn with the
// wrong pointer.
func (e *EventBroker) unsubscribeFn(id string) func() {
	return func() {
		e.mu.Lock()
		defer e.mu.Unlock()

		sub, ok := e.subscriptions[id]
		if !ok {
			return
		}

		// close the subscription
		sub.forceClose()
		delete(e.subscriptions, id)
	}
}
