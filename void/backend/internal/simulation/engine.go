// Package simulation implements VOID's Simulation Engine: a highly
// concurrent, deterministic tick loop that evaluates millions of entities'
// behaviors in parallel via a bounded worker pool, feeds resulting actions
// into the Event Engine, and reports metrics every tick.
package simulation

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"void/internal/behavior"
	"void/internal/entity"
	"void/internal/event"
	"void/internal/metrics"
	"void/internal/rng"
	"void/internal/world"
)

// Status enumerates simulation lifecycle states.
type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusPaused  Status = "paused"
	StatusStopped Status = "stopped"
	StatusDone    Status = "done"
)

// Config controls one simulation run.
type Config struct {
	ID          string
	WorldID     string
	Seed        int64
	MaxTicks    int64
	WorkerCount int
	TimeScale   float64 // simulated-seconds per tick
	BatchSize   int     // entities per worker batch (memory pooling unit)
	Headless    bool
}

// Engine owns one running simulation: the entity population, event bus,
// behavior engine, and the goroutine worker pool that ticks it forward.
type Engine struct {
	cfg      Config
	world    *world.World
	entities []*entity.Entity
	entIndex sync.Map // id -> *entity.Entity, for O(1) lookups from handlers

	bus      *event.Bus
	behavior *behavior.Engine
	metrics  *metrics.Registry
	root     *rng.Source

	status   atomic.Value // Status
	tick     int64
	mu       sync.RWMutex
	pauseCh  chan struct{}
	stopCh   chan struct{}
	stepOnce chan struct{}

	onTick func(tick int64, processed []event.Event)
}

// New constructs an Engine ready to Run.
func New(cfg Config, w *world.World, beh *behavior.Engine, m *metrics.Registry) *Engine {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 8
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	e := &Engine{
		cfg:      cfg,
		world:    w,
		behavior: beh,
		metrics:  m,
		bus:      event.NewBus(2_000_000), // backpressure ceiling
		root:     rng.New(cfg.Seed),
		pauseCh:  make(chan struct{}, 1),
		stopCh:   make(chan struct{}),
		stepOnce: make(chan struct{}, 1),
	}
	e.status.Store(StatusPending)
	return e
}

// AddEntity registers an entity into the simulation's live population.
func (e *Engine) AddEntity(ent *entity.Entity) {
	e.mu.Lock()
	e.entities = append(e.entities, ent)
	e.mu.Unlock()
	e.entIndex.Store(ent.ID, ent)
}

// Bus exposes the event bus (for API subscriptions / handler registration).
func (e *Engine) Bus() *event.Bus { return e.bus }

// OnTick registers a callback invoked after every tick with the events that
// were processed (used to stream to WebSocket clients / storage).
func (e *Engine) OnTick(fn func(tick int64, processed []event.Event)) { e.onTick = fn }

func (e *Engine) Status() Status { return e.status.Load().(Status) }
func (e *Engine) Tick() int64    { return atomic.LoadInt64(&e.tick) }

func (e *Engine) Pause() {
	if e.Status() == StatusRunning {
		e.status.Store(StatusPaused)
	}
}
func (e *Engine) Resume() {
	if e.Status() == StatusPaused {
		e.status.Store(StatusRunning)
		select {
		case e.pauseCh <- struct{}{}:
		default:
		}
	}
}
func (e *Engine) Stop() {
	e.status.Store(StatusStopped)
	select {
	case <-e.stopCh:
	default:
		close(e.stopCh)
	}
}

// Step advances exactly one tick while paused (Step-by-Step execution).
func (e *Engine) Step() {
	select {
	case e.stepOnce <- struct{}{}:
	default:
	}
}

