// Package rng provides a deterministic, seedable pseudo-random source used
// everywhere in VOID so that a World generated (or replayed) with the same
// seed always produces bit-identical results.
package rng

import (
	"math"
	"math/rand"
	"sync"
)

// Source is a concurrency-safe, seedable RNG. Every Entity / Worker gets its
// own child Source derived deterministically from the World seed plus a
// "stream id", so parallel workers never contend on a single generator and
// results stay deterministic regardless of scheduling order.
type Source struct {
	mu  sync.Mutex
	r   *rand.Rand
	Seed int64
}

// New creates a root RNG for the given seed.
func New(seed int64) *Source {
	return &Source{r: rand.New(rand.NewSource(seed)), Seed: seed}
}

// Child derives a new, independent-but-deterministic RNG stream from this
// one. streamID should be stable (e.g. entity id hash, worker index).
var goldenRatio64 = func() int64 {
	var u uint64 = 0x9E3779B97F4A7C15
	return int64(u)
}()

func (s *Source) Child(streamID int64) *Source {
	s.mu.Lock()
	mixed := s.Seed ^ (streamID*2654435761 + goldenRatio64)
	s.mu.Unlock()
	return New(mixed)
}

func (s *Source) Float64() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.r.Float64()
}

func (s *Source) Intn(n int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.r.Intn(n)
}

func (s *Source) Int63() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.r.Int63()
}

// Bool returns true with probability p (0..1).
func (s *Source) Bool(p float64) bool {
	return s.Float64() < p
}

// Normal returns a sample from N(mean, stddev).
func (s *Source) Normal(mean, stddev float64) float64 {
	s.mu.Lock()
	v := s.r.NormFloat64()
	s.mu.Unlock()
	return mean + v*stddev
}

// Uniform returns a sample from U(min, max).
func (s *Source) Uniform(min, max float64) float64 {
	return min + s.Float64()*(max-min)
}

// Poisson samples from a Poisson distribution with the given lambda using
// Knuth's algorithm (fine for the lambda ranges used in population/event
// generation; for very large lambda callers should use a normal approx).
func (s *Source) Poisson(lambda float64) int {
	if lambda <= 0 {
		return 0
	}
	L := math.Exp(-lambda)
	k := 0
	p := 1.0
	for {
		k++
		p *= s.Float64()
		if p <= L {
			return k - 1
		}
	}
}

// Zipf samples a rank in [1, n] following a Zipf/power-law distribution with
// exponent s (s > 1 for a "long tail" shape, typical values 1.5-2.5).
func (s *Source) Zipf(exp float64, n uint64) uint64 {
	s.mu.Lock()
	z := rand.NewZipf(s.r, exp, 1, n-1)
	s.mu.Unlock()
	if z == nil {
		return 1
	}
	return z.Uint64() + 1
}

// Choice picks a uniformly random element index from [0, n).
func (s *Source) Choice(n int) int {
	if n <= 0 {
		return 0
	}
	return s.Intn(n)
}

// WeightedChoice picks an index according to relative weights.
func (s *Source) WeightedChoice(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return s.Choice(len(weights))
	}
	r := s.Float64() * total
	acc := 0.0
	for i, w := range weights {
		acc += w
		if r <= acc {
			return i
		}
	}
	return len(weights) - 1
}
