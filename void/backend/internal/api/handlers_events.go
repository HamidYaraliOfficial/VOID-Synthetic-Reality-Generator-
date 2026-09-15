package api

import (
	"net/http"
	"strconv"
)

func (a *App) handleListEvents(w http.ResponseWriter, r *http.Request) {
	all := a.ring.All()

	q := r.URL.Query()
	typeFilter := q.Get("type")
	sourceFilter := q.Get("source")
	targetFilter := q.Get("target")
	severityFilter := q.Get("severity")
	scenarioFilter := q.Get("scenario_id")
	limit := 500
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	out := make([]interface{}, 0, limit)
	for i := len(all) - 1; i >= 0 && len(out) < limit; i-- {
		e := all[i]
		if typeFilter != "" && e.Type != typeFilter {
			continue
		}
		if sourceFilter != "" && e.SourceID != sourceFilter {
			continue
		}
		if targetFilter != "" && e.TargetID != targetFilter {
			continue
		}
		if severityFilter != "" && string(e.Severity) != severityFilter {
			continue
		}
		if scenarioFilter != "" && e.ScenarioID != scenarioFilter {
			continue
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}
