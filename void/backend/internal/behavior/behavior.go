// Package behavior implements VOID's Behavior Engine. Entities are not
// static rows: on every tick they can be evaluated against Rules, a State
// Machine, and a Utility-AI scorer that picks the highest-utility Action
// given the entity's current state, goals and recent memory.
package behavior

import (
	"void/internal/entity"
	"void/internal/event"
	"void/internal/rng"
)

// Action is something a behavior rule set can decide an entity does next.
type Action struct {
	Name    string
	EventType string
	Payload map[string]interface{}
	Utility float64
}

// Rule is a simple condition -> action mapping ("if churn_risk > 0.8 and
// last_purchase > 30 ticks ago then possibly churn").
type Rule struct {
	Name      string
	Condition func(e *entity.Entity, tick int64, r *rng.Source) bool
	Action    func(e *entity.Entity, tick int64, r *rng.Source) *Action
	Weight    float64
}

// StateTransition defines a State Machine edge.
type StateTransition struct {
	From      string
	To        string
	Condition func(e *entity.Entity, tick int64, r *rng.Source) bool
	Action    *Action
}

// RuleSet groups Rules + a StateMachine + a utility scorer under a name
// (e.g. "customer_default"), assignable to an Archetype.
type RuleSet struct {
	Name         string
	Rules        []Rule
	Transitions  []StateTransition
	UtilityScore func(e *entity.Entity, goal entity.Goal, tick int64) float64
}

// Engine holds all registered RuleSets and evaluates entities each tick.
type Engine struct {
	sets map[string]*RuleSet
}

func NewEngine() *Engine {
	eng := &Engine{sets: map[string]*RuleSet{}}
	for _, rs := range builtinRuleSets() {
		eng.Register(rs)
	}
	return eng
}

func (eng *Engine) Register(rs *RuleSet) { eng.sets[rs.Name] = rs }

func (eng *Engine) Get(name string) (*RuleSet, bool) {
	rs, ok := eng.sets[name]
	return rs, ok
}

// Evaluate runs the state machine + rules + utility-AI goal selection for
// one entity on one tick, returning zero or more Actions the caller
// (Simulation Engine) should turn into Events.
func (eng *Engine) Evaluate(rulesetName string, e *entity.Entity, tick int64, r *rng.Source) []Action {
	rs, ok := eng.sets[rulesetName]
	if !ok {
		return nil
	}
	var actions []Action

	// 1. State machine transitions.
	status := e.State["status"]
	for _, t := range rs.Transitions {
		if t.From != "" && t.From != status {
			continue
		}
		if t.Condition != nil && t.Condition(e, tick, r) {
			e.SetState("status", t.To)
			if t.Action != nil {
				actions = append(actions, *t.Action)
			}
		}
	}

	// 2. Rule evaluation (probabilistic weight gate).
	for _, rule := range rs.Rules {
		if rule.Condition == nil || !rule.Condition(e, tick, r) {
			continue
		}
		if rule.Weight > 0 && rule.Weight < 1 && !r.Bool(rule.Weight) {
			continue
		}
		if rule.Action == nil {
			continue
		}
		if a := rule.Action(e, tick, r); a != nil {
			actions = append(actions, *a)
		}
	}

	// 3. Utility-AI: pick the highest scoring goal and nudge progress.
	if rs.UtilityScore != nil && len(e.Goals) > 0 {
		bestIdx, bestScore := -1, -1e18
		for i, g := range e.Goals {
			score := rs.UtilityScore(e, g, tick)
			if score > bestScore {
				bestScore, bestIdx = score, i
			}
		}
		if bestIdx >= 0 {
			e.Goals[bestIdx].Progress += 0.01 * bestScore
		}
	}

	return actions
}

// ToEvent converts an Action into a structured event.Event ready for the bus.
func ToEvent(a Action, sourceID, worldID, simID string, tick int64) event.Event {
	return event.Event{
		Type:      a.EventType,
		Tick:      tick,
		SourceID:  sourceID,
		Payload:   a.Payload,
		Priority:  1,
		Severity:  event.SeverityInfo,
		WorldID:   worldID,
		SimulationID: simID,
	}
}
