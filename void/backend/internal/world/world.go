// Package world implements the World Builder: a seeded, deterministic
// description of a synthetic reality - its type, size, population,
// regions/cities, organizations, resources, economy, network, rules, time
// scale and initial conditions.
package world

import "time"

// Region models a city/area with its own local economy + population.
type Region struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Population int     `json:"population"`
	WealthIdx  float64 `json:"wealth_index"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

// EconomyConfig configures the synthetic macro-economy.
type EconomyConfig struct {
	BaseCurrency   string  `json:"base_currency"`
	InflationRate  float64 `json:"inflation_rate"`
	InterestRate   float64 `json:"interest_rate"`
	UnemploymentPc float64 `json:"unemployment_pct"`
}

// Rule is a global world constraint/policy (e.g. max transaction amount,
// content moderation, fraud thresholds) available to behaviors/events.
type Rule struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Config is the full, user-specified World definition (input to Builder).
type Config struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // e.g. "ecommerce", "banking", "social", "iot", "smart_city"
	Seed        int64         `json:"seed"`
	Size        string        `json:"size"` // "small" | "medium" | "large" | "massive"
	Regions     []Region      `json:"regions"`
	Economy     EconomyConfig `json:"economy"`
	Rules       []Rule        `json:"rules"`
	TimeScale   float64       `json:"time_scale"` // ticks per simulated hour multiplier
	TickUnit    string        `json:"tick_unit"`  // "second" | "minute" | "hour" | "day"
	StartTime   time.Time     `json:"start_time"`
	Description string        `json:"description"`
}

// World is the materialized, running instance of a Config.
type World struct {
	ID        string    `json:"id"`
	Config    Config    `json:"config"`
	CreatedAt time.Time `json:"created_at"`
	Tick      int64     `json:"tick"`
	Version   int       `json:"version"`
}

// New creates a fresh World from a Config with a generated ID.
func New(id string, cfg Config) *World {
	return &World{
		ID:        id,
		Config:    cfg,
		CreatedAt: time.Now(),
		Version:   1,
	}
}
