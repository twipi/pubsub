// Package pubsub implements an unbounded channel for the pub/sub pattern.
package pubsub

import (
	"context"
	"sync"
)

// Subscriber is a subscriber that subscribes to a Pipe. A zero-value Subscriber
// is a valid Subscriber.
type Subscriber[T any] struct {
	mu   sync.RWMutex
	subs map[chan<- T]pipeSub[T]
}

type pipeSub[T any] struct {
	filter func(T) bool
	queue  chan<- T // unbounded FIFO queue
	stop   chan struct{}
}

// NewSubscriber creates a new Subscriber.
func NewSubscriber[T any]() *Subscriber[T] {
	return &Subscriber[T]{}
}

// Listen starts broadcasting messages received from the given src channel.
// It blocks until the src channel is closed or ctx is canceled.
func (s *Subscriber[T]) Listen(ctx context.Context, src <-chan T) error {
	s.mu.Lock()
	if s.subs == nil {
		s.subs = make(map[chan<- T]pipeSub[T])
	}
	s.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-src:
			if !ok {
				return nil
			}
			s.publish(ctx, msg)
		}
	}
}

func (s *Subscriber[T]) publish(ctx context.Context, value T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sub := range s.subs {
		if !sub.filter(value) {
			continue
		}

		// We're ok with waiting for sub.queue to accept the value, since
		// they go to unbounded queues.
		select {
		case <-ctx.Done():
			return
		case <-sub.stop:
			return
		case sub.queue <- value:
			// ok
		}
	}
}

// FilterFunc is a filter function for any type.
// If the function returns true, the message will be sent to the subscriber.
// If the function is nil, all messages will be sent to the subscriber.
type FilterFunc[T any] func(T) bool

// Subscribe subscribes ch to incoming messages from the given recipient.
// Calls to Subscribe should always be paired with Unsubscribe. It is
// recommended to use defer.
//
// Subscribe panics if it's called on a Subscriber w/ a src that's already
// closed.
func (s *Subscriber[T]) Subscribe(ch chan<- T, filter FilterFunc[T]) {
	if filter == nil {
		filter = func(T) bool { return true }
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subs == nil {
		s.subs = make(map[chan<- T]pipeSub[T])
	}

	_, ok := s.subs[ch]
	if ok {
		panic("twipi: channel already subscribed")
	}

	queue := make(chan T, 1)
	stop := make(chan struct{})

	go func() {
		var dst chan<- T
		pending := NewQueue[T]()

		for {
			select {
			case msg := <-queue:
				pending.Enqueue(msg)
				dst = ch

			case dst <- pending.PendingOrZero():
				pending.Dequeue()
				// If the pending queue is drained, then stop sending.
				if pending.IsEmpty() {
					dst = nil
				}

			case <-stop:
				close(ch)
				return
			}
		}
	}()

	s.subs[ch] = pipeSub[T]{
		filter: filter,
		queue:  queue,
		stop:   stop,
	}
}

// Unsubscribe unsubscribes ch from incoming messages from all its recipients.
// Once unsubscribed, ch will be closed.
func (s *Subscriber[T]) Unsubscribe(ch chan<- T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.unsub(ch)
}

func (s *Subscriber[T]) unsub(ch chan<- T) {
	sub, ok := s.subs[ch]
	if !ok {
		return
	}

	close(sub.stop)
	delete(s.subs, ch)
}
