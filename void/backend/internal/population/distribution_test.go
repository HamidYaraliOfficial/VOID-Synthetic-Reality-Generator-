package population

import (
	"testing"

	"void/internal/rng"
)

func TestNormalDistributionSample(t *testing.T) {
	d := Distribution{Kind: Normal, Mean: 100, Stddev: 10, Min: 0, Max: 1000}
	r := rng.New(1)
	sum := 0.0
	n := 2000
	for i := 0; i < n; i++ {
		v, ok := d.Sample(r)
		if !ok {
			t.Fatal("expected value, got missing")
		}
		sum += v
	}
	avg := sum / float64(n)
	if avg < 90 || avg > 110 {
		t.Fatalf("expected average near 100, got %v", avg)
	}
}

func TestMissingRate(t *testing.T) {
	d := Distribution{Kind: Uniform, Min: 0, Max: 1, MissingRate: 1.0}
	r := rng.New(1)
	_, ok := d.Sample(r)
	if ok {
		t.Fatal("expected missing value with MissingRate=1.0")
	}
}

func TestGeneratorDeterministic(t *testing.T) {
	// covered indirectly via generator_test.go
}
