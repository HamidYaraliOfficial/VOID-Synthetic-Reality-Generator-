// Package api implements VOID's REST + WebSocket control plane: it exposes
// every engine (World Builder, Population Generator, Simulation Engine,
// Coordinator, Scenario Builder, Chaos Engine, Metrics, Scheduler,
// Snapshots, Replay, Export) over HTTP for the Next.js Control Center and
// the CLI. An OpenAPI document is served at /openapi.json.
package api

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"void/internal/behavior"
	"void/internal/chaos"
	"void/internal/config"
	"void/internal/coordinator"
	"void/internal/entity"
	"void/internal/event"
	"void/internal/metrics"
	"void/internal/scenario"
	"void/internal/scheduler"
	"void/internal/simulation"
	"void/internal/storage"
	"void/internal/world"
)

// App is VOID's shared runtime: one process (headless or API-serving) wires
// up an App and either drives it via CLI commands or exposes it over HTTP.
type App struct {
	Cfg config.Config

	EntityRegistry   *entity.Registry
	BehaviorEngine   *behavior.Engine
	ScenarioRegistry *scenario.Registry
	Metrics          *metrics.Registry
	Chaos            *chaos.Engine
	Scheduler        *scheduler.Scheduler
	OperatingHours   *scheduler.OperatingHours
	Storage          *storage.Provider

	mu          sync.RWMutex
	Worlds      map[string]*world.World
	Simulations map[string]*simulation.Engine
	Coordinators map[string]*coordinator.Coordinator
	ring        *eventRing // recent events for the Event Explorer
	hub         *wsHub     // websocket subscriber registry
}

func NewApp(cfg config.Config) *App {
	return &App{
		Cfg:              cfg,
		EntityRegistry:   entity.NewRegistry(),
		BehaviorEngine:   behavior.NewEngine(),
		ScenarioRegistry: scenario.NewRegistry(),
		Metrics:          metrics.NewRegistry(),
		Chaos:            chaos.NewEngine(),
		Scheduler:        scheduler.New(),
		OperatingHours:   scheduler.NewOperatingHours("UTC", nil),
		Storage:          storage.NewDefaultProvider(cfg.DataDir),
		Worlds:           map[string]*world.World{},
		Simulations:      map[string]*simulation.Engine{},
		Coordinators:     map[string]*coordinator.Coordinator{},
		ring:             newEventRing(5000),
		hub:              newWSHub(),
	}
}

// eventRing is a small fixed-size ring buffer of recent events, backing the
// Event Explorer without requiring a real event-store query engine.
type eventRing struct {
	mu    sync.Mutex
	items []event.Event
	cap   int
	next  int
	full  bool
}

func newEventRing(capacity int) *eventRing {
	return &eventRing{items: make([]event.Event, capacity), cap: capacity}
}

func (r *eventRing) Add(events []event.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range events {
		r.items[r.next] = e
		r.next = (r.next + 1) % r.cap
		if r.next == 0 {
			r.full = true
		}
	}
}

func (r *eventRing) All() []event.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.full {
		return append([]event.Event{}, r.items[:r.next]...)
	}
	out := make([]event.Event, 0, r.cap)
	out = append(out, r.items[r.next:]...)
	out = append(out, r.items[:r.next]...)
	return out
}

