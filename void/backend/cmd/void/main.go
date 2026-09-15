// Command void is VOID's professional CLI and API server entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"void/internal/api"
	"void/internal/behavior"
	"void/internal/config"
	"void/internal/entity"
	"void/internal/export"
	"void/internal/metrics"
	"void/internal/population"
	"void/internal/rng"
	"void/internal/simulation"
	"void/internal/world"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "serve":
		cmdServe()
	case "create-world":
		cmdCreateWorld(args)
	case "run", "run-headless":
		cmdRun(args, cmd == "run-headless")
	case "generate-dataset":
		cmdGenerateDataset(args)
	case "benchmark":
		cmdBenchmark(args)
	case "version":
		fmt.Println("VOID — Synthetic Reality Generator v1.0.0")
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`VOID — Synthetic Reality Generator

Usage:
  void serve                          Start the API + WebSocket server (headless-capable control plane)
  void create-world --seed N --size S Create and print a deterministic World
  void run --agents N --ticks T       Run an in-process simulation to completion
  void run-headless --agents N --ticks T  Same as run, but never touches stdout beyond a final summary
  void generate-dataset --agents N --out FILE --format json|csv|ndjson  Generate + export a synthetic population
  void benchmark --agents N --ticks T Report tick rate / throughput for the given population size
  void version                        Print version

Environment variables (see internal/config): VOID_HTTP_ADDR, VOID_DATA_DIR, VOID_WORKER_COUNT, VOID_JWT_SECRET, ...`)
}

func cmdServe() {
	cfg := config.Load()
	app := api.NewApp(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}

func cmdCreateWorld(args []string) {
	fs := parseFlags(args)
	seed := fs.int64("seed", time.Now().UnixNano())
	size := fs.str("size", "medium")
	name := fs.str("name", "Untitled World")
	wtype := fs.str("type", "generic")

	cfg := world.Config{Name: name, Type: wtype, Seed: seed, Size: size, StartTime: time.Now()}
	w, err := world.NewBuilder().Build(fmt.Sprintf("world-%d", time.Now().UnixNano()), cfg)
	must(err)
	fmt.Printf("Created world %q (seed=%d, size=%s, regions=%d)\n", w.ID, seed, size, len(w.Config.Regions))
}

func cmdRun(args []string, headless bool) {
	fs := parseFlags(args)
	agents := fs.int("agents", 10_000)
	ticks := fs.int64("ticks", 100)
	seed := fs.int64("seed", 42)
	workers := fs.int("workers", 8)

	reg := entity.NewRegistry()
	beh := behavior.NewEngine()
	met := metrics.NewRegistry()

	wcfg := world.Config{Name: "cli-world", Type: "generic", Seed: seed, Size: "medium", StartTime: time.Now()}
	w, err := world.NewBuilder().Build("cli-world", wcfg)
	must(err)

	scfg := simulation.Config{ID: "cli-sim", WorldID: w.ID, Seed: seed, MaxTicks: ticks, WorkerCount: workers, Headless: headless}
	eng := simulation.New(scfg, w, beh, met)

	gen := population.NewGenerator(reg)
	root := rng.New(seed)
	_, err = gen.Generate(root, population.Spec{
		Archetype: "customer", Count: agents,
		Attributes: []population.AttributeSpec{
			{Name: "loyalty", Distribution: population.Distribution{Kind: population.Uniform, Min: 0, Max: 1}},
		},
	}, func(e *entity.Entity) { eng.AddEntity(e) })
	must(err)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	must(eng.Run(ctx))
	elapsed := time.Since(start)

	fmt.Printf("Simulation complete: %d agents, %d ticks in %s (%.1f ticks/sec)\n",
		agents, ticks, elapsed.Round(time.Millisecond), float64(ticks)/elapsed.Seconds())
}

func cmdGenerateDataset(args []string) {
	fs := parseFlags(args)
	agents := fs.int("agents", 10_000)
	out := fs.str("out", "dataset.json")
	format := fs.str("format", "json")
	seed := fs.int64("seed", 42)

	reg := entity.NewRegistry()
	gen := population.NewGenerator(reg)
	root := rng.New(seed)

	var entities []entity.Entity
	_, err := gen.Generate(root, population.Spec{
		Archetype: "customer", Count: agents,
		Attributes: []population.AttributeSpec{
			{Name: "income", Distribution: population.Distribution{Kind: population.Normal, Mean: 4000, Stddev: 1200, Min: 0, Max: 50000}},
			{Name: "loyalty", Distribution: population.Distribution{Kind: population.Uniform, Min: 0, Max: 1}},
		},
	}, func(e *entity.Entity) { entities = append(entities, e.Snapshot()) })
	must(err)

	switch format {
	case "csv":
		must(export.ToCSV(out, entities))
	case "ndjson":
		must(export.ToNDJSON(out, entities))
	default:
		must(export.ToJSON(out, entities))
	}
	fmt.Printf("Wrote %d entities to %s (%s)\n", len(entities), out, format)
}

func cmdBenchmark(args []string) {
	fs := parseFlags(args)
	agents := fs.int("agents", 100_000)
	ticks := fs.int64("ticks", 50)
	workers := fs.int("workers", 8)

	reg := entity.NewRegistry()
	beh := behavior.NewEngine()
	met := metrics.NewRegistry()
	wcfg := world.Config{Name: "bench-world", Type: "generic", Seed: 7, Size: "large", StartTime: time.Now()}
	w, err := world.NewBuilder().Build("bench-world", wcfg)
	must(err)

	scfg := simulation.Config{ID: "bench-sim", WorldID: w.ID, Seed: 7, MaxTicks: ticks, WorkerCount: workers, Headless: true}
	eng := simulation.New(scfg, w, beh, met)

	gen := population.NewGenerator(reg)
	root := rng.New(7)
	_, err = gen.Generate(root, population.Spec{Archetype: "user", Count: agents}, func(e *entity.Entity) { eng.AddEntity(e) })
	must(err)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	must(eng.Run(ctx))
	elapsed := time.Since(start)

	fmt.Printf("agents=%d ticks=%d workers=%d elapsed=%s ticks/sec=%.2f\n",
		agents, ticks, workers, elapsed.Round(time.Millisecond), float64(ticks)/elapsed.Seconds())
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
