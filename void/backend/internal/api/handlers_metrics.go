package api

import "net/http"

func (a *App) handleMetricsJSON(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.Metrics.Snapshot())
}

func (a *App) handleMetricsPrometheus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(a.Metrics.PrometheusText()))
}
