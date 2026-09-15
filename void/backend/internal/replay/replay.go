// Package replay implements VOID's Replay Engine: given a Seed,
// Configuration and (optionally) a starting Snapshot, it deterministically
// re-runs a simulation and reports how the new run's final state differs
// from a previously recorded run, enabling exact-replay verification and
// "what changed since last time" analysis.
package replay

import (
	"context"

	"void/internal/behavior"
	"void/internal/entity"
	"void/internal/metrics"
	"void/internal/simulation"
	"void/internal/snapshot"
	"void/internal/world"
)

// Request describes one replay: reuse the same seed/config, optionally
// restoring entity state from a starting Snapshot instead of generating a
// fresh population.
type Request struct {
	World         *world.World
	SimConfig     simulation.Config
	StartSnapshot *snapshot.Snapshot // nil = fresh population must be added by caller before Run
	Behavior      *behavior.Engine
	Metrics       *metrics.Registry
}

// Result carries the replayed engine plus, if a comparison snapshot was
// supplied, a Diff against it.
type Result struct {
	Engine *simulation.Engine
	Diff   *snapshot.Diff
}

// Run executes the replay to completion (respecting SimConfig.MaxTicks) and
// optionally diffs the final population against `compareAgainst`.
func Run(ctx context.Context, req Request, compareAgainst *snapshot.Snapshot) (*Result, error) {
	eng := simulation.New(req.SimConfig, req.World, req.Behavior, req.Metrics)

	if req.StartSnapshot != nil {
		reg := entity.NewRegistry()
		for _, es := range req.StartSnapshot.Entities {
			arch, ok := reg.Get(es.Archetype)
			if !ok {
				continue
			}
			ent := entity.New(es.ID, arch, es.CreatedAt)
			for k, v := range es.Attrs {
				ent.SetAttr(k, v)
			}
			eng.AddEntity(ent)
		}
	}

	if err := eng.Run(ctx); err != nil && err != context.Canceled {
		return nil, err
	}

	res := &Result{Engine: eng}
	if compareAgainst != nil {
		final := snapshot.New("replay-final", req.World.ID, req.SimConfig.ID, eng.Tick(), req.SimConfig.Seed, eng.Entities(0), nil)
		d := snapshot.Compare(compareAgainst, final)
		res.Diff = &d
	}
	return res, nil
}
