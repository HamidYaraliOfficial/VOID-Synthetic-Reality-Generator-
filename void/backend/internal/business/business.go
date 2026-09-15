// Package business implements VOID's Business Simulation Engine: companies
// compete on price/quality/marketing for a shared customer population,
// track revenue/cost/inventory, and can be pushed through growth or crisis
// scenarios to study business outcomes.
package business

import "sync"

type Company struct {
	ID           string
	Name         string
	Cash         float64
	Revenue      float64
	Cost         float64
	MarketShare  float64
	Reputation   float64
	EmployeeIDs  []string
	ProductIDs   []string
	WarehouseQty map[string]float64 // productID -> quantity
}

// Market coordinates competing companies over a shared product/customer base.
type Market struct {
	mu        sync.RWMutex
	Companies map[string]*Company
}

func NewMarket() *Market { return &Market{Companies: map[string]*Company{}} }

func (m *Market) AddCompany(c *Company) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.WarehouseQty == nil {
		c.WarehouseQty = map[string]float64{}
	}
	m.Companies[c.ID] = c
}

// RecordSale books revenue for a company and depletes warehouse inventory.
func (m *Market) RecordSale(companyID, productID string, quantity, unitPrice float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.Companies[companyID]
	if !ok {
		return
	}
	c.Revenue += quantity * unitPrice
	c.Cash += quantity * unitPrice
	if c.WarehouseQty[productID] >= quantity {
		c.WarehouseQty[productID] -= quantity
	} else {
		c.WarehouseQty[productID] = 0
	}
}

// RecordCost books an operating cost (salaries, logistics, marketing, ...).
func (m *Market) RecordCost(companyID string, amount float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.Companies[companyID]; ok {
		c.Cost += amount
		c.Cash -= amount
	}
}

// RecomputeMarketShare distributes market share proportionally to revenue
// across all tracked companies - a simple but effective approximation for
// dashboard purposes.
func (m *Market) RecomputeMarketShare() {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0.0
	for _, c := range m.Companies {
		total += c.Revenue
	}
	if total <= 0 {
		return
	}
	for _, c := range m.Companies {
		c.MarketShare = c.Revenue / total
	}
}

func (m *Market) Snapshot() []Company {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Company, 0, len(m.Companies))
	for _, c := range m.Companies {
		out = append(out, *c)
	}
	return out
}
