// Package population implements VOID's Population Generator: it can
// materialize millions of entities whose attributes follow configurable
// statistical distributions (Normal, Uniform, Gaussian Mixture, Power Law,
// Zipf, Poisson, Custom) with optional cross-attribute correlation.
package population

import (
	"math"

	"void/internal/rng"
)

// DistributionKind enumerates supported distribution shapes.
type DistributionKind string

const (
	Normal          DistributionKind = "normal"
	Uniform         DistributionKind = "uniform"
	GaussianMixture DistributionKind = "gaussian_mixture"
	PowerLaw        DistributionKind = "power_law"
	ZipfDist        DistributionKind = "zipf"
	PoissonDist     DistributionKind = "poisson"
	Custom          DistributionKind = "custom"
)

// MixtureComponent is one weighted Gaussian in a mixture distribution.
type MixtureComponent struct {
	Weight float64 `json:"weight"`
	Mean   float64 `json:"mean"`
	Stddev float64 `json:"stddev"`
}

// Distribution fully configures how one attribute's values are sampled.
type Distribution struct {
	Kind       DistributionKind   `json:"kind"`
	Mean       float64            `json:"mean,omitempty"`
	Stddev     float64            `json:"stddev,omitempty"`
	Min        float64            `json:"min,omitempty"`
	Max        float64            `json:"max,omitempty"`
	Exponent   float64            `json:"exponent,omitempty"` // power-law / zipf
	Lambda     float64            `json:"lambda,omitempty"`   // poisson
	Mixture    []MixtureComponent `json:"mixture,omitempty"`
	CustomBins []CustomBin        `json:"custom_bins,omitempty"`

	// NoiseStddev adds Gaussian jitter (data realism controls: Noise).
	NoiseStddev float64 `json:"noise_stddev,omitempty"`
	// MissingRate: probability the value is omitted (Missing Data control).
	MissingRate float64 `json:"missing_rate,omitempty"`
	// OutlierRate / OutlierScale: probability + magnitude of injected outliers.
	OutlierRate  float64 `json:"outlier_rate,omitempty"`
	OutlierScale float64 `json:"outlier_scale,omitempty"`
}

// CustomBin lets callers supply an arbitrary discrete/custom distribution
// as (value, weight) pairs.
type CustomBin struct {
	Value  float64 `json:"value"`
	Weight float64 `json:"weight"`
}

// Sample draws one value from the distribution using r. ok=false means the
// value was intentionally omitted (missing-data simulation).
func (d Distribution) Sample(r *rng.Source) (value float64, ok bool) {
	if d.MissingRate > 0 && r.Bool(d.MissingRate) {
		return 0, false
	}

	var v float64
	switch d.Kind {
	case Normal:
		v = r.Normal(d.Mean, d.Stddev)
	case Uniform:
		lo, hi := d.Min, d.Max
		if hi <= lo {
			hi = lo + 1
		}
		v = r.Uniform(lo, hi)
	case GaussianMixture:
		v = sampleMixture(r, d.Mixture)
	case PowerLaw:
		v = samplePowerLaw(r, d.Exponent, d.Min, d.Max)
	case ZipfDist:
		n := uint64(d.Max)
		if n < 2 {
			n = 1000
		}
		v = float64(r.Zipf(maxf(d.Exponent, 1.1), n))
	case PoissonDist:
		v = float64(r.Poisson(maxf(d.Lambda, 0.01)))
	case Custom:
		v = sampleCustom(r, d.CustomBins)
	default:
		v = r.Uniform(0, 1)
	}

	if d.NoiseStddev > 0 {
		v += r.Normal(0, d.NoiseStddev)
	}
	if d.OutlierRate > 0 && r.Bool(d.OutlierRate) {
		scale := d.OutlierScale
		if scale == 0 {
			scale = 5
		}
		if r.Bool(0.5) {
			v *= scale
		} else {
			v /= scale
		}
	}
	if d.Max > d.Min && d.Kind != Uniform {
		if v < d.Min {
			v = d.Min
		}
		if v > d.Max {
			v = d.Max
		}
	}
	return v, true
}

func sampleMixture(r *rng.Source, comps []MixtureComponent) float64 {
	if len(comps) == 0 {
		return r.Normal(0, 1)
	}
	weights := make([]float64, len(comps))
	for i, c := range comps {
		weights[i] = c.Weight
	}
	idx := r.WeightedChoice(weights)
	c := comps[idx]
	return r.Normal(c.Mean, c.Stddev)
}

func samplePowerLaw(r *rng.Source, exponent, min, max float64) float64 {
	if exponent <= 0 {
		exponent = 2.5
	}
	if min <= 0 {
		min = 1
	}
	if max <= min {
		max = min * 1000
	}
	u := r.Float64()
	// inverse-CDF sampling for a bounded power law
	g := 1 - exponent
	minG, maxG := pow(min, g), pow(max, g)
	return pow(u*(maxG-minG)+minG, 1/g)
}

func sampleCustom(r *rng.Source, bins []CustomBin) float64 {
	if len(bins) == 0 {
		return 0
	}
	weights := make([]float64, len(bins))
	for i, b := range bins {
		weights[i] = b.Weight
	}
	idx := r.WeightedChoice(weights)
	return bins[idx].Value
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func pow(base, exp float64) float64 {
	if base <= 0 {
		return 0
	}
	return math.Pow(base, exp)
}
