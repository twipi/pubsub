package pubsub

// Queue implements a simple first-in-first-out queue.
// It is not thread-safe. A zero-value Queue is a valid queue.
type Queue[T any] struct {
	circ  []T
	start int
	end   int
}

// NewQueue creates a new queue with a default capacity of 16.
func NewQueue[T any]() *Queue[T] {
	return newQueue[T](16)
}

// NewQueueWithCapacity creates a new queue with the given capacity.
// The capacity is set to be at least 16.
func NewQueueWithCapacity[T any](cap int) *Queue[T] {
	return newQueue[T](min(cap, 16))
}

func newQueue[T any](cap int) *Queue[T] {
	return &Queue[T]{
		circ:  make([]T, cap),
		start: 0,
		end:   0,
	}
}

func (q *Queue[T]) len() int {
	if q.start <= q.end {
		return q.end - q.start
	}
	return len(q.circ) - q.start + q.end
}

func (q *Queue[T]) at(i int) T {
	return q.circ[(q.start+i)%len(q.circ)]
}

// grow grows the queue to double its current capacity.
func (q *Queue[T]) grow() {
	oldLen := q.len()
	newCirc := make([]T, len(q.circ)*2)
	for i := range oldLen {
		newCirc[i] = q.at(i)
	}
	q.circ = newCirc
	q.start = 0
	q.end = oldLen
}

func (q *Queue[T]) isFull() bool {
	return q.len() == len(q.circ)-1
}

// Pending returns the pending item in the queue and a boolean indicating whether
// the item was found.
func (q *Queue[T]) Pending() (T, bool) {
	if q.len() == 0 {
		var z T
		return z, false
	}
	return q.at(0), true
}

// PendingOrZero returns the pending item in the queue or the zero value if the
// queue is empty.
func (q *Queue[T]) PendingOrZero() T {
	v, _ := q.Pending()
	return v
}

// Dequeue pops the first element from the queue and discards it.
func (q *Queue[T]) Dequeue() {
	if q.len() == 0 {
		return
	}
	var z T
	q.circ[q.start] = z
	q.start = (q.start + 1) % len(q.circ)
}

// IsEmpty returns true if the queue is empty.
func (q *Queue[T]) IsEmpty() bool {
	return q.len() == 0
}

// Enqueue queues the given item.
func (q *Queue[T]) Enqueue(v T) {
	if q.isFull() {
		q.grow()
	}

	q.circ[q.end] = v
	q.end = (q.end + 1) % len(q.circ)
}
