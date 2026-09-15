package api

import (
	"net/http"
	"time"

	"void/internal/event"
)

func (a *App) handleListScenarios(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.ScenarioRegistry.List())
}

// handleInjectScenario schedules every Injection in the named Scenario onto
// the target simulation's event bus, offset from the simulation's current tick.
func (a *App) handleInjectScenario(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	scenarioID := r.PathValue("scenarioId")
	sc, ok := a.ScenarioRegistry.Get(scenarioID)
	if !ok {
		writeError(w, http.StatusNotFound, "scenario not found")
		return
	}

	baseTick := eng.Tick()
	scheduled := 0
	for _, inj := range sc.Injections {
		for _, e := range inj.Events {
			e.Tick = baseTick + inj.AtTickOffset
			e.ScenarioID = sc.ID
			if e.Timestamp.IsZero() {
				e.Timestamp = time.Now()
			}
			if e.Severity == "" {
				e.Severity = event.SeverityInfo
			}
			eng.Bus().Emit(e)
			scheduled++
		}
		// Synthetic marker event so param overrides are visible in the Event
		// Explorer even when a scenario carries no explicit Events.
		if len(inj.ParamOverrides) > 0 {
			eng.Bus().Emit(event.Event{
				Type: "scenario_param_override", Tick: baseTick + inj.AtTickOffset,
				Timestamp: time.Now(), ScenarioID: sc.ID, Severity: event.SeverityWarning,
				Payload: map[string]interface{}{"overrides": inj.ParamOverrides, "description": inj.Description},
			})
			scheduled++
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"scenario_id": sc.ID, "events_scheduled": scheduled})
}
