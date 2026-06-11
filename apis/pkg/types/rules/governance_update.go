/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, Tencent. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package rules

type GovernanceRuleUpdates struct {
	RouterRules          []*RouterConfig
	RateLimitRules       []*RateLimit
	CircuitBreakerRules  []*CircuitBreakerRule
	FaultDetectRules     []*FaultDetectRule
	LosslessRules        []*LosslessRule
	TrafficSecurityRules []*TrafficGovernanceRule
	TrafficMirrorRules   []*TrafficGovernanceRule
	TrafficMockRules     []*TrafficGovernanceRule
	LaneGroups           map[string]*LaneGroup
}

type GovernanceRuleReleaseUpdates struct {
	RouterRules          []*RouterRuleRelease
	RateLimitRules       []*RateLimitRelease
	CircuitBreakerRules  []*CircuitBreakerRelease
	FaultDetectRules     []*FaultDetectRelease
	LosslessRules        []*LosslessRuleRelease
	TrafficSecurityRules []*TrafficGovernanceRuleRelease
	TrafficMirrorRules   []*TrafficGovernanceRuleRelease
	TrafficMockRules     []*TrafficGovernanceRuleRelease
	LaneGroupRules       []*LaneGroupRelease
}
