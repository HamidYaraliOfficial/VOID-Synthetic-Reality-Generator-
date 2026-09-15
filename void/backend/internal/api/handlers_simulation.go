package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"void/internal/coordinator"
	"void/internal/event"
	"void/internal/simulation"
)

type createSimulationRequest struct {
	WorldID     string  `json:"world_id"`
	Seed        int64   `json:"seed"`
	MaxTicks    int64   `json:"max_ticks"`
	WorkerCount int     `json:"worker_count"`
	TimeScale   float64 `json:"time_scale"`
	BatchSize   int     `json:"batch_size"`
	Headless    bool    `json:"headless"`
}

func (a *App) handleCreateSimulation(w http.ResponseWriter, r *http.Request) {
	var req createSimulationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	a.mu.RLock()
	wd, ok := a.Worlds[req.WorldID]
	a.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "world not found")
		return
	}
	seed := req.Seed
	if seed == 0 {
		seed = wd.Config.Seed
	}
	id := fmt.Sprintf("sim-%d", time.Now().UnixNano())
	workerCount := req.WorkerCount
	if workerCount <= 0 {
		workerCount = a.Cfg.DefaultWorkerCount
	}
	cfg := simulation.Config{
		ID: id, WorldID: req.WorldID, Seed: seed, MaxTicks: req.MaxTicks,
		WorkerCount: workerCount, TimeScale: req.TimeScale, BatchSize: req.BatchSize,
		Headless: req.Headless,
	}
	eng := simulation.New(cfg, wd, a.BehaviorEngine, a.Metrics)
	eng.OnTick(func(tick int64, processed []event.Event) {
		a.ring.Add(processed)
		a.broadcastTick(id, tick, processed)
	})

	a.mu.Lock()
	a.Simulations[id] = eng
	a.Coordinators[id] = coordinator.New(id)
	a.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "world_id": req.WorldID, "status": eng.Status()})
}

func (a *App) handleListSimulations(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]map[string]interface{}, 0, len(a.Simulations))
	for id, eng := range a.Simulations {
		out = append(out, map[string]interface{}{
			"id": id, "status": eng.Status(), "tick": eng.Tick(), "entity_count": eng.EntityCount(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) handleGetSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": eng.Status(), "tick": eng.Tick(), "entity_count": eng.EntityCount(),
	})
}

func (a *App) simOr404(w http.ResponseWriter, r *http.Request) (*simulation.Engine, bool) {
	id := r.PathValue("id")
	a.mu.RLock()
	eng, ok := a.Simulations[id]
	a.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "simulation not found")
		return nil, false
	}
	return eng, true
}

func (a *App) handleStartSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	go func() {
		ctx := context.Background()
		_ = eng.Run(ctx)
	}()
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (a *App) handlePauseSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	eng.Pause()
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (a *App) handleResumeSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	eng.Resume()
	writeJSON(w, http.StatusOK, map[string]string{"status": "running"})
}

func (a *App) handleStopSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	eng.Stop()
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (a *App) handleStepSimulation(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	eng.Step()
	writeJSON(w, http.StatusOK, map[string]string{"status": "stepped", "tick": fmt.Sprintf("%d", eng.Tick())})
}
