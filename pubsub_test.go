package pubsub_test

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/twipi/pubsub"
)

func TestSubscriber(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This test is in a separate package to ensure that the API is usable from
	// other packages.
	s := pubsub.NewSubscriber[int]()
	src := make(chan int)

	wg.Add(1)
	go func() {
		s.Listen(ctx, src)
		wg.Done()
	}()

	// Create a slow subscriber.
	slowCh := make(chan int)

	s.Subscribe(slowCh, nil)
	defer s.Unsubscribe(slowCh)

	// Test broadcast. This should not block.
	broadcast := []int{1, 2, 3, 4, 5}
	(func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		for _, n := range broadcast {
			t.Logf("broadcasting: %v", n)
			select {
			case <-ctx.Done():
				t.Fatalf("broadcast canceled: %v", ctx.Err())
			case src <- n:
			}
		}
	})()

	// Try to receive all broadcasts.
	var slowReceives []int
	for len(slowReceives) < len(broadcast) {
		select {
		case <-ctx.Done():
			return
		case n := <-slowCh:
			t.Logf("slow subscriber received: %v", n)
			slowReceives = append(slowReceives, n)
			time.Sleep(10 * time.Millisecond)
		}
	}

	if !reflect.DeepEqual(broadcast, slowReceives) {
		t.Errorf("slow subscriber did not receive all broadcasts")
	}
}