// NewRouter builds the full HTTP mux for the Control Center + CLI + external
// integrations. Uses Go 1.22's stdlib pattern-based routing (method + path
// params) so the whole API ships with zero external router dependency.
func (a *App) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Worlds
	mux.HandleFunc("POST /api/v1/worlds", a.withCORS(a.handleCreateWorld))
	mux.HandleFunc("GET /api/v1/worlds", a.withCORS(a.handleListWorlds))
	mux.HandleFunc("GET /api/v1/worlds/{id}", a.withCORS(a.handleGetWorld))

	// Archetypes
	mux.HandleFunc("GET /api/v1/archetypes", a.withCORS(a.handleListArchetypes))

	// Population
	mux.HandleFunc("POST /api/v1/worlds/{id}/population", a.withCORS(a.handleGeneratePopulation))

	// Simulations
	mux.HandleFunc("POST /api/v1/simulations", a.withCORS(a.handleCreateSimulation))
	mux.HandleFunc("GET /api/v1/simulations", a.withCORS(a.handleListSimulations))
	mux.HandleFunc("GET /api/v1/simulations/{id}", a.withCORS(a.handleGetSimulation))
	mux.HandleFunc("POST /api/v1/simulations/{id}/start", a.withCORS(a.handleStartSimulation))
	mux.HandleFunc("POST /api/v1/simulations/{id}/pause", a.withCORS(a.handlePauseSimulation))
	mux.HandleFunc("POST /api/v1/simulations/{id}/resume", a.withCORS(a.handleResumeSimulation))
	mux.HandleFunc("POST /api/v1/simulations/{id}/stop", a.withCORS(a.handleStopSimulation))
	mux.HandleFunc("POST /api/v1/simulations/{id}/step", a.withCORS(a.handleStepSimulation))

	// Entities (inspector)
	mux.HandleFunc("GET /api/v1/simulations/{id}/entities", a.withCORS(a.handleListEntities))
	mux.HandleFunc("GET /api/v1/simulations/{id}/entities/{entityId}", a.withCORS(a.handleGetEntity))

	// Events (explorer)
	mux.HandleFunc("GET /api/v1/events", a.withCORS(a.handleListEvents))

	// Scenarios
	mux.HandleFunc("GET /api/v1/scenarios", a.withCORS(a.handleListScenarios))
	mux.HandleFunc("POST /api/v1/simulations/{id}/scenarios/{scenarioId}/inject", a.withCORS(a.handleInjectScenario))

	// Chaos
	mux.HandleFunc("POST /api/v1/chaos/faults", a.withCORS(a.handleInjectFault))
	mux.HandleFunc("GET /api/v1/chaos/impact", a.withCORS(a.handleChaosImpact))

	// Workers / Coordinator
	mux.HandleFunc("GET /api/v1/simulations/{id}/workers", a.withCORS(a.handleListWorkers))

	// Metrics
	mux.HandleFunc("GET /api/v1/metrics", a.withCORS(a.handleMetricsJSON))
	mux.HandleFunc("GET /metrics", a.handleMetricsPrometheus)

	// Scheduler + Operating Hours
	mux.HandleFunc("GET /api/v1/scheduler/jobs", a.withCORS(a.handleListJobs))
	mux.HandleFunc("POST /api/v1/scheduler/jobs", a.withCORS(a.handleEnqueueJob))
	mux.HandleFunc("GET /api/v1/scheduler/operating-hours", a.withCORS(a.handleGetOperatingHours))
	mux.HandleFunc("PUT /api/v1/scheduler/operating-hours", a.withCORS(a.handleSetOperatingHours))
	mux.HandleFunc("GET /api/v1/scheduler/operating-hours/status", a.withCORS(a.handleOperatingHoursStatus))

	// Export / Snapshot
	mux.HandleFunc("POST /api/v1/simulations/{id}/export", a.withCORS(a.handleExport))
	mux.HandleFunc("POST /api/v1/simulations/{id}/snapshot", a.withCORS(a.handleSnapshot))

	// Realtime
	mux.HandleFunc("GET /ws/simulations/{id}", a.handleWebSocket)

	// OpenAPI + health
	mux.HandleFunc("GET /openapi.json", a.withCORS(a.handleOpenAPI))
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /readyz", a.handleHealth)

	return mux
}

func (a *App) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.Cfg.CORSAllowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Run starts the HTTP server with graceful shutdown on ctx cancellation.
func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{Addr: a.Cfg.HTTPAddr, Handler: a.NewRouter()}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("VOID API listening on %s", a.Cfg.HTTPAddr)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
