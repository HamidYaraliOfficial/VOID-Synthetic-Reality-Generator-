package entity

// Registry holds all known Archetypes (built-in + user/plugin defined).
type Registry struct {
	items map[string]*Archetype
}

func NewRegistry() *Registry {
	r := &Registry{items: map[string]*Archetype{}}
	for _, a := range builtinArchetypes() {
		r.Register(a)
	}
	return r
}

func (r *Registry) Register(a *Archetype) { r.items[a.Name] = a }

func (r *Registry) Get(name string) (*Archetype, bool) {
	a, ok := r.items[name]
	return a, ok
}

func (r *Registry) List() []*Archetype {
	out := make([]*Archetype, 0, len(r.items))
	for _, a := range r.items {
		out = append(out, a)
	}
	return out
}

// builtinArchetypes ships VOID with the entity types requested by every
// bundled scenario template (e-commerce, banking, social, IoT, logistics...).
func builtinArchetypes() []*Archetype {
	return []*Archetype{
		{
			Name: "customer", Description: "A retail/e-commerce customer",
			DefaultAttrs: map[string]float64{"income": 4000, "loyalty": 0.5, "price_sensitivity": 0.5},
			DefaultTags:  []string{"person"},
			DefaultGoals: []Goal{{Name: "find_value", Priority: 0.7}, {Name: "stay_satisfied", Priority: 0.6}},
			ResourceTemplate: map[string]float64{"wallet": 500},
			BehaviorRuleSet:  "customer_default",
		},
		{
			Name: "user", Description: "A generic app/platform user",
			DefaultAttrs: map[string]float64{"engagement": 0.4, "churn_risk": 0.1},
			DefaultTags:  []string{"person"},
			ResourceTemplate: map[string]float64{"session_tokens": 0},
			BehaviorRuleSet:  "user_default",
		},
		{
			Name: "employee", Description: "A company employee",
			DefaultAttrs: map[string]float64{"productivity": 0.6, "satisfaction": 0.6, "salary": 3500},
			DefaultTags:  []string{"person"},
			BehaviorRuleSet: "employee_default",
		},
		{
			Name: "company", Description: "A business organization",
			DefaultAttrs: map[string]float64{"revenue": 0, "cost": 0, "market_share": 0.01, "reputation": 0.5},
			ResourceTemplate: map[string]float64{"cash": 1000000, "inventory": 10000},
			BehaviorRuleSet: "company_default",
		},
		{
			Name: "product", Description: "A sellable product/SKU",
			DefaultAttrs: map[string]float64{"price": 20, "quality": 0.5, "popularity": 0.1},
			BehaviorRuleSet: "product_default",
		},
		{
			Name: "account", Description: "A financial/bank account",
			DefaultAttrs:     map[string]float64{"credit_score": 650},
			ResourceTemplate: map[string]float64{"balance": 1000},
			BehaviorRuleSet:  "account_default",
		},
		{
			Name: "device", Description: "An IoT / smart device",
			DefaultAttrs: map[string]float64{"battery": 100, "signal": 1, "health": 1},
			BehaviorRuleSet: "device_default",
		},
		{
			Name: "vehicle", Description: "A vehicle in a fleet / traffic simulation",
			DefaultAttrs: map[string]float64{"fuel": 100, "speed": 0, "wear": 0},
			BehaviorRuleSet: "vehicle_default",
		},
		{
			Name: "server", Description: "A compute node / microservice instance",
			DefaultAttrs: map[string]float64{"cpu_load": 0.1, "mem_load": 0.2, "error_rate": 0},
			BehaviorRuleSet: "server_default",
		},
		{
			Name: "social_account", Description: "A social network account",
			DefaultAttrs: map[string]float64{"followers": 0, "influence": 0.01, "post_freq": 0.1},
			BehaviorRuleSet: "social_default",
		},
		{
			Name: "location", Description: "A city / region / point of interest",
			DefaultAttrs: map[string]float64{"population": 0, "wealth_index": 0.5, "traffic": 0.1},
		},
		{
			Name: "driver", Description: "A logistics/rideshare driver",
			DefaultAttrs: map[string]float64{"rating": 4.8, "availability": 1},
			BehaviorRuleSet: "driver_default",
		},
	}
}
