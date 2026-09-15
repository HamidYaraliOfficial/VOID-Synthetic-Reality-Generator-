package api

import (
	"net/http"
	"strconv"
)

func (a *App) handleListEntities(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	writeJSON(w, http.StatusOK, eng.Entities(limit))
}

func (a *App) handleGetEntity(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	entityID := r.PathValue("entityId")
	for _, e := range eng.Entities(0) {
		if e.ID == entityID {
			writeJSON(w, http.StatusOK, e)
			return
		}
	}
	writeError(w, http.StatusNotFound, "entity not found")
}
