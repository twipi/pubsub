package pubsub

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestQueue(t *testing.T) {
	q := newQueue[int](2)
	q.Enqueue(1)
	q.Enqueue(2)

	assert.Equal(t, 4, len(q.circ), "capacity")

	assert.Equal(t, 1, q.PendingOrZero(), "first item")
	q.Dequeue()

	assert.Equal(t, false, q.IsEmpty(), "not empty")

	q.Enqueue(3)

	assert.Equal(t, 2, q.PendingOrZero(), "second item")
	q.Dequeue()

	assert.Equal(t, 3, q.PendingOrZero(), "third item")
	q.Dequeue()

	q.Enqueue(4)
	q.Enqueue(5)
	q.Enqueue(6)

	assert.Equal(t, 4, len(q.circ), "capacity after grow")

	assert.Equal(t, 4, q.PendingOrZero(), "fourth item")
	q.Dequeue()

	assert.Equal(t, 5, q.PendingOrZero(), "fifth item")
	q.Dequeue()

	assert.Equal(t, 6, q.PendingOrZero(), "sixth item")
	q.Dequeue()

	assert.Equal(t, true, q.IsEmpty(), "empty")

	q.Dequeue()
	q.Dequeue()
	assert.Equal(t, true, q.IsEmpty(), "empty after invalid dequeues")
}
