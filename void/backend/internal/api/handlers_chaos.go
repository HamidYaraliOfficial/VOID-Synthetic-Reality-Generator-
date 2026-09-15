package api

import (
	"fmt"
	"net/http"
	"time"

	"void/internal/chaos"
)

type injectFaultRequest struct {
	Kind          string  `json:"kind"`
	Target        string  `json:"target"`
	Probability   float64 `json:"probability"`
	Magnitude     float64 `json:"magnitude"`
	StartTick     int64   `json:"start_tick"`
	DurationTicks int64   `json:"duration_ticks"`
}

func (a *App) handleInjectFault(w http.ResponseWriter, r *http.Request) {
	var req injectFaultRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Target == "" {
		req.Target = "*"
	}
	f := &chaos.Fault{
		ID: fmt.Sprintf("fault-%d", time.Now().UnixNano()),
		Kind: chaos.FaultKind(req.Kind), Target: req.Target,
		Probability: req.Probability, Magnitude: req.Magnitude,
		StartTick: req.StartTick, DurationTicks: req.DurationTicks,
	}
	a.Chaos.Inject(f)
	writeJSON(w, http.StatusCreated, f)
}

func (a *App) handleChaosImpact(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.Chaos.ImpactReport())
}
