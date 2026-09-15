package simulation

import (
	"sync/atomic"
	"time"
)

// WorkerHealth enumerates health states the Coordinator tracks.
type WorkerHealth string

const (
	HealthHealthy   WorkerHealth = "healthy"
	HealthDegraded  WorkerHealth = "degraded"
	HealthUnhealthy WorkerHealth = "unhealthy"
	HealthUnknown   WorkerHealth = "unknown"
)

// Worker represents one simulation worker (in-process goroutine pool today;
// the same struct is what a remote worker process would report over the
// gRPC/REST worker-registration API in a multi-machine deployment).
type Worker struct {
	ID           string       `json:"id"`
	Address      string       `json:"address"` // "local" or host:port
	Health       WorkerHealth `json:"health"`
	AssignedIDs  []string     `json:"assigned_partition"`
	LastHeartbeat time.Time   `json:"last_heartbeat"`

	processed int64
	failures  int64
}

func NewWorker(id, address string) *Worker {
	return &Worker{ID: id, Address: address, Health: HealthUnknown, LastHeartbeat: time.Now()}
}

func (w *Worker) Heartbeat() {
	w.LastHeartbeat = time.Now()
	w.Health = HealthHealthy
}

func (w *Worker) RecordProcessed(n int64) { atomic.AddInt64(&w.processed, n) }
func (w *Worker) RecordFailure()          { atomic.AddInt64(&w.failures, 1) }

func (w *Worker) Processed() int64 { return atomic.LoadInt64(&w.processed) }
func (w *Worker) Failures() int64  { return atomic.LoadInt64(&w.failures) }

// IsStale reports whether the worker missed its heartbeat window, used by
// the Coordinator's Health Monitoring to trigger recovery/reassignment.
func (w *Worker) IsStale(window time.Duration) bool {
	return time.Since(w.LastHeartbeat) > window
}
