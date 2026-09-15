// Package event implements VOID's Event Engine: every change in the world
// happens through a structured Event with trigger, conditions, priority,
// timestamp, payload, source/target, probability, dependencies and
// consequences, dispatched over a high-throughput in-process bus and
// optionally streamed to NATS/Kafka via the storage.EventStream interface.
package event

import "time"

// Severity classifies an event for filtering/alerting in the Event Explorer.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event is the atomic unit of change in a VOID world.
type Event struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // e.g. "purchase", "login_failure"
	Tick         int64                  `json:"tick"`
	Timestamp    time.Time              `json:"timestamp"`
	SourceID     string                 `json:"source_id,omitempty"`
	TargetID     string                 `json:"target_id,omitempty"`
	Priority     int                    `json:"priority"` // higher runs first within a tick
	Probability  float64                `json:"probability"`
	DurationTicks int64                 `json:"duration_ticks,omitempty"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"` // event IDs that must precede this one
	Severity     Severity               `json:"severity"`
	ScenarioID   string                 `json:"scenario_id,omitempty"`
	WorldID      string                 `json:"world_id"`
	SimulationID string                 `json:"simulation_id"`
}

// Consequence is a follow-up event a Handler may schedule as a reaction.
type Consequence struct {
	DelayTicks int64
	Event      Event
}

// Condition gates whether an Event should fire, evaluated against arbitrary
// world/entity state supplied by the caller (kept decoupled from concrete
// entity types so plugins can define custom conditions).
type Condition func(ctx Context) bool

// Trigger decides, at a given tick, whether to emit new events.
type Trigger func(ctx Context) []Event

// Handler reacts to an event and may return follow-up Consequences.
type Handler func(ctx Context, e Event) []Consequence

// Context is passed to conditions/triggers/handlers so they can read (and
// queue mutations to) world state without the event package depending on
// the simulation/world packages, avoiding import cycles.
type Context struct {
	Tick     int64
	WorldID  string
	SimID    string
	Get      func(key string) (interface{}, bool)
	Set      func(key string, value interface{})
	Emit     func(e Event)
	Rand     func() float64
}
