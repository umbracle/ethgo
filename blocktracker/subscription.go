package blocktracker

import (
	"sync"
)

// subscription is the Blockchain event subscription object
type subscription struct {
	updateCh chan struct{} // Channel for update information
	closeCh  chan struct{} // Channel for close signals
	elem     *eventElem    // Reference to the blockchain event wrapper
}

// GetEventCh creates a new event channel, and returns it
func (s *subscription) GetEventCh() chan *BlockEvent {
	eventCh := make(chan *BlockEvent)
	go func() {
		for {
			evnt := s.GetEvent()
			if evnt == nil {
				return
			}
			eventCh <- evnt
		}
	}()

	return eventCh
}

// GetEvent returns the event from the subscription (BLOCKING)
func (s *subscription) GetEvent() *BlockEvent {
	for {
		if s.elem.next != nil {
			s.elem = s.elem.next
			evnt := s.elem.event

			return evnt
		}

		// Wait for an update
		select {
		case <-s.updateCh:
			continue
		case <-s.closeCh:
			return nil
		}
	}
}

// Close closes the subscription
func (s *subscription) Close() {
	close(s.closeCh)
}

// Subscription is the blockchain subscription interface
type Subscription interface {
	GetEventCh() chan *BlockEvent
	GetEvent() *BlockEvent
	Close()
}

// SubscribeEvents returns a blockchain event subscription
func (b *BlockTracker) SubscribeEvents() Subscription {
	return b.stream.subscribe()
}

// eventElem contains the event, as well as the next list event
type eventElem struct {
	event *BlockEvent
	next  *eventElem
}

// eventStream is the structure that contains the event list,
// as well as the update channel which it uses to notify of updates
type eventStream struct {
	lock sync.Mutex
	head *eventElem

	// channel to notify updates
	updateCh []chan struct{}
}

// subscribe Creates a new blockchain event subscription
func (e *eventStream) subscribe() *subscription {
	head, updateCh := e.Head()
	s := &subscription{
		elem:     head,
		updateCh: updateCh,
		closeCh:  make(chan struct{}),
	}

	return s
}

// Head returns the event list head
func (e *eventStream) Head() (*eventElem, chan struct{}) {
	e.lock.Lock()
	head := e.head

	ch := make(chan struct{})
	if e.updateCh == nil {
		e.updateCh = make([]chan struct{}, 0)
	}
	e.updateCh = append(e.updateCh, ch)

	e.lock.Unlock()

	return head, ch
}

// push adds a new Event, and notifies listeners
func (e *eventStream) push(event *BlockEvent) {
	e.lock.Lock()

	newHead := &eventElem{
		event: event,
	}

	if e.head != nil {
		e.head.next = newHead
	}
	e.head = newHead

	// Notify the listeners
	for _, update := range e.updateCh {
		select {
		case update <- struct{}{}:
		default:
		}
	}

	e.lock.Unlock()
}
