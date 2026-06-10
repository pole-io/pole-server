/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, Tencent. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package sqldb

import (
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func (s *stableStore) GetMoreGovernanceRuleUpdates(mtime time.Time, firstUpdate bool) (*rules.GovernanceRuleUpdates, error) {
	repo := newGovernanceRuleRepository(s.master, s.slave)
	records, err := repo.GetMoreRules(mtime, firstUpdate)
	if err != nil {
		return nil, err
	}
	out := &rules.GovernanceRuleUpdates{
		LaneGroups: map[string]*rules.LaneGroup{},
	}
	for i := range records {
		switch records[i].RuleType {
		case governanceRuleTypeRoute:
			item, err := governanceRuleRecordToRouterConfig(records[i])
			if err != nil {
				return nil, err
			}
			out.RouterRules = append(out.RouterRules, item)
		case governanceRuleTypeRateLimit:
			item, err := governanceRuleRecordToRateLimit(records[i])
			if err != nil {
				return nil, err
			}
			out.RateLimitRules = append(out.RateLimitRules, item)
		case governanceRuleTypeCircuitBreaker:
			item, err := governanceRuleRecordToCircuitBreakerRule(records[i])
			if err != nil {
				return nil, err
			}
			out.CircuitBreakerRules = append(out.CircuitBreakerRules, item)
		case governanceRuleTypeFaultDetect:
			item, err := governanceRuleRecordToFaultDetectRule(records[i])
			if err != nil {
				return nil, err
			}
			out.FaultDetectRules = append(out.FaultDetectRules, item)
		case governanceRuleTypeLossless:
			item, err := governanceRuleRecordToLosslessRule(records[i])
			if err != nil {
				return nil, err
			}
			out.LosslessRules = append(out.LosslessRules, item)
		case governanceRuleTypeLaneGroup:
			item, err := governanceRuleRecordToLaneGroup(records[i])
			if err != nil {
				return nil, err
			}
			if item != nil {
				out.LaneGroups[item.ID] = item
			}
		}
	}
	return out, nil
}

func (s *stableStore) GetMoreGovernanceRuleReleaseUpdates(mtime time.Time, firstUpdate bool) (*rules.GovernanceRuleReleaseUpdates, error) {
	repo := newGovernanceRuleRepository(s.master, s.slave)
	records, err := repo.GetMoreReleases(mtime, firstUpdate)
	if err != nil {
		return nil, err
	}
	out := &rules.GovernanceRuleReleaseUpdates{}
	for i := range records {
		switch records[i].RuleType {
		case governanceRuleTypeRoute:
			item, err := governanceRuleReleaseRecordToRouterRuleRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.RouterRules = append(out.RouterRules, item)
		case governanceRuleTypeRateLimit:
			item, err := governanceRuleReleaseRecordToRateLimitRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.RateLimitRules = append(out.RateLimitRules, item)
		case governanceRuleTypeCircuitBreaker:
			item, err := governanceRuleReleaseRecordToCircuitBreakerRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.CircuitBreakerRules = append(out.CircuitBreakerRules, item)
		case governanceRuleTypeFaultDetect:
			item, err := governanceRuleReleaseRecordToFaultDetectRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.FaultDetectRules = append(out.FaultDetectRules, item)
		case governanceRuleTypeLossless:
			item, err := governanceRuleReleaseRecordToLosslessRuleRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.LosslessRules = append(out.LosslessRules, item)
		case governanceRuleTypeLaneGroup:
			item, err := governanceRuleReleaseRecordToLaneGroupRelease(records[i])
			if err != nil {
				return nil, err
			}
			out.LaneGroupRules = append(out.LaneGroupRules, item)
		}
	}
	return out, nil
}
