// Package chaos implements VOID's Chaos Simulation Engine: it can inject
// faults (network delay, packet loss, service/database/worker failure,
// resource starvation, random errors, dependency failure) into a running
// simulation and measure the resulting cascading effects via metrics.
package chaos

import (
	"sync"
	"time"

	"void/internal/rng"
)

// FaultKind enumerates supported chaos injections.
type FaultKind string

const (
	FaultNetworkDelay      FaultKind = "network_delay"
	FaultPacketLoss        FaultKind = "packet_loss"
	FaultServiceFailure    FaultKind = "service_failure"
	FaultDatabaseFailure   FaultKind = "database_failure"
	FaultWorkerCrash       FaultKind = "worker_crash"
	FaultResourceStarvation FaultKind = "resource_starvation"
	FaultRandomError       FaultKind = "random_error"
	FaultDependencyFailure FaultKind = "dependency_failure"
)

// Fault configures one active chaos injection.
type Fault struct {
	ID          string        `json:"id"`
	Kind        FaultKind     `json:"kind"`
	Target      string        `json:"target"` // entity id, service name, or "*"
	Probability float64       `json:"probability"`
	Magnitude   float64       `json:"magnitude"` // e.g. ms of delay, % packet loss
	StartTick   int64         `json:"start_tick"`
	DurationTicks int64       `json:"duration_ticks"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Engine tracks active faults and lets callers query "is this affected right now".
type Engine struct {
	mu     sync.RWMutex
	faults map[string]*Fault
	effects map[string]int64 // faultID -> count of times it fired, for impact measurement
}

func NewEngine() *Engine {
	return &Engine{faults: map[string]*Fault{}, effects: map[string]int64{}}
}

func (e *Engine) Inject(f *Fault) {
	e.mu.Lock()
	defer e.mu.Unlock()
	f.CreatedAt = time.Now()
	e.faults[f.ID] = f
	e.effects[f.ID] = 0
}

func (e *Engine) Remove(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.faults, id)
}

func (e *Engine) Active(tick int64) []*Fault {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var out []*Fault
	for _, f := range e.faults {
		if tick < f.StartTick {
			continue
		}
		if f.DurationTicks > 0 && tick > f.StartTick+f.DurationTicks {
			continue
		}
		out = append(out, f)
	}
	return out
}

// Affects evaluates whether `target` is hit by any active fault this tick,
// returning the fault (if any) and whether it fired probabilistically.
func (e *Engine) Affects(tick int64, target string, r *rng.Source) (*Fault, bool) {
	for _, f := range e.Active(tick) {
		if f.Target != "*" && f.Target != target {
			continue
		}
		if r.Bool(f.Probability) {
			e.mu.Lock()
			e.effects[f.ID]++
			e.mu.Unlock()
			return f, true
		}
	}
	return nil, false
}

// ImpactReport summarizes how many times each active/recent fault fired,
// used to measure cascading effects for the dashboard's Chaos panel.
func (e *Engine) ImpactReport() map[string]int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]int64, len(e.effects))
	for k, v := range e.effects {
		out[k] = v
	}
	return out
}
