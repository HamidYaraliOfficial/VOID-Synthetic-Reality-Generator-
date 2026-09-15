// Package entity implements VOID's generic Entity Model: every simulated
// object (User, Customer, Employee, Company, Product, Account, Device,
// Vehicle, Location, ...) is represented as one flexible Entity built from
// an Archetype definition, so new entity types never require engine changes.
package entity

import (
	"sync"
	"time"
)

// Archetype describes a class of entity (e.g. "customer", "company",
// "vehicle") - its default attributes, goals and behavior rule set.
type Archetype struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	DefaultAttrs     map[string]float64     `json:"default_attrs"`
	DefaultTags      []string               `json:"default_tags"`
	DefaultGoals     []Goal                 `json:"default_goals"`
	BehaviorRuleSet  string                 `json:"behavior_rule_set"`
	ResourceTemplate map[string]float64     `json:"resource_template"`
	Meta             map[string]interface{} `json:"meta,omitempty"`
}

// Goal represents something an Entity is trying to achieve; the Behavior
// Engine consults goals (utility AI) when deciding what an entity does next.
type Goal struct {
	Name     string  `json:"name"`
	Priority float64 `json:"priority"`
	Target   float64 `json:"target"`
	Progress float64 `json:"progress"`
}

// Relationship links two entities (friend, employer, customer-of, follows,
// depends-on, ...) with a strength/weight that behaviors can use.
type Relationship struct {
	TargetID string  `json:"target_id"`
	Type     string  `json:"type"`
	Strength float64 `json:"strength"`
	Since    int64   `json:"since_tick"`
}

// MemoryEntry is a compact record of something that happened to the entity,
// used by goal-oriented / historical-context behaviors.
type MemoryEntry struct {
	Tick    int64                  `json:"tick"`
	Kind    string                 `json:"kind"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// Entity is the concrete, mutable simulated object.
type Entity struct {
	mu *sync.RWMutex

	ID        string   `json:"id"`
	Archetype string   `json:"archetype"`
	Tags      []string `json:"tags"`
	CreatedAt int64    `json:"created_tick"`

	Attrs     map[string]float64 `json:"attrs"`     // e.g. income, age, trust_score
	State     map[string]string  `json:"state"`      // e.g. status=active, churn_risk=low
	Resources map[string]float64 `json:"resources"`  // e.g. balance, inventory, energy
	Skills    map[string]float64 `json:"skills"`
	Prefs     map[string]float64 `json:"preferences"`
	RiskProfile float64          `json:"risk_profile"`

	Goals         []Goal          `json:"goals"`
	Relationships []Relationship  `json:"relationships"`
	Memory        []MemoryEntry   `json:"memory"`

	LastActiveTick int64 `json:"last_active_tick"`
}

const maxMemory = 200

// New creates an Entity from an Archetype.
func New(id string, a *Archetype, tick int64) *Entity {
	e := &Entity{
		ID:        id,
		Archetype: a.Name,
		Tags:      append([]string{}, a.DefaultTags...),
		CreatedAt: tick,
		Attrs:     map[string]float64{},
		State:     map[string]string{"status": "active"},
		Resources: map[string]float64{},
		Skills:    map[string]float64{},
		Prefs:     map[string]float64{},
		Goals:     append([]Goal{}, a.DefaultGoals...),
		mu:        &sync.RWMutex{},
	}
	for k, v := range a.DefaultAttrs {
		e.Attrs[k] = v
	}
	for k, v := range a.ResourceTemplate {
		e.Resources[k] = v
	}
	return e
}

func (e *Entity) SetAttr(k string, v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Attrs[k] = v
}

func (e *Entity) Attr(k string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Attrs[k]
}

func (e *Entity) SetState(k, v string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.State[k] = v
}

func (e *Entity) Remember(tick int64, kind string, payload map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Memory = append(e.Memory, MemoryEntry{Tick: tick, Kind: kind, Payload: payload})
	if len(e.Memory) > maxMemory {
		e.Memory = e.Memory[len(e.Memory)-maxMemory:]
	}
	e.LastActiveTick = tick
}

func (e *Entity) AddRelationship(r Relationship) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Relationships = append(e.Relationships, r)
}

// Snapshot returns a deep-enough copy safe to serialize/export concurrently.
func (e *Entity) Snapshot() Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	cp := *e
	cp.Attrs = cloneF(e.Attrs)
	cp.Resources = cloneF(e.Resources)
	cp.Skills = cloneF(e.Skills)
	cp.Prefs = cloneF(e.Prefs)
	cp.State = cloneS(e.State)
	cp.Goals = append([]Goal{}, e.Goals...)
	cp.Relationships = append([]Relationship{}, e.Relationships...)
	cp.Memory = append([]MemoryEntry{}, e.Memory...)
	cp.Tags = append([]string{}, e.Tags...)
	return cp
}

func cloneF(m map[string]float64) map[string]float64 {
	c := make(map[string]float64, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}
func cloneS(m map[string]string) map[string]string {
	c := make(map[string]string, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// Now is a small helper kept here to avoid importing time in callers just
// for wall-clock metadata on generated payloads.
func Now() int64 { return time.Now().Unix() }
