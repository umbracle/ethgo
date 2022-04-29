package blocktracker

import (
	"context"
	"errors"
	"sync/atomic"
)

const (
	// subscriptionStateOpen is the default state of a subscription. An open
	// subscription may receive new events.
	subscriptionStateOpen uint32 = 0

	// subscriptionStateClosed indicates that the subscription was closed, possibly
	// as a result of a change to an ACL token, and will not receive new events.
	// The subscriber must issue a new Subscribe request.
	subscriptionStateClosed uint32 = 1
)

// ErrSubscriptionClosed is a error signalling the subscription has been
// closed. The client should Unsubscribe, then re-Subscribe.
var ErrSubscriptionClosed = errors.New("subscription closed by server, client should resubscribe")

type Subscription struct {
	// state must be accessed atomically 0 means open, 1 means closed with reload
	state uint32

	// currentItem stores the current buffer item we are on. It
	// is mutated by calls to Next.
	currentItem *bufferItem

	// forceClosed is closed when forceClose is called. It is used by
	// EventBroker to cancel Next().
	forceClosed chan struct{}

	// unsub is a function set by EventBroker that is called to free resources
	// when the subscription is no longer needed.
	// It must be safe to call the function from multiple goroutines and the function
	// must be idempotent.
	unsub func()
}

func newSubscription(item *bufferItem, unsub func()) *Subscription {
	return &Subscription{
		forceClosed: make(chan struct{}),
		currentItem: item,
		unsub:       unsub,
	}
}

func (s *Subscription) Next(ctx context.Context) (BlockEvent, error) {
	if atomic.LoadUint32(&s.state) == subscriptionStateClosed {
		return BlockEvent{}, ErrSubscriptionClosed
	}

	for {
		next, err := s.currentItem.Next(ctx, s.forceClosed)

		switch {
		case err != nil && atomic.LoadUint32(&s.state) == subscriptionStateClosed:
			return BlockEvent{}, ErrSubscriptionClosed
		case err != nil:
			return BlockEvent{}, err
		}
		s.currentItem = next

		if len(next.Events.Added) == 0 && len(next.Events.Removed) == 0 {
			continue
		}
		return BlockEvent{Added: next.Events.Added, Removed: next.Events.Removed}, nil
	}
}

func (s *Subscription) Unsubscribe() {
	s.unsub()
}
