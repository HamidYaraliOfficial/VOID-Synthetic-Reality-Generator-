package api

import (
	"fmt"
	"net/http"
	"time"

	"void/internal/entity"
	"void/internal/population"
	"void/internal/rng"
	"void/internal/world"
)

type createWorldRequest struct {
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Seed        int64              `json:"seed"`
	Size        string             `json:"size"`
	Regions     []world.Region     `json:"regions,omitempty"`
	Economy     world.EconomyConfig `json:"economy,omitempty"`
	TimeScale   float64            `json:"time_scale,omitempty"`
	TickUnit    string             `json:"tick_unit,omitempty"`
	Description string             `json:"description,omitempty"`
}

func (a *App) handleCreateWorld(w http.ResponseWriter, r *http.Request) {
	var req createWorldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Seed == 0 {
		req.Seed = time.Now().UnixNano()
	}
	cfg := world.Config{
		Name: req.Name, Type: req.Type, Seed: req.Seed, Size: req.Size,
		Regions: req.Regions, Economy: req.Economy, TimeScale: req.TimeScale,
		TickUnit: req.TickUnit, StartTime: time.Now(), Description: req.Description,
	}
	id := fmt.Sprintf("world-%d", time.Now().UnixNano())
	w2, err := world.NewBuilder().Build(id, cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.mu.Lock()
	a.Worlds[id] = w2
	a.mu.Unlock()

	writeJSON(w, http.StatusCreated, w2)
}

func (a *App) handleListWorlds(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*world.World, 0, len(a.Worlds))
	for _, wd := range a.Worlds {
		out = append(out, wd)
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) handleGetWorld(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a.mu.RLock()
	wd, ok := a.Worlds[id]
	a.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "world not found")
		return
	}
	writeJSON(w, http.StatusOK, wd)
}

func (a *App) handleListArchetypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.EntityRegistry.List())
}

type generatePopulationRequest struct {
	SimulationID string             `json:"simulation_id"`
	Specs        []population.Spec  `json:"specs"`
}

// handleGeneratePopulation streams a generated population directly into a
// target (already-created) simulation engine.
func (a *App) handleGeneratePopulation(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("id")
	a.mu.RLock()
	wd, worldOK := a.Worlds[worldID]
	a.mu.RUnlock()
	if !worldOK {
		writeError(w, http.StatusNotFound, "world not found")
		return
	}

	var req generatePopulationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	a.mu.RLock()
	eng, simOK := a.Simulations[req.SimulationID]
	a.mu.RUnlock()
	if !simOK {
		writeError(w, http.StatusNotFound, "simulation not found; create it first")
		return
	}

	gen := population.NewGenerator(a.EntityRegistry)
	root := rng.New(wd.Config.Seed)
	total := 0
	for _, spec := range req.Specs {
		n, err := gen.Generate(root, spec, func(e *entity.Entity) {
			eng.AddEntity(e)
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		total += n
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"generated": total, "simulation_id": req.SimulationID})
}
