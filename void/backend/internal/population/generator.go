package population

import (
	"fmt"

	"void/internal/entity"
	"void/internal/rng"
)

// AttributeSpec configures one generated attribute, plus optional
// correlation with a previously generated attribute (e.g. "income"
// correlates positively with "loyalty").
type AttributeSpec struct {
	Name         string           `json:"name"`
	Distribution Distribution     `json:"distribution"`
	CorrelatesWith string         `json:"correlates_with,omitempty"`
	Correlation    float64        `json:"correlation,omitempty"` // -1..1
}

// Spec configures a population generation batch for one archetype.
type Spec struct {
	Archetype  string          `json:"archetype"`
	Count      int             `json:"count"`
	Attributes []AttributeSpec `json:"attributes"`
	IDPrefix   string          `json:"id_prefix"`
}

// Generator materializes Entities according to a Spec, streaming them
// through a callback so callers (Simulation Engine, Export) never need the
// full population resident in memory at once.
type Generator struct {
	Registry *entity.Registry
}

func NewGenerator(reg *entity.Registry) *Generator {
	return &Generator{Registry: reg}
}

// Generate streams `spec.Count` entities to `emit`. Deterministic for a
// given root seed: same seed + spec always yields the same population.
func (g *Generator) Generate(root *rng.Source, spec Spec, emit func(*entity.Entity)) (int, error) {
	arch, ok := g.Registry.Get(spec.Archetype)
	if !ok {
		return 0, fmt.Errorf("unknown archetype %q", spec.Archetype)
	}
	prefix := spec.IDPrefix
	if prefix == "" {
		prefix = spec.Archetype
	}

	generated := 0
	for i := 0; i < spec.Count; i++ {
		id := fmt.Sprintf("%s-%08d", prefix, i)
		r := root.Child(int64(i)*1000003 + hashStr(spec.Archetype))
		e := entity.New(id, arch, 0)

		sampled := map[string]float64{}
		for _, attrSpec := range spec.Attributes {
			dist := attrSpec.Distribution
			if attrSpec.CorrelatesWith != "" && attrSpec.Correlation != 0 {
				if base, ok := sampled[attrSpec.CorrelatesWith]; ok {
					dist = applyCorrelation(dist, base, attrSpec.Correlation)
				}
			}
			v, ok := dist.Sample(r)
			if !ok {
				continue // missing value: entity simply lacks this attribute
			}
			sampled[attrSpec.Name] = v
			e.SetAttr(attrSpec.Name, v)
		}
		emit(e)
		generated++
	}
	return generated, nil
}

// applyCorrelation nudges the mean of `dist` toward/away from a normalized
// version of `baseValue` scaled by `corr` (-1..1). This is a simplified but
// effective way to introduce realistic cross-attribute correlation without
// requiring a full covariance-matrix sampler.
func applyCorrelation(dist Distribution, baseValue, corr float64) Distribution {
	shift := corr * baseValue * 0.05
	dist.Mean += shift
	return dist
}

func hashStr(s string) int64 {
	var h int64 = 1469598103934665603
	for _, c := range s {
		h ^= int64(c)
		h *= 1099511628211
	}
	return h
}
