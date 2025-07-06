package rules

type RuleRelease struct {
	Id          string
	ReleaseName string
	Description string
	ReleaseType string
	Active      bool
	Version     uint64
	Valid       bool
}

type LaneGroupRelease struct {
	RuleRelease
	Rule *LaneGroup
}

type CustomRouteRelease struct {
	RuleRelease
	Rule *RouterConfig
}

type FaultDetectRelease struct {
	RuleRelease
	Rule *FaultDetectRule
}

type RateLimitRelease struct {
	RuleRelease
	Rule *RateLimit
}

type CircuitBreakerRelease struct {
	RuleRelease
	Rule *CircuitBreakerRule
}
