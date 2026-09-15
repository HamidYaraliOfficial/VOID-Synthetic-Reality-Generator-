// Package iot implements VOID's IoT / Smart City Simulation Mode: sensors,
// vehicles, buildings, cameras, energy grids and traffic signals modeled as
// a lightweight digital twin, suitable for Smart Infrastructure testing.
package iot

import (
	"sync"

	"void/internal/rng"
)

type SensorKind string

const (
	SensorTemperature SensorKind = "temperature"
	SensorHumidity    SensorKind = "humidity"
	SensorTraffic     SensorKind = "traffic"
	SensorEnergy      SensorKind = "energy"
	SensorAirQuality  SensorKind = "air_quality"
	SensorOccupancy   SensorKind = "occupancy"
)

type Sensor struct {
	ID       string
	Kind     SensorKind
	Location [2]float64 // lat, lng
	Value    float64
	Online   bool
}

type TrafficSignal struct {
	ID    string
	State string // "red" | "yellow" | "green"
	QueueLength int
}

// CityModel aggregates all IoT entities for one simulated smart city.
type CityModel struct {
	mu             sync.RWMutex
	Sensors        map[string]*Sensor
	Signals        map[string]*TrafficSignal
	EnergyDemandMW float64
	EnergySupplyMW float64
}

func NewCityModel() *CityModel {
	return &CityModel{Sensors: map[string]*Sensor{}, Signals: map[string]*TrafficSignal{}}
}

func (c *CityModel) AddSensor(s *Sensor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Sensors[s.ID] = s
}

func (c *CityModel) AddSignal(sig *TrafficSignal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Signals[sig.ID] = sig
}

// TickSensors updates every sensor's reading with small correlated drift
// plus occasional dropout (device goes offline), for realistic telemetry.
func (c *CityModel) TickSensors(r *rng.Source) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range c.Sensors {
		if r.Bool(0.002) {
			s.Online = !s.Online
		}
		if !s.Online {
			continue
		}
		switch s.Kind {
		case SensorTemperature:
			s.Value += r.Normal(0, 0.3)
		case SensorHumidity:
			s.Value = clamp(s.Value+r.Normal(0, 0.5), 0, 100)
		case SensorTraffic:
			s.Value = clamp(s.Value+r.Normal(0, 0.05), 0, 1)
		case SensorEnergy:
			s.Value = clamp(s.Value+r.Normal(0, 2), 0, 10000)
		case SensorAirQuality:
			s.Value = clamp(s.Value+r.Normal(0, 1), 0, 500)
		case SensorOccupancy:
			s.Value = clamp(s.Value+r.Normal(0, 0.02), 0, 1)
		}
	}
}

// TickSignals rotates traffic signal state and adjusts queue length based
// on nearby traffic-sensor load.
func (c *CityModel) TickSignals(r *rng.Source) {
	c.mu.Lock()
	defer c.mu.Unlock()
	order := map[string]string{"red": "green", "green": "yellow", "yellow": "red"}
	for _, sig := range c.Signals {
		if r.Bool(0.1) {
			sig.State = order[sig.State]
		}
		if sig.State == "red" {
			sig.QueueLength += r.Intn(3)
		} else if sig.QueueLength > 0 {
			sig.QueueLength -= r.Intn(3)
			if sig.QueueLength < 0 {
				sig.QueueLength = 0
			}
		}
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
