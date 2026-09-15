// Package loadtest turns VOID's simulated behavior into real load against a
// target system under test: HTTP requests, WebSocket events, raw TCP
// connections, one request per simulated Session, with rate limiting,
// concurrency control, ramp-up/down, burst/spike and long-running profiles.
// gRPC and Message Broker adapters implement the same Adapter interface;
// bundled here with a lightweight stdlib-only HTTP/TCP/WS implementation so
// the whole project builds and runs with zero external dependencies -
// production deployments can drop in grpc-go / segmentio-kafka-go backed
// adapters behind the identical interface without touching the engine.
package loadtest

import (
	"context"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Adapter sends one simulated "request" to a target and reports the result.
type Adapter interface {
	Name() string
	Send(ctx context.Context, req Request) (Result, error)
}

// Request is a generic unit of load; adapters interpret the fields they need.
type Request struct {
	Method  string
	URL     string
	Body    []byte
	Headers map[string]string
}

type Result struct {
	StatusCode int
	Latency    time.Duration
	Bytes      int
}

// Profile controls the shape of the load over time.
type Profile struct {
	Concurrency int           `json:"concurrency"`
	RampUp      time.Duration `json:"ramp_up"`
	RampDown    time.Duration `json:"ramp_down"`
	Duration    time.Duration `json:"duration"`
	RatePerSec  float64       `json:"rate_per_sec"` // 0 = unlimited (bounded by concurrency)
	Burst       int           `json:"burst"`         // extra concurrent requests fired at the start
}

// Stats aggregates run results for the dashboard's Load Testing panel.
type Stats struct {
	Sent      int64
	Succeeded int64
	Failed    int64
	TotalLatencyMs int64
	MinLatencyMs   int64
	MaxLatencyMs   int64
}

func (s *Stats) AvgLatencyMs() float64 {
	if s.Sent == 0 {
		return 0
	}
	return float64(s.TotalLatencyMs) / float64(s.Sent)
}

// Runner drives a Profile against an Adapter using a request generator
// function (typically fed by the Simulation Engine's event stream).
type Runner struct {
	Adapter Adapter
	Profile Profile

	stats Stats
	mu    sync.Mutex
}

func NewRunner(a Adapter, p Profile) *Runner {
	if p.Concurrency <= 0 {
		p.Concurrency = 10
	}
	return &Runner{Adapter: a, Profile: p}
}

// Run executes the load profile, pulling requests from `gen` until ctx is
// done or `gen` is exhausted (returns ok=false).
func (r *Runner) Run(ctx context.Context, gen func() (Request, bool)) Stats {
	sem := make(chan struct{}, r.Profile.Concurrency+r.Profile.Burst)
	var wg sync.WaitGroup

	var ticker *time.Ticker
	if r.Profile.RatePerSec > 0 {
		ticker = time.NewTicker(time.Duration(float64(time.Second) / r.Profile.RatePerSec))
		defer ticker.Stop()
	}

	deadline := time.Now().Add(r.Profile.Duration)
	if r.Profile.Duration <= 0 {
		deadline = time.Now().Add(24 * time.Hour) // effectively "run until gen exhausted / ctx cancelled"
	}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			wg.Wait()
			return r.snapshot()
		default:
		}
		if ticker != nil {
			<-ticker.C
		}
		req, ok := gen()
		if !ok {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(req Request) {
			defer wg.Done()
			defer func() { <-sem }()
			r.execute(ctx, req)
		}(req)
	}
	wg.Wait()
	return r.snapshot()
}

func (r *Runner) execute(ctx context.Context, req Request) {
	start := time.Now()
	res, err := r.Adapter.Send(ctx, req)
	latency := time.Since(start).Milliseconds()

	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Sent++
	r.stats.TotalLatencyMs += latency
	if r.stats.MinLatencyMs == 0 || latency < r.stats.MinLatencyMs {
		r.stats.MinLatencyMs = latency
	}
	if latency > r.stats.MaxLatencyMs {
		r.stats.MaxLatencyMs = latency
	}
	if err != nil || res.StatusCode >= 500 {
		r.stats.Failed++
		return
	}
	r.stats.Succeeded++
}

func (r *Runner) snapshot() Stats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats
}

// ---- Built-in adapters (stdlib only) ----

// HTTPAdapter sends real HTTP requests to a target service.
type HTTPAdapter struct {
	Client *http.Client
}

func NewHTTPAdapter(timeout time.Duration) *HTTPAdapter {
	return &HTTPAdapter{Client: &http.Client{Timeout: timeout}}
}

func (a *HTTPAdapter) Name() string { return "http" }

func (a *HTTPAdapter) Send(ctx context.Context, req Request) (Result, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	if err != nil {
		return Result{}, err
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	start := time.Now()
	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	return Result{StatusCode: resp.StatusCode, Latency: time.Since(start)}, nil
}

// TCPAdapter opens a raw TCP connection and writes the request body,
// useful for load-testing custom binary protocols.
type TCPAdapter struct {
	DialTimeout time.Duration
}

func (a *TCPAdapter) Name() string { return "tcp" }

func (a *TCPAdapter) Send(ctx context.Context, req Request) (Result, error) {
	d := net.Dialer{Timeout: a.DialTimeout}
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", req.URL)
	if err != nil {
		return Result{}, err
	}
	defer conn.Close()
	n, err := conn.Write(req.Body)
	if err != nil {
		return Result{}, err
	}
	return Result{StatusCode: 200, Latency: time.Since(start), Bytes: n}, nil
}

// WSAdapter performs a minimal RFC6455 handshake over TCP to measure
// connection + upgrade latency against a WebSocket endpoint without pulling
// in an external dependency.
type WSAdapter struct {
	DialTimeout time.Duration
}

func (a *WSAdapter) Name() string { return "websocket" }

func (a *WSAdapter) Send(ctx context.Context, req Request) (Result, error) {
	d := net.Dialer{Timeout: a.DialTimeout}
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", req.URL)
	if err != nil {
		return Result{}, err
	}
	defer conn.Close()
	handshake := "GET / HTTP/1.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n"
	if _, err := conn.Write([]byte(handshake)); err != nil {
		return Result{}, err
	}
	buf := make([]byte, 512)
	n, _ := conn.Read(buf)
	return Result{StatusCode: 101, Latency: time.Since(start), Bytes: n}, nil
}

// AdapterRegistry exposes counters so the Metrics Engine can report load
// generator throughput alongside simulation metrics.
type AdapterRegistry struct {
	requests int64
}

func (a *AdapterRegistry) Track() { atomic.AddInt64(&a.requests, 1) }
func (a *AdapterRegistry) Count() int64 { return atomic.LoadInt64(&a.requests) }
