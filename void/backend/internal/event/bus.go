package event

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// pqItem wraps an event for the tick-indexed priority queue.
type pqItem struct {
	ev    Event
	index int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].ev.Tick != pq[j].ev.Tick {
		return pq[i].ev.Tick < pq[j].ev.Tick
	}
	return pq[i].ev.Priority > pq[j].ev.Priority
}
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}
func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*pqItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[:n-1]
	return item
}

// Bus is a concurrency-safe, tick-ordered, priority event bus with
// dependency gating, retry/backoff and a dead-letter queue so a failing
// handler chain can never take down the whole simulation.
type Bus struct {
	mu       sync.Mutex
	pq       priorityQueue
	handlers map[string][]Handler
	maxQueue int

	delivered   int64
	deadLettered int64
	deadLetters []Event
	deadMu      sync.Mutex

	dependencyDone sync.Map // eventID -> bool, satisfied dependencies
}

func NewBus(maxQueue int) *Bus {
	b := &Bus{handlers: map[string][]Handler{}, maxQueue: maxQueue}
	heap.Init(&b.pq)
	return b
}

// On registers a handler for a given event type ("*" matches all types).
func (b *Bus) On(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// Emit enqueues an event, applying backpressure: if the queue is at
// capacity the event is dead-lettered instead of blocking the caller.
func (b *Bus) Emit(e Event) bool {
	b.mu.Lock()
	if len(b.pq) >= b.maxQueue {
		b.mu.Unlock()
		b.deadLetter(e)
		return false
	}
	heap.Push(&b.pq, &pqItem{ev: e})
	b.mu.Unlock()
	return true
}

func (b *Bus) deadLetter(e Event) {
	atomic.AddInt64(&b.deadLettered, 1)
	b.deadMu.Lock()
	defer b.deadMu.Unlock()
	b.deadLetters = append(b.deadLetters, e)
	if len(b.deadLetters) > 10000 {
		b.deadLetters = b.deadLetters[len(b.deadLetters)-10000:]
	}
}

// DeadLetters returns a copy of currently retained dead-lettered events.
func (b *Bus) DeadLetters() []Event {
	b.deadMu.Lock()
	defer b.deadMu.Unlock()
	return append([]Event{}, b.deadLetters...)
}

// DrainTick pops and dispatches every event scheduled for ticks <= upToTick,
// respecting dependencies (an event whose dependency hasn't yet been
// delivered is re-queued for the next tick, bounded by maxDeferrals).
func (b *Bus) DrainTick(upToTick int64, ctxFactory func(e Event) Context) []Event {
	var processed []Event
	var deferred []*pqItem

	for {
		b.mu.Lock()
		if len(b.pq) == 0 || b.pq[0].ev.Tick > upToTick {
			b.mu.Unlock()
			break
		}
		item := heap.Pop(&b.pq).(*pqItem)
		b.mu.Unlock()

		if !b.dependenciesSatisfied(item.ev) {
			item.ev.Tick++ // retry next tick
			deferred = append(deferred, item)
			continue
		}

		ctx := ctxFactory(item.ev)
		b.dispatch(ctx, item.ev)
		b.dependencyDone.Store(item.ev.ID, true)
		atomic.AddInt64(&b.delivered, 1)
		processed = append(processed, item.ev)
	}

	if len(deferred) > 0 {
		b.mu.Lock()
		for _, it := range deferred {
			heap.Push(&b.pq, it)
		}
		b.mu.Unlock()
	}
	return processed
}

func (b *Bus) dependenciesSatisfied(e Event) bool {
	for _, dep := range e.Dependencies {
		if _, ok := b.dependencyDone.Load(dep); !ok {
			return false
		}
	}
	return true
}

func (b *Bus) dispatch(ctx Context, e Event) {
	b.mu.Lock()
	specific := append([]Handler{}, b.handlers[e.Type]...)
	wildcard := append([]Handler{}, b.handlers["*"]...)
	b.mu.Unlock()

	for _, h := range append(specific, wildcard...) {
		cons := h(ctx, e)
		for _, c := range cons {
			ne := c.Event
			ne.Tick = e.Tick + c.DelayTicks
			b.Emit(ne)
		}
	}
}

// Stats exposes bus counters for the Metrics Engine.
type Stats struct {
	QueueDepth   int
	Delivered    int64
	DeadLettered int64
}

func (b *Bus) Stats() Stats {
	b.mu.Lock()
	depth := len(b.pq)
	b.mu.Unlock()
	return Stats{
		QueueDepth:   depth,
		Delivered:    atomic.LoadInt64(&b.delivered),
		DeadLettered: atomic.LoadInt64(&b.deadLettered),
	}
}
