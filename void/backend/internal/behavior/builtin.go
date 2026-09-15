package behavior

import (
	"void/internal/entity"
	"void/internal/rng"
)

func builtinRuleSets() []*RuleSet {
	return []*RuleSet{
		customerRuleSet(),
		userRuleSet(),
		employeeRuleSet(),
		companyRuleSet(),
		productRuleSet(),
		accountRuleSet(),
		deviceRuleSet(),
		vehicleRuleSet(),
		serverRuleSet(),
		socialRuleSet(),
		driverRuleSet(),
	}
}

func customerRuleSet() *RuleSet {
	return &RuleSet{
		Name: "customer_default",
		Rules: []Rule{
			{
				Name: "purchase",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.State["status"] == "active" && r.Bool(0.05+0.1*e.Attr("loyalty"))
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					amount := r.Normal(50, 20)
					if amount < 1 {
						amount = 1
					}
					e.Remember(tick, "purchase", map[string]interface{}{"amount": amount})
					return &Action{Name: "purchase", EventType: "purchase", Payload: map[string]interface{}{"amount": amount}}
				},
				Weight: 1,
			},
			{
				Name: "complain",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return r.Bool(0.01 * (1 - e.Attr("loyalty")))
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "complain", EventType: "complaint"}
				},
			},
			{
				Name: "switch_brand",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.Attr("loyalty") < 0.2 && r.Bool(0.02)
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.SetState("status", "churned")
					return &Action{Name: "churn", EventType: "customer_churn"}
				},
			},
		},
		UtilityScore: func(e *entity.Entity, g entity.Goal, tick int64) float64 {
			return g.Priority * (1 - g.Progress)
		},
	}
}

func userRuleSet() *RuleSet {
	return &RuleSet{
		Name: "user_default",
		Rules: []Rule{
			{
				Name: "login",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return r.Bool(0.2 + 0.3*e.Attr("engagement"))
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.Remember(tick, "login", nil)
					return &Action{Name: "login", EventType: "login"}
				},
			},
			{
				Name: "login_failure",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return r.Bool(0.01) },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "login_failure", EventType: "login_failure"}
				},
			},
		},
	}
}

func employeeRuleSet() *RuleSet {
	return &RuleSet{
		Name: "employee_default",
		Rules: []Rule{
			{
				Name: "resign",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.Attr("satisfaction") < 0.25 && r.Bool(0.003)
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.SetState("status", "resigned")
					return &Action{Name: "resign", EventType: "employee_resignation"}
				},
			},
		},
	}
}

func companyRuleSet() *RuleSet {
	return &RuleSet{
		Name: "company_default",
		Rules: []Rule{
			{
				Name: "price_change",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return r.Bool(0.01) },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "price_change", EventType: "price_change"}
				},
			},
			{
				Name: "hire",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.Resources["cash"] > 50000 && r.Bool(0.02)
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "hire", EventType: "employee_hiring"}
				},
			},
		},
	}
}

func productRuleSet() *RuleSet {
	return &RuleSet{
		Name: "product_default",
		Rules: []Rule{
			{
				Name: "popularity_drift",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return true },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					delta := r.Normal(0, 0.01)
					e.SetAttr("popularity", clamp01(e.Attr("popularity")+delta))
					return nil
				},
			},
		},
	}
}

func accountRuleSet() *RuleSet {
	return &RuleSet{
		Name: "account_default",
		Rules: []Rule{
			{
				Name: "transaction",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return r.Bool(0.1) },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					amt := r.Normal(0, 100)
					fraud := r.Bool(0.002)
					return &Action{Name: "transaction", EventType: "payment", Payload: map[string]interface{}{"amount": amt, "fraud": fraud}}
				},
			},
		},
	}
}

func deviceRuleSet() *RuleSet {
	return &RuleSet{
		Name: "device_default",
		Rules: []Rule{
			{
				Name: "telemetry",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return true },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.SetAttr("battery", clamp01(e.Attr("battery")-r.Uniform(0, 0.5)))
					if e.Attr("battery") <= 0.05 && r.Bool(0.3) {
						return &Action{Name: "low_battery", EventType: "device_low_battery"}
					}
					return nil
				},
			},
		},
	}
}

func vehicleRuleSet() *RuleSet {
	return &RuleSet{
		Name: "vehicle_default",
		Rules: []Rule{
			{
				Name: "trip",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return r.Bool(0.15) },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.SetAttr("fuel", clamp01(e.Attr("fuel")-r.Uniform(0.01, 0.05)))
					return &Action{Name: "trip", EventType: "vehicle_trip"}
				},
			},
		},
	}
}

func serverRuleSet() *RuleSet {
	return &RuleSet{
		Name: "server_default",
		Rules: []Rule{
			{
				Name: "crash",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.Attr("cpu_load") > 0.95 && r.Bool(0.05)
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					e.SetState("status", "crashed")
					return &Action{Name: "crash", EventType: "server_crash"}
				},
			},
		},
	}
}

func socialRuleSet() *RuleSet {
	return &RuleSet{
		Name: "social_default",
		Rules: []Rule{
			{
				Name: "post",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return r.Bool(e.Attr("post_freq"))
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "post", EventType: "post"}
				},
			},
			{
				Name: "follow",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool { return r.Bool(0.02) },
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "follow", EventType: "follow"}
				},
			},
		},
	}
}

func driverRuleSet() *RuleSet {
	return &RuleSet{
		Name: "driver_default",
		Rules: []Rule{
			{
				Name: "accept_ride",
				Condition: func(e *entity.Entity, tick int64, r *rng.Source) bool {
					return e.Attr("availability") > 0 && r.Bool(0.3)
				},
				Action: func(e *entity.Entity, tick int64, r *rng.Source) *Action {
					return &Action{Name: "accept_ride", EventType: "ride_accepted"}
				},
			},
		},
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
