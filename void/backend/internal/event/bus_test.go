package event

import "testing"

func TestEmitAndDrain(t *testing.T) {
	b := NewBus(100)
	delivered := 0
	b.On("test_event", func(ctx Context, e Event) []Consequence {
		delivered++
		return nil
	})
	b.Emit(Event{Type: "test_event", Tick: 1})
	b.Emit(Event{Type: "test_event", Tick: 1})
	b.Emit(Event{Type: "test_event", Tick: 2})

	processed := b.DrainTick(1, func(e Event) Context { return Context{} })
	if len(processed) != 2 {
		t.Fatalf("expected 2 events processed at tick 1, got %d", len(processed))
	}
	if delivered != 2 {
		t.Fatalf("expected handler called twice, got %d", delivered)
	}

	processed = b.DrainTick(2, func(e Event) Context { return Context{} })
	if len(processed) != 1 {
		t.Fatalf("expected 1 event at tick 2, got %d", len(processed))
	}
}

func TestBackpressureDeadLetters(t *testing.T) {
	b := NewBus(1)
	ok1 := b.Emit(Event{Type: "a", Tick: 1})
	ok2 := b.Emit(Event{Type: "b", Tick: 1})
	if !ok1 {
		t.Fatal("first emit should succeed")
	}
	if ok2 {
		t.Fatal("second emit should be dead-lettered at capacity 1")
	}
	if len(b.DeadLetters()) != 1 {
		t.Fatalf("expected 1 dead letter, got %d", len(b.DeadLetters()))
	}
}

func TestDependencyGating(t *testing.T) {
	b := NewBus(100)
	var order []string
	b.On("*", func(ctx Context, e Event) []Consequence {
		order = append(order, e.ID)
		return nil
	})
	b.Emit(Event{ID: "child", Type: "x", Tick: 1, Dependencies: []string{"parent"}})
	b.Emit(Event{ID: "parent", Type: "x", Tick: 1})

	b.DrainTick(1, func(e Event) Context { return Context{} })
	b.DrainTick(2, func(e Event) Context { return Context{} })

	if len(order) != 2 || order[0] != "parent" || order[1] != "child" {
		t.Fatalf("expected parent before child, got %v", order)
	}
}
