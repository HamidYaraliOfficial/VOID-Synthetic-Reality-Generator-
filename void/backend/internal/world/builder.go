package world

import (
	"fmt"

	"void/internal/rng"
)

// SizeToAgentCount maps a human size label to a target agent population,
// used by the CLI/API when the user doesn't specify an exact count.
var SizeToAgentCount = map[string]int{
	"small":   1_000,
	"medium":  100_000,
	"large":   1_000_000,
	"massive": 10_000_000,
}

// Builder deterministically generates World structure (regions, economy
// defaults) from a Config + seed, so identical input always yields an
// identical World for exact-replay guarantees.
type Builder struct{}

func NewBuilder() *Builder { return &Builder{} }

// Build materializes a World, auto-generating regions when the caller
// didn't specify any explicitly (common case: "give me a world of size X").
func (b *Builder) Build(id string, cfg Config) (*World, error) {
	if cfg.Seed == 0 {
		return nil, fmt.Errorf("world seed must be non-zero for deterministic generation")
	}
	if len(cfg.Regions) == 0 {
		cfg.Regions = b.generateRegions(cfg)
	}
	if cfg.Economy.BaseCurrency == "" {
		cfg.Economy.BaseCurrency = "USD"
	}
	if cfg.TickUnit == "" {
		cfg.TickUnit = "hour"
	}
	if cfg.TimeScale == 0 {
		cfg.TimeScale = 1
	}
	w := New(id, cfg)
	return w, nil
}

func (b *Builder) generateRegions(cfg Config) []Region {
	r := rng.New(cfg.Seed)
	total := SizeToAgentCount[cfg.Size]
	if total == 0 {
		total = 10_000
	}
	numRegions := 3 + r.Intn(7)
	regions := make([]Region, 0, numRegions)
	remaining := total
	for i := 0; i < numRegions; i++ {
		share := remaining
		if i != numRegions-1 {
			share = int(float64(remaining) * r.Uniform(0.1, 0.5))
		}
		regions = append(regions, Region{
			ID:         fmt.Sprintf("region-%02d", i),
			Name:       fmt.Sprintf("Region %d", i+1),
			Population: share,
			WealthIdx:  r.Uniform(0.2, 0.9),
			Lat:        r.Uniform(-60, 60),
			Lng:        r.Uniform(-180, 180),
		})
		remaining -= share
		if remaining <= 0 {
			break
		}
	}
	return regions
}
