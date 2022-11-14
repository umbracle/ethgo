package blocktracker

import (
	"context"
	"sync"
)

// blockStream is used to keep the stream of new block events and allow subscriptions
// of the stream at any point
type blockStream struct {
	lock sync.Mutex
	head *headElem
}

func newBlockStream() *blockStream {
	b := &blockStream{}
	b.push(&BlockEvent{})
	return b
}

func (b *blockStream) Head() *headElem {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.head
}

func (b *blockStream) push(event *BlockEvent) {
	b.lock.Lock()
	defer b.lock.Unlock()

	newHead := newHeadElem(event)
	if b.head != nil {
		b.head.link.next = newHead
		close(b.head.link.nextCh)
	}

	b.head = newHead
}

type headElem struct {
	event *BlockEvent
	link  *headLink
}

func newHeadElem(event *BlockEvent) *headElem {
	return &headElem{
		event: event,
		link: &headLink{
			nextCh: make(chan struct{}),
		},
	}
}

type headLink struct {
	// next is a refernece to the next item
	next *headElem

	// nextCh is closed when the next item is ready
	nextCh chan struct{}
}

func (h *headElem) flush() *headElem {
	cur := h

	for cur.link.next != nil {
		cur = cur.link.next
	}

	return cur
}

func (h *headElem) next(ctx context.Context) (*headElem, error) {
	cur := h

	if cur.link.next == nil {
		// wait for the item to be ready
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-cur.link.nextCh:
		}
	}

	cur = cur.link.next
	return cur, nil
}
