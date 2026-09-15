// Package scenario implements the Scenario Builder: a scenario is a
// declarative set of Injections (parameter overrides + synthetic events)
// applied to a running or not-yet-started simulation, either from the
// start, mid-run, or replayed from a Snapshot.
package scenario

import "void/internal/event"

// Injection is one perturbation a Scenario applies: it can override world
// parameters and/or schedule synthetic events at a relative tick offset.
type Injection struct {
	AtTickOffset int64                  `json:"at_tick_offset"`
	Description  string                 `json:"description"`
	ParamOverrides map[string]float64   `json:"param_overrides,omitempty"`
	Events       []event.Event          `json:"events,omitempty"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
}

// Scenario is a named, reusable, shareable collection of Injections.
type Scenario struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Category    string      `json:"category"` // "growth" | "chaos" | "fraud" | "market" | "infra"
	Description string      `json:"description"`
	Injections  []Injection `json:"injections"`
}

// Registry holds all known scenarios (built-in templates + user-defined).
type Registry struct {
	items map[string]*Scenario
}

func NewRegistry() *Registry {
	r := &Registry{items: map[string]*Scenario{}}
	for _, s := range Templates() {
		r.Register(s)
	}
	return r
}

func (r *Registry) Register(s *Scenario) { r.items[s.ID] = s }
func (r *Registry) Get(id string) (*Scenario, bool) {
	s, ok := r.items[id]
	return s, ok
}
func (r *Registry) List() []*Scenario {
	out := make([]*Scenario, 0, len(r.items))
	for _, s := range r.items {
		out = append(out, s)
	}
	return out
}
