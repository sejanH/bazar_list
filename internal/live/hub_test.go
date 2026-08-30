package live_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/sejan/bazarlist/internal/live"
)

func TestLiveHub_SubscribeAndBroadcast(t *testing.T) {
	hub := live.NewHub()
	ch, unsubscribe := hub.Subscribe(42)
	defer unsubscribe()

	payload := map[string]any{
		"id":        101,
		"name":      "Apples",
		"purchased": true,
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		hub.Broadcast(42, "item_purchased", payload)
	}()

	select {
	case event, ok := <-ch:
		if !ok {
			t.Fatal("expected event channel to be open")
		}
		if event.Type != "item_purchased" {
			t.Fatalf("expected event type item_purchased, got %s", event.Type)
		}

		var received map[string]any
		if err := json.Unmarshal(event.Data, &received); err != nil {
			t.Fatalf("failed to unmarshal event data: %v", err)
		}
		if received["name"] != "Apples" || received["purchased"] != true {
			t.Fatalf("unexpected received payload: %v", received)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for live hub broadcast")
	}
}

func TestLiveHub_MultipleSubscribersSameList(t *testing.T) {
	hub := live.NewHub()

	ch1, unsub1 := hub.Subscribe(10)
	defer unsub1()
	ch2, unsub2 := hub.Subscribe(10)
	defer unsub2()

	payload := map[string]string{"message": "hello"}
	hub.Broadcast(10, "chat", payload)

	for i, ch := range []<-chan live.Event{ch1, ch2} {
		select {
		case ev := <-ch:
			if ev.Type != "chat" {
				t.Fatalf("sub %d: expected event type 'chat', got %s", i+1, ev.Type)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("sub %d: timed out waiting for event", i+1)
		}
	}
}

func TestLiveHub_IsolationBetweenLists(t *testing.T) {
	hub := live.NewHub()

	chList1, unsub1 := hub.Subscribe(1)
	defer unsub1()
	chList2, unsub2 := hub.Subscribe(2)
	defer unsub2()

	hub.Broadcast(1, "list1_event", map[string]string{"target": "list1"})

	// chList1 should receive event
	select {
	case ev := <-chList1:
		if ev.Type != "list1_event" {
			t.Fatalf("expected list1_event, got %s", ev.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("list 1 subscriber timed out")
	}

	// chList2 should not receive anything
	select {
	case ev := <-chList2:
		t.Fatalf("list 2 subscriber unexpectedly received event: %v", ev)
	case <-time.After(50 * time.Millisecond):
		// Expected timeout
	}
}

func TestLiveHub_Unsubscribe(t *testing.T) {
	hub := live.NewHub()

	ch, unsubscribe := hub.Subscribe(99)

	// Idempotent unsubscribe calls should not panic
	unsubscribe()
	unsubscribe()

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed after unsubscribe")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}

	// Broadcast to empty list should not panic or error
	hub.Broadcast(99, "test_event", map[string]string{"test": "data"})
}

func TestLiveHub_SlowSubscriberBufferFull(t *testing.T) {
	hub := live.NewHub()
	ch, unsubscribe := hub.Subscribe(50)
	defer unsubscribe()

	// Fill buffer (buffer size is 32)
	for i := 0; i < 32; i++ {
		hub.Broadcast(50, "flood", map[string]int{"seq": i})
	}

	// 33rd event exceeds buffer and should be dropped without blocking
	done := make(chan struct{})
	go func() {
		hub.Broadcast(50, "flood_extra", map[string]int{"seq": 999})
		close(done)
	}()

	select {
	case <-done:
		// Succeeded non-blocking
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Broadcast blocked on full subscriber channel buffer")
	}

	// Drain 32 events
	for i := 0; i < 32; i++ {
		<-ch
	}
}

func TestLiveHub_ConcurrentBroadcastAndSubscribe(t *testing.T) {
	hub := live.NewHub()
	var wg sync.WaitGroup

	numSubscribers := 20
	numBroadcasts := 50

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			ch, unsub := hub.Subscribe(id % 3)
			defer unsub()

			timeout := time.After(200 * time.Millisecond)
			for {
				select {
				case <-ch:
				case <-timeout:
					return
				}
			}
		}(uint(i))
	}

	for i := 0; i < numBroadcasts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hub.Broadcast(uint(idx%3), "item_updated", map[string]int{"idx": idx})
		}(i)
	}

	wg.Wait()
}
