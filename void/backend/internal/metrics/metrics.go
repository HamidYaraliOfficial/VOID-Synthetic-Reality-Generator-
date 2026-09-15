// Package metrics implements VOID's Telemetry & Metrics Engine: it collects
// entity counts, event rate, tick rate, CPU/memory, queue depth, worker
// health, throughput, latency and error counters in real time and exposes
// them in a Prometheus-compatible text format for Grafana-compatible
// dashboards, plus a structured JSON snapshot for the web UI.
package metrics

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	mu       sync.Mutex
	counters map[string]float64
	gauges   map[string]float64
	started  time.Time
}

func NewRegistry() *Registry {
	return &Registry{
		counters: map[string]float64{},
		gauges:   map[string]float64{},
		started:  time.Now(),
	}
}

func (r *Registry) Inc(name string, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += v
}

func (r *Registry) Set(name string, v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[name] = v
}

func (r *Registry) Get(name string) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.counters[name]; ok {
		return v
	}
	return r.gauges[name]
}

// Snapshot is a point-in-time JSON-friendly view of all metrics, plus Go
// runtime stats (goroutines, heap) useful for the Simulation Control Center.
type Snapshot struct {
	Counters     map[string]float64 `json:"counters"`
	Gauges       map[string]float64 `json:"gauges"`
	Goroutines   int                `json:"goroutines"`
	HeapAllocMB  float64            `json:"heap_alloc_mb"`
	UptimeSecond float64            `json:"uptime_seconds"`
}

func (r *Registry) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c := make(map[string]float64, len(r.counters))
	for k, v := range r.counters {
		c[k] = v
	}
	g := make(map[string]float64, len(r.gauges))
	for k, v := range r.gauges {
		g[k] = v
	}
	return Snapshot{
		Counters:     c,
		Gauges:       g,
		Goroutines:   runtime.NumGoroutine(),
		HeapAllocMB:  float64(ms.HeapAlloc) / (1024 * 1024),
		UptimeSecond: time.Since(r.started).Seconds(),
	}
}

// PrometheusText renders all metrics in Prometheus exposition format so
// VOID can be scraped directly, or fed into a Grafana-compatible pipeline.
func (r *Registry) PrometheusText() string {
	snap := r.Snapshot()
	var sb strings.Builder
	writeSorted := func(prefix string, m map[string]float64) {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&sb, "void_%s{} %v\n", sanitize(k), m[k])
			_ = prefix
		}
	}
	writeSorted("counter", snap.Counters)
	writeSorted("gauge", snap.Gauges)
	fmt.Fprintf(&sb, "void_goroutines{} %d\n", snap.Goroutines)
	fmt.Fprintf(&sb, "void_heap_alloc_mb{} %v\n", snap.HeapAllocMB)
	fmt.Fprintf(&sb, "void_uptime_seconds{} %v\n", snap.UptimeSecond)
	return sb.String()
}

func sanitize(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}
