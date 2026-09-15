package api

import "net/http"

func (a *App) handleListWorkers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a.mu.RLock()
	coord, ok := a.Coordinators[id]
	a.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "simulation not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"workers":    coord.Workers(),
		"partitions": coord.Partitions(),
	})
}
