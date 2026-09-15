// Package snapshot implements VOID's State Management System: point-in-time
// captures of a running simulation's world + entity state that can be
// saved, restored, cloned, forked, diffed and rolled back to, each carrying
// a version and metadata for the Replay Engine and Experiment Manager.
package snapshot

import (
	"encoding/json"
	"fmt"
	"time"

	"void/internal/entity"
)

// Snapshot is a fully self-contained, replayable capture of simulation state.
type Snapshot struct {
	ID           string           `json:"id"`
	WorldID      string           `json:"world_id"`
	SimulationID string           `json:"simulation_id"`
	Tick         int64            `json:"tick"`
	Seed         int64            `json:"seed"`
	Version      int              `json:"version"`
	ParentID     string           `json:"parent_id,omitempty"` // set when forked from another snapshot
	CreatedAt    time.Time        `json:"created_at"`
	Entities     []entity.Entity  `json:"entities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// New captures a Snapshot from a live entity slice.
func New(id, worldID, simID string, tick, seed int64, entities []entity.Entity, meta map[string]interface{}) *Snapshot {
	return &Snapshot{
		ID: id, WorldID: worldID, SimulationID: simID, Tick: tick, Seed: seed,
		Version: 1, CreatedAt: time.Now(), Entities: entities, Metadata: meta,
	}
}

// Fork creates a new Snapshot derived from this one (e.g. to branch a
// simulation before applying a different scenario), bumping the version and
// recording lineage via ParentID.
func (s *Snapshot) Fork(newID string) *Snapshot {
	cp := *s
	cp.ID = newID
	cp.ParentID = s.ID
	cp.Version = s.Version + 1
	cp.CreatedAt = time.Now()
	cp.Entities = append([]entity.Entity{}, s.Entities...)
	return &cp
}

// Diff compares two snapshots' entity population and returns a compact
// change summary (added/removed ids + attribute-level deltas for shared
// entities), used by the Replay Engine to show "what changed" between runs.
type Diff struct {
	Added   []string                        `json:"added"`
	Removed []string                        `json:"removed"`
	Changed map[string]map[string][2]float64 `json:"changed"` // id -> attr -> [old,new]
}

func Compare(a, b *Snapshot) Diff {
	aIdx := indexByID(a.Entities)
	bIdx := indexByID(b.Entities)

	d := Diff{Changed: map[string]map[string][2]float64{}}
	for id := range bIdx {
		if _, ok := aIdx[id]; !ok {
			d.Added = append(d.Added, id)
		}
	}
	for id := range aIdx {
		if _, ok := bIdx[id]; !ok {
			d.Removed = append(d.Removed, id)
		}
	}
	for id, ae := range aIdx {
		be, ok := bIdx[id]
		if !ok {
			continue
		}
		changes := map[string][2]float64{}
		for k, av := range ae.Attrs {
			bv := be.Attrs[k]
			if av != bv {
				changes[k] = [2]float64{av, bv}
			}
		}
		if len(changes) > 0 {
			d.Changed[id] = changes
		}
	}
	return d
}

func indexByID(entities []entity.Entity) map[string]entity.Entity {
	m := make(map[string]entity.Entity, len(entities))
	for _, e := range entities {
		m[e.ID] = e
	}
	return m
}

// Marshal/Unmarshal support persisting snapshots to the ObjectStore.
func (s *Snapshot) Marshal() ([]byte, error) { return json.Marshal(s) }

func Unmarshal(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &s, nil
}
