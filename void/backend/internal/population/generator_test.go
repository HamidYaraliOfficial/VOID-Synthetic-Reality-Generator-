package population

import (
	"testing"

	"void/internal/entity"
	"void/internal/rng"
)

func TestGenerateIsDeterministicForSameSeed(t *testing.T) {
	reg := entity.NewRegistry()
	spec := Spec{
		Archetype: "customer", Count: 500,
		Attributes: []AttributeSpec{
			{Name: "income", Distribution: Distribution{Kind: Normal, Mean: 4000, Stddev: 500}},
		},
	}

	gen := NewGenerator(reg)
	var a, b []float64
	_, err := gen.Generate(rng.New(99), spec, func(e *entity.Entity) { a = append(a, e.Attr("income")) })
	if err != nil {
		t.Fatal(err)
	}
	_, err = gen.Generate(rng.New(99), spec, func(e *entity.Entity) { b = append(b, e.Attr("income")) })
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b) {
		t.Fatalf("length mismatch: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("mismatch at %d: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestUnknownArchetypeErrors(t *testing.T) {
	reg := entity.NewRegistry()
	gen := NewGenerator(reg)
	_, err := gen.Generate(rng.New(1), Spec{Archetype: "does-not-exist", Count: 1}, func(e *entity.Entity) {})
	if err == nil {
		t.Fatal("expected error for unknown archetype")
	}
}