// Run drives the simulation forward until MaxTicks or ctx cancellation.
// Entity behavior evaluation for each tick is distributed across
// cfg.WorkerCount goroutines pulling fixed-size batches (memory pooling via
// batch reuse) from the population, giving predictable, bounded resource
// use even with tens of millions of entities.
func (e *Engine) Run(ctx context.Context) error {
	e.status.Store(StatusRunning)
	batches := e.makeBatches()

	for atomic.LoadInt64(&e.tick) < e.cfg.MaxTicks || e.cfg.MaxTicks == 0 {
		select {
		case <-ctx.Done():
			e.status.Store(StatusStopped)
			return ctx.Err()
		case <-e.stopCh:
			return nil
		default:
		}

		if e.Status() == StatusPaused {
			select {
			case <-e.pauseCh:
			case <-e.stepOnce:
			case <-e.stopCh:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		start := time.Now()
		tick := atomic.AddInt64(&e.tick, 1)
		e.evaluateTickConcurrent(tick, batches)

		processed := e.bus.DrainTick(tick, e.contextFactory)
		if e.onTick != nil {
			e.onTick(tick, processed)
		}

		e.recordTickMetrics(tick, start, len(processed))

		if e.cfg.MaxTicks > 0 && tick >= e.cfg.MaxTicks {
			e.status.Store(StatusDone)
			return nil
		}
	}
	return nil
}

// makeBatches partitions the entity slice into fixed-size chunks once, so
// the worker pool can reuse the same slice headers every tick (no
// re-allocation -> memory pooling).
func (e *Engine) makeBatches() [][]*entity.Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	n := len(e.entities)
	var batches [][]*entity.Entity
	for i := 0; i < n; i += e.cfg.BatchSize {
		end := i + e.cfg.BatchSize
		if end > n {
			end = n
		}
		batches = append(batches, e.entities[i:end])
	}
	return batches
}

func (e *Engine) evaluateTickConcurrent(tick int64, batches [][]*entity.Entity) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, e.cfg.WorkerCount)

	for bi, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}
		go func(bi int, batch []*entity.Entity) {
			defer wg.Done()
			defer func() { <-sem }()
			r := e.root.Child(int64(bi)*7919 + tick)
			for _, ent := range batch {
				rs, ok := e.behavior.Get(ruleSetFor(ent))
				if !ok {
					continue
				}
				actions := e.behavior.Evaluate(rs.Name, ent, tick, r)
				for _, a := range actions {
					ev := behavior.ToEvent(a, ent.ID, e.world.ID, e.cfg.ID, tick)
					e.bus.Emit(ev)
				}
			}
		}(bi, batch)
	}
	wg.Wait()
}

func ruleSetFor(e *entity.Entity) string { return e.Archetype + "_default" }

func (e *Engine) contextFactory(ev event.Event) event.Context {
	tick := ev.Tick
	r := e.root.Child(tick)
	return event.Context{
		Tick:    tick,
		WorldID: e.world.ID,
		SimID:   e.cfg.ID,
		Get: func(key string) (interface{}, bool) {
			v, ok := e.entIndex.Load(key)
			return v, ok
		},
		Set:  func(key string, value interface{}) { e.entIndex.Store(key, value) },
		Emit: func(ne event.Event) { e.bus.Emit(ne) },
		Rand: r.Float64,
	}
}

func (e *Engine) recordTickMetrics(tick int64, start time.Time, eventCount int) {
	elapsed := time.Since(start)
	e.metrics.Set(fmt.Sprintf("sim.%s.tick", e.cfg.ID), float64(tick))
	e.metrics.Set(fmt.Sprintf("sim.%s.tick_duration_ms", e.cfg.ID), float64(elapsed.Microseconds())/1000.0)
	e.metrics.Inc(fmt.Sprintf("sim.%s.events_processed", e.cfg.ID), float64(eventCount))
	stats := e.bus.Stats()
	e.metrics.Set(fmt.Sprintf("sim.%s.queue_depth", e.cfg.ID), float64(stats.QueueDepth))
	e.metrics.Set(fmt.Sprintf("sim.%s.dead_lettered", e.cfg.ID), float64(stats.DeadLettered))
	e.metrics.Set(fmt.Sprintf("sim.%s.entity_count", e.cfg.ID), float64(len(e.entities)))
	if elapsed > 0 {
		e.metrics.Set(fmt.Sprintf("sim.%s.ticks_per_sec", e.cfg.ID), 1.0/elapsed.Seconds())
	}
}

// EntityCount returns the current live population size.
func (e *Engine) EntityCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.entities)
}

// Entities returns a snapshot slice of entity snapshots (safe copy) up to
// `limit` (0 = all) for API/UI consumption.
func (e *Engine) Entities(limit int) []entity.Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	n := len(e.entities)
	if limit > 0 && limit < n {
		n = limit
	}
	out := make([]entity.Entity, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, e.entities[i].Snapshot())
	}
	return out
}
