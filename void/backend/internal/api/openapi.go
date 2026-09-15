package api

import "net/http"

// handleOpenAPI serves a hand-maintained-but-accurate OpenAPI 3.0 document
// covering every route registered in NewRouter, so external tools and the
// dashboard's typed API client can be generated/validated against it.
func (a *App) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	doc := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "VOID — Synthetic Reality Generator API",
			"version":     "1.0.0",
			"description": "Control plane for Worlds, Populations, Simulations, Scenarios, Chaos, Metrics, Scheduler and Export.",
		},
		"paths": map[string]interface{}{
			"/api/v1/worlds":                                            method("POST", "Create a World", "GET", "List Worlds"),
			"/api/v1/worlds/{id}":                                       method("GET", "Get a World"),
			"/api/v1/archetypes":                                        method("GET", "List Entity Archetypes"),
			"/api/v1/worlds/{id}/population":                            method("POST", "Generate a Population into a Simulation"),
			"/api/v1/simulations":                                       method("POST", "Create a Simulation", "GET", "List Simulations"),
			"/api/v1/simulations/{id}":                                  method("GET", "Get Simulation status"),
			"/api/v1/simulations/{id}/start":                            method("POST", "Start a Simulation"),
			"/api/v1/simulations/{id}/pause":                            method("POST", "Pause a Simulation"),
			"/api/v1/simulations/{id}/resume":                           method("POST", "Resume a Simulation"),
			"/api/v1/simulations/{id}/stop":                             method("POST", "Stop a Simulation"),
			"/api/v1/simulations/{id}/step":                             method("POST", "Advance one tick"),
			"/api/v1/simulations/{id}/entities":                         method("GET", "List Entities"),
			"/api/v1/simulations/{id}/entities/{entityId}":              method("GET", "Get one Entity"),
			"/api/v1/events":                                            method("GET", "Search/filter recent Events"),
			"/api/v1/scenarios":                                         method("GET", "List Scenario templates"),
			"/api/v1/simulations/{id}/scenarios/{scenarioId}/inject":    method("POST", "Inject a Scenario"),
			"/api/v1/chaos/faults":                                      method("POST", "Inject a Chaos fault"),
			"/api/v1/chaos/impact":                                      method("GET", "Chaos impact report"),
			"/api/v1/simulations/{id}/workers":                          method("GET", "List Workers + Partitions"),
			"/api/v1/metrics":                                           method("GET", "Metrics snapshot (JSON)"),
			"/metrics":                                                  method("GET", "Metrics (Prometheus text format)"),
			"/api/v1/scheduler/jobs":                                    method("GET", "List scheduled Jobs", "POST", "Enqueue a Job"),
			"/api/v1/scheduler/operating-hours":                         method("GET", "Get Operating Hours", "PUT", "Set Operating Hours"),
			"/api/v1/scheduler/operating-hours/status":                  method("GET", "Live open/closed status + time to next transition"),
			"/api/v1/simulations/{id}/export":                           method("POST", "Export entity dataset (json|csv|ndjson|parquet_json)"),
			"/api/v1/simulations/{id}/snapshot":                         method("POST", "Create a Snapshot"),
			"/ws/simulations/{id}":                                      method("GET", "WebSocket: live tick/event stream"),
			"/healthz":                                                  method("GET", "Liveness probe"),
			"/readyz":                                                   method("GET", "Readiness probe"),
		},
	}
	writeJSON(w, http.StatusOK, doc)
}

func method(pairs ...string) map[string]interface{} {
	out := map[string]interface{}{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i]] = map[string]string{"summary": pairs[i+1]}
	}
	return out
}
