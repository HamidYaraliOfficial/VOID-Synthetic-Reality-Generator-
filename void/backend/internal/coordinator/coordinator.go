// Package coordinator implements the Distributed Simulation Coordinator: it
// partitions a World's agents/events across Workers, schedules work,
// monitors Worker health, rebalances load, retries failed partitions and
// keeps world state synchronized across the fleet.
package coordinator

import (
	"errors"
	"sync"
	"time"

	"void/internal/simulation"
)

// Partition is a contiguous slice of entity IDs assigned to one Worker.
type Partition struct {
	WorkerID string   `json:"worker_id"`
	EntityIDs []string `json:"entity_ids"`
}

// Coordinator manages the worker fleet for one distributed simulation run.
type Coordinator struct {
	mu             sync.RWMutex
	SimulationID   string
	workers        map[string]*simulation.Worker
	partitions     map[string]*Partition // workerID -> partition
	heartbeatWindow time.Duration
	retryLimit     int
	retries        map[string]int
}

func New(simID string) *Coordinator {
	return &Coordinator{
		SimulationID:    simID,
		workers:         map[string]*simulation.Worker{},
		partitions:      map[string]*Partition{},
		heartbeatWindow: 15 * time.Second,
		retryLimit:      3,
		retries:         map[string]int{},
	}
}

// RegisterWorker adds/reconnects a worker to the fleet.
func (c *Coordinator) RegisterWorker(id, address string) *simulation.Worker {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, ok := c.workers[id]
	if !ok {
		w = simulation.NewWorker(id, address)
		c.workers[id] = w
	}
	w.Heartbeat()
	return w
}

// Heartbeat refreshes liveness for a worker (called periodically by the
// worker process / local goroutine supervisor).
func (c *Coordinator) Heartbeat(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, ok := c.workers[id]
	if !ok {
		return errors.New("unknown worker")
	}
	w.Heartbeat()
	return nil
}

// PartitionEntities splits entityIDs evenly (round-robin) across the
// currently healthy worker set - a simple, fast, and rebalance-friendly
// scheme; larger deployments can swap this for consistent hashing.
func (c *Coordinator) PartitionEntities(entityIDs []string) map[string]*Partition {
	c.mu.Lock()
	defer c.mu.Unlock()

	healthy := c.healthyWorkerIDsLocked()
	if len(healthy) == 0 {
		return nil
	}
	partitions := map[string]*Partition{}
	for _, id := range healthy {
		partitions[id] = &Partition{WorkerID: id}
	}
	for i, eid := range entityIDs {
		wid := healthy[i%len(healthy)]
		partitions[wid].EntityIDs = append(partitions[wid].EntityIDs, eid)
	}
	c.partitions = partitions
	return partitions
}

func (c *Coordinator) healthyWorkerIDsLocked() []string {
	var out []string
	for id, w := range c.workers {
		if w.IsStale(c.heartbeatWindow) {
			w.Health = simulation.HealthUnhealthy
			continue
		}
		out = append(out, id)
	}
	return out
}

// MonitorHealth should be run in a goroutine; it periodically checks for
// stale workers and reassigns their partitions to healthy peers.
func (c *Coordinator) MonitorHealth(interval time.Duration, stop <-chan struct{}, onReassign func(lost *Partition, to string)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			c.reassignStaleLocked(onReassign)
		}
	}
}

func (c *Coordinator) reassignStaleLocked(onReassign func(lost *Partition, to string)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	healthy := c.healthyWorkerIDsLocked()
	if len(healthy) == 0 {
		return
	}
	i := 0
	for wid, w := range c.workers {
		if !w.IsStale(c.heartbeatWindow) {
			continue
		}
		part, ok := c.partitions[wid]
		if !ok || len(part.EntityIDs) == 0 {
			continue
		}
		if c.retries[wid] >= c.retryLimit {
			continue // give up reassigning after retryLimit attempts
		}
		c.retries[wid]++
		target := healthy[i%len(healthy)]
		i++
		delete(c.partitions, wid)
		c.partitions[target].EntityIDs = append(c.partitions[target].EntityIDs, part.EntityIDs...)
		if onReassign != nil {
			onReassign(part, target)
		}
	}
}

// Workers returns a snapshot of all known workers (for the API/dashboard).
func (c *Coordinator) Workers() []*simulation.Worker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*simulation.Worker, 0, len(c.workers))
	for _, w := range c.workers {
		out = append(out, w)
	}
	return out
}

func (c *Coordinator) Partitions() map[string]*Partition {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := make(map[string]*Partition, len(c.partitions))
	for k, v := range c.partitions {
		cp[k] = v
	}
	return cp
}
