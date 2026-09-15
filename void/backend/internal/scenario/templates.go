package scenario

// Templates returns VOID's built-in Scenario Template Library, covering
// both the generic "chaos/growth" scenarios and ready-made vertical
// templates (E-commerce, Banking, SaaS, Social Network, IoT, Logistics,
// Cyber Range, Smart City, Cloud Infrastructure).
func Templates() []*Scenario {
	return []*Scenario{
		{
			ID: "sudden_user_surge", Name: "Sudden User Surge", Category: "growth",
			Description: "A sharp, sustained spike in new signups and traffic.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Signup rate x10", ParamOverrides: map[string]float64{"signup_rate_multiplier": 10}},
			},
		},
		{
			ID: "service_attack", Name: "Service Under Attack", Category: "chaos",
			Description: "DDoS-like burst of malicious requests targeting the service layer.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Malicious request flood", ParamOverrides: map[string]float64{"malicious_traffic_multiplier": 50}},
			},
		},
		{
			ID: "company_bankruptcy", Name: "Company Bankruptcy", Category: "market",
			Description: "A simulated company's cash reserves collapse, triggering layoffs and closures.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Cash shock", ParamOverrides: map[string]float64{"cash_shock_pct": -0.9}},
			},
		},
		{
			ID: "viral_product_growth", Name: "Viral Product Growth", Category: "growth",
			Description: "A product's popularity grows virally through the social graph.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Virality coefficient boost", ParamOverrides: map[string]float64{"virality_k": 2.5}},
			},
		},
		{
			ID: "database_crash", Name: "Database Crash", Category: "infra",
			Description: "Primary datastore becomes unavailable, testing failover and degraded modes.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "DB unavailable", ParamOverrides: map[string]float64{"db_available": 0}},
			},
		},
		{
			ID: "network_outage", Name: "Network Disruption", Category: "infra",
			Description: "Elevated latency and packet loss across the network topology.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Latency/packet-loss injected", ParamOverrides: map[string]float64{"latency_ms": 800, "packet_loss_pct": 15}},
			},
		},
		{
			ID: "buying_frenzy", Name: "Buying Frenzy", Category: "market",
			Description: "Flash-sale-style rush that stresses inventory and checkout systems.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Purchase rate x8", ParamOverrides: map[string]float64{"purchase_rate_multiplier": 8}},
			},
		},
		{
			ID: "fraud_spike", Name: "Fraud Spike", Category: "fraud",
			Description: "Coordinated increase in fraudulent transactions to test detection systems.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Fraud rate x20", ParamOverrides: map[string]float64{"fraud_rate_multiplier": 20}},
			},
		},
		{
			ID: "resource_shortage", Name: "Resource Shortage", Category: "market",
			Description: "Key resources/inventory become scarce, forcing rationing behavior.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Inventory shock", ParamOverrides: map[string]float64{"inventory_shock_pct": -0.7}},
			},
		},
		{
			ID: "customer_behavior_shift", Name: "Customer Behavior Shift", Category: "market",
			Description: "Population-wide shift in price sensitivity and loyalty.",
			Injections: []Injection{
				{AtTickOffset: 0, Description: "Loyalty/price-sensitivity shift", ParamOverrides: map[string]float64{"loyalty_delta": -0.2, "price_sensitivity_delta": 0.3}},
			},
		},

		// Vertical starter templates
		{ID: "template_ecommerce", Name: "E-commerce Starter", Category: "template", Description: "Customers, products, orders, payments, reviews."},
		{ID: "template_banking", Name: "Banking Starter", Category: "template", Description: "Accounts, transactions, fraud detection, credit scoring."},
		{ID: "template_saas", Name: "SaaS Starter", Category: "template", Description: "Users, subscriptions, churn, feature usage."},
		{ID: "template_social", Name: "Social Network Starter", Category: "template", Description: "Accounts, posts, follows, virality, communities."},
		{ID: "template_iot", Name: "IoT Starter", Category: "template", Description: "Devices, sensors, telemetry, failures."},
		{ID: "template_logistics", Name: "Logistics Starter", Category: "template", Description: "Drivers, vehicles, routes, deliveries."},
		{ID: "template_cyber_range", Name: "Cyber Range Starter", Category: "template", Description: "Servers, network topology, attackers, defenders."},
		{ID: "template_smart_city", Name: "Smart City Starter", Category: "template", Description: "Traffic, energy, sensors, citizens, infrastructure."},
		{ID: "template_cloud_infra", Name: "Cloud Infrastructure Starter", Category: "template", Description: "Services, dependencies, load, failures, scaling."},
	}
}
