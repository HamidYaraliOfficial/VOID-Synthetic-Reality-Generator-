package api

import (
	"fmt"
	"net/http"
	"time"

	"void/internal/scheduler"
)

func (a *App) handleListJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.Scheduler.List())
}

type enqueueJobRequest struct {
	Kind       string                 `json:"kind"`
	Priority   int                    `json:"priority"`
	RunAt      time.Time              `json:"run_at"`
	MaxRetries int                    `json:"max_retries"`
	Payload    map[string]interface{} `json:"payload"`
}

func (a *App) handleEnqueueJob(w http.ResponseWriter, r *http.Request) {
	var req enqueueJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.RunAt.IsZero() {
		req.RunAt = time.Now()
	}
	job := &scheduler.Job{
		ID: fmt.Sprintf("job-%d", time.Now().UnixNano()), Kind: req.Kind,
		Priority: req.Priority, RunAt: req.RunAt, MaxRetries: req.MaxRetries, Payload: req.Payload,
	}
	a.Scheduler.Enqueue(job)
	writeJSON(w, http.StatusCreated, job)
}

// handleGetOperatingHours returns the user-configured weekly schedule
// exactly as entered (timezone + list of open windows) - nothing is
// hardcoded or assumed by VOID.
func (a *App) handleGetOperatingHours(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"timezone": a.OperatingHours.Timezone,
		"windows":  a.OperatingHours.Windows,
	})
}

type setOperatingHoursRequest struct {
	Timezone string              `json:"timezone"`
	Windows  []scheduler.Window  `json:"windows"`
}

// handleSetOperatingHours lets the user fully define (or redefine) when the
// system is considered "open" - which weekdays, and the exact open/close
// clock time for each, in their chosen timezone.
func (a *App) handleSetOperatingHours(w http.ResponseWriter, r *http.Request) {
	var req setOperatingHoursRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	a.OperatingHours.Set(req.Timezone, req.Windows)
	writeJSON(w, http.StatusOK, map[string]interface{}{"timezone": req.Timezone, "windows": req.Windows})
}

// handleOperatingHoursStatus reports, live, whether the schedule is open
// right now and exactly how long until the next open/close transition.
func (a *App) handleOperatingHoursStatus(w http.ResponseWriter, r *http.Request) {
	status := a.OperatingHours.Evaluate(time.Now())
	writeJSON(w, http.StatusOK, status)
}
