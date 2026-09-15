package rng

import "testing"

func TestDeterminism(t *testing.T) {
	a := New(42)
	b := New(42)
	for i := 0; i < 100; i++ {
		if a.Float64() != b.Float64() {
			t.Fatalf("same seed produced different sequences at iteration %d", i)
		}
	}
}

func TestChildStreamsAreIndependentButDeterministic(t *testing.T) {
	root1 := New(7)
	root2 := New(7)
	c1 := root1.Child(5)
	c2 := root2.Child(5)
	for i := 0; i < 50; i++ {
		if c1.Float64() != c2.Float64() {
			t.Fatalf("same parent seed + streamID should reproduce identical child sequence")
		}
	}
}

func TestDistributionsInRange(t *testing.T) {
	r := New(1)
	for i := 0; i < 1000; i++ {
		v := r.Uniform(0, 1)
		if v < 0 || v > 1 {
			t.Fatalf("uniform out of range: %v", v)
		}
	}
}
