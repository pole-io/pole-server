/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package sqldb

import (
	"encoding/json"
	"strconv"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/encoding/protojson"
)

func routerConfigToGovernanceRuleRecord(conf *rules.RouterConfig) *governanceRuleRecord {
	return &governanceRuleRecord{
		ID:          conf.ID,
		RuleType:    governanceRuleTypeRoute,
		Namespace:   conf.Namespace,
		Name:        conf.Name,
		Policy:      conf.Policy,
		Config:      conf.Config,
		Rule:        marshalRouterConfig(conf),
		Priority:    int(conf.Priority),
		Enable:      boolToInt(conf.Enable),
		Revision:    conf.Revision,
		Description: conf.Description,
		Metadata:    marshalMetadata(conf.Metadata),
		Valid:       conf.Valid,
	}
}

func governanceRuleRecordToRouterConfig(record *governanceRuleRecord) (*rules.RouterConfig, error) {
	if record == nil {
		return nil, nil
	}
	conf := &rules.RouterConfig{}
	if record.Rule != "" {
		if err := json.Unmarshal([]byte(record.Rule), conf); err != nil {
			return nil, err
		}
	}
	conf.ID = record.ID
	conf.Namespace = record.Namespace
	conf.Name = record.Name
	conf.Policy = record.Policy
	conf.Config = record.Config
	conf.Priority = uint32(record.Priority)
	conf.Enable = record.Enable == 1
	conf.Revision = record.Revision
	conf.Description = record.Description
	conf.Valid = record.Valid
	conf.CreateTime = record.CreateTime
	conf.ModifyTime = record.ModifyTime
	conf.EnableTime = record.EnableTime
	return conf, nil
}

func routerRuleReleaseToGovernanceReleaseRecord(release *rules.RouterRuleRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeRoute,
		ReleaseName:  release.ReleaseName,
		RuleID:       release.RuleId,
		RuleName:     release.RuleName,
		Description:  release.Description,
		ReleaseType:  string(release.ReleaseType),
		Version:      release.Version,
		Active:       release.Active,
		ClientLabels: marshalClientLabels(release.ClientLabels),
		Valid:        release.Valid,
	}
	if release.Rule != nil && release.Rule.RouterConfig != nil {
		record.RuleID = utilsDefaultString(record.RuleID, release.Rule.ID)
		record.RuleName = utilsDefaultString(record.RuleName, release.Rule.Name)
		record.Namespace = release.Rule.Namespace
		record.Rule = marshalRouterConfig(release.Rule.RouterConfig)
	}
	return record
}

func governanceRuleReleaseRecordToRouterRuleRelease(record *governanceRuleReleaseRecord) (*rules.RouterRuleRelease, error) {
	if record == nil {
		return nil, nil
	}
	conf := &rules.RouterConfig{}
	if record.Rule != "" {
		if err := json.Unmarshal([]byte(record.Rule), conf); err != nil {
			return nil, err
		}
	}
	extend, err := conf.ToExpendRoutingConfig()
	if err != nil {
		return nil, err
	}
	return &rules.RouterRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          record.ID,
			ReleaseName: record.ReleaseName,
			RuleId:      record.RuleID,
			RuleName:    record.RuleName,
			Description: record.Description,
			Resource:    apimodel.RuleRelease_RouteRules,
			ReleaseType: rules.ReleaseType(record.ReleaseType),
			Active:      record.Active,
			Version:     record.Version,
			Valid:       record.Valid,
			Ctime:       record.CreateTime,
			Mtime:       record.ModifyTime,
		},
		Rule: extend,
	}, nil
}

func losslessRuleToGovernanceRuleRecord(rule *rules.LosslessRule) *governanceRuleRecord {
	ruleJSON := marshalLosslessRule(rule)
	return &governanceRuleRecord{
		ID:          rule.ID,
		RuleType:    governanceRuleTypeLossless,
		Namespace:   rule.Namespace,
		Name:        rule.Service,
		Service:     rule.Service,
		Enable:      1,
		Config:      ruleJSON,
		Rule:        ruleJSON,
		Revision:    rule.Revision,
		Description: rule.Description,
		Metadata:    marshalMetadata(rule.Metadata),
		Valid:       rule.Valid,
	}
}

func governanceRuleRecordToLosslessRule(record *governanceRuleRecord) (*rules.LosslessRule, error) {
	if record == nil {
		return nil, nil
	}
	rule := &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
	if record.Rule != "" {
		if err := protojson.Unmarshal([]byte(record.Rule), rule.Proto); err != nil {
			return nil, err
		}
	}
	rule.ID = record.ID
	rule.Namespace = record.Namespace
	rule.Service = record.Service
	if rule.Service == "" {
		rule.Service = record.Name
	}
	rule.Revision = record.Revision
	rule.Description = record.Description
	rule.Valid = record.Valid
	rule.CTime = record.CreateTime
	rule.MTime = record.ModifyTime
	return rule, nil
}

func losslessRuleReleaseToGovernanceReleaseRecord(release *rules.LosslessRuleRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeLossless,
		ReleaseName:  release.ReleaseName,
		RuleID:       release.RuleId,
		RuleName:     release.RuleName,
		Description:  release.Description,
		ReleaseType:  string(release.ReleaseType),
		Version:      release.Version,
		Active:       release.Active,
		ClientLabels: marshalClientLabels(release.ClientLabels),
		Valid:        release.Valid,
	}
	if release.Rule != nil {
		record.RuleID = utilsDefaultString(record.RuleID, release.Rule.ID)
		record.RuleName = utilsDefaultString(record.RuleName, release.Rule.Service)
		record.Namespace = release.Rule.Namespace
		record.Service = release.Rule.Service
		record.Rule = marshalLosslessRule(release.Rule)
	}
	return record
}

func governanceRuleReleaseRecordToLosslessRuleRelease(record *governanceRuleReleaseRecord) (*rules.LosslessRuleRelease, error) {
	if record == nil {
		return nil, nil
	}
	rule := &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
	if record.Rule != "" {
		if err := protojson.Unmarshal([]byte(record.Rule), rule.Proto); err != nil {
			return nil, err
		}
	}
	rule.ID = record.RuleID
	rule.Namespace = record.Namespace
	rule.Service = record.Service
	rule.Valid = true
	return &rules.LosslessRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          record.ID,
			ReleaseName: record.ReleaseName,
			RuleId:      record.RuleID,
			RuleName:    record.RuleName,
			Description: record.Description,
			Resource:    apimodel.RuleRelease_LosslessRules,
			ReleaseType: rules.ReleaseType(record.ReleaseType),
			Active:      record.Active,
			Version:     record.Version,
			Valid:       record.Valid,
			Ctime:       record.CreateTime,
			Mtime:       record.ModifyTime,
		},
		Rule: rule,
	}, nil
}

func rateLimitToGovernanceRuleRecord(limit *rules.RateLimit) *governanceRuleRecord {
	return &governanceRuleRecord{
		ID:        limit.ID,
		RuleType:  governanceRuleTypeRateLimit,
		Name:      limit.Name,
		ServiceID: limit.ServiceID,
		Method:    limit.Method,
		Labels:    limit.Labels,
		Priority:  int(limit.Priority),
		Enable:    boolToInt(!limit.Disable),
		Disable:   boolToInt(limit.Disable),
		Rule:      limit.Rule,
		Revision:  limit.Revision,
		Metadata:  marshalMetadata(limit.Metadata),
		Valid:     limit.Valid,
	}
}

func governanceRuleRecordToRateLimit(record *governanceRuleRecord) (*rules.RateLimit, error) {
	if record == nil {
		return nil, nil
	}
	limit := &rules.RateLimit{
		ID:         record.ID,
		Name:       record.Name,
		ServiceID:  record.ServiceID,
		Method:     record.Method,
		Labels:     record.Labels,
		Priority:   uint32(record.Priority),
		Rule:       record.Rule,
		Revision:   record.Revision,
		Disable:    record.Disable == 1,
		Valid:      record.Valid,
		CreateTime: record.CreateTime,
		ModifyTime: record.ModifyTime,
		EnableTime: record.EnableTime,
	}
	return limit, nil
}

func rateLimitReleaseToGovernanceReleaseRecord(release *rules.RateLimitRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeRateLimit,
		ReleaseName:  release.ReleaseName,
		RuleID:       release.RuleId,
		RuleName:     release.RuleName,
		Description:  release.Description,
		ReleaseType:  string(release.ReleaseType),
		Version:      release.Version,
		Active:       release.Active,
		ClientLabels: marshalClientLabels(release.ClientLabels),
		Valid:        release.Valid,
	}
	if release.Rule != nil {
		record.RuleID = utilsDefaultString(record.RuleID, release.Rule.ID)
		record.RuleName = utilsDefaultString(record.RuleName, release.Rule.Name)
		record.Rule = marshalRateLimit(release.Rule)
	}
	return record
}

func governanceRuleReleaseRecordToRateLimitRelease(record *governanceRuleReleaseRecord) (*rules.RateLimitRelease, error) {
	if record == nil {
		return nil, nil
	}
	limit := &rules.RateLimit{}
	if record.Rule != "" {
		if err := json.Unmarshal([]byte(record.Rule), limit); err != nil {
			return nil, err
		}
	}
	return &rules.RateLimitRelease{
		RuleRelease: rules.RuleRelease{
			Id:          record.ID,
			ReleaseName: record.ReleaseName,
			RuleId:      record.RuleID,
			RuleName:    record.RuleName,
			Description: record.Description,
			Resource:    apimodel.RuleRelease_RateLimitRules,
			ReleaseType: rules.ReleaseType(record.ReleaseType),
			Active:      record.Active,
			Version:     record.Version,
			Valid:       record.Valid,
			Ctime:       record.CreateTime,
			Mtime:       record.ModifyTime,
		},
		Rule: limit,
	}, nil
}

func circuitBreakerRuleToGovernanceRuleRecord(rule *rules.CircuitBreakerRule) *governanceRuleRecord {
	return &governanceRuleRecord{
		ID:           rule.ID,
		RuleType:     governanceRuleTypeCircuitBreaker,
		Namespace:    rule.Namespace,
		Name:         rule.Name,
		Enable:       boolToInt(rule.Enable),
		Level:        strconv.Itoa(rule.Level),
		SrcService:   rule.SrcService,
		SrcNamespace: rule.SrcNamespace,
		DstService:   rule.DstService,
		DstNamespace: rule.DstNamespace,
		DstMethod:    rule.DstMethod,
		Config:       rule.Rule,
		Rule:         rule.Rule,
		Revision:     rule.Revision,
		Description:  rule.Description,
		Metadata:     marshalMetadata(nil),
		Valid:        rule.Valid,
	}
}

func governanceRuleRecordToCircuitBreakerRule(record *governanceRuleRecord) (*rules.CircuitBreakerRule, error) {
	if record == nil {
		return nil, nil
	}
	level, _ := strconv.Atoi(record.Level)
	return &rules.CircuitBreakerRule{
		ID:           record.ID,
		Name:         record.Name,
		Namespace:    record.Namespace,
		Description:  record.Description,
		Level:        level,
		SrcService:   record.SrcService,
		SrcNamespace: record.SrcNamespace,
		DstService:   record.DstService,
		DstNamespace: record.DstNamespace,
		DstMethod:    record.DstMethod,
		Rule:         record.Rule,
		Revision:     record.Revision,
		Enable:       record.Enable == 1,
		Valid:        record.Valid,
		CreateTime:   record.CreateTime,
		ModifyTime:   record.ModifyTime,
		EnableTime:   record.EnableTime,
	}, nil
}

func circuitBreakerReleaseToGovernanceReleaseRecord(release *rules.CircuitBreakerRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeCircuitBreaker,
		ReleaseName:  release.ReleaseName,
		RuleID:       release.RuleId,
		RuleName:     release.RuleName,
		Description:  release.Description,
		ReleaseType:  string(release.ReleaseType),
		Version:      release.Version,
		Active:       release.Active,
		ClientLabels: marshalClientLabels(release.ClientLabels),
		Valid:        release.Valid,
	}
	if release.Rule != nil {
		record.RuleID = utilsDefaultString(record.RuleID, release.Rule.ID)
		record.RuleName = utilsDefaultString(record.RuleName, release.Rule.Name)
		record.Namespace = release.Rule.Namespace
		record.Service = release.Rule.DstService
		record.Rule = marshalCircuitBreakerRule(release.Rule)
	}
	return record
}

func governanceRuleReleaseRecordToCircuitBreakerRelease(record *governanceRuleReleaseRecord) (*rules.CircuitBreakerRelease, error) {
	if record == nil {
		return nil, nil
	}
	rule := &rules.CircuitBreakerRule{}
	if record.Rule != "" {
		if err := json.Unmarshal([]byte(record.Rule), rule); err != nil {
			return nil, err
		}
	}
	return &rules.CircuitBreakerRelease{
		RuleRelease: rules.RuleRelease{
			Id:          record.ID,
			ReleaseName: record.ReleaseName,
			RuleId:      record.RuleID,
			RuleName:    record.RuleName,
			Description: record.Description,
			Resource:    apimodel.RuleRelease_CircuitBreakerRules,
			ReleaseType: rules.ReleaseType(record.ReleaseType),
			Active:      record.Active,
			Version:     record.Version,
			Valid:       record.Valid,
			Ctime:       record.CreateTime,
			Mtime:       record.ModifyTime,
		},
		Rule: rule,
	}, nil
}

func faultDetectRuleToGovernanceRuleRecord(rule *rules.FaultDetectRule) *governanceRuleRecord {
	return &governanceRuleRecord{
		ID:           rule.ID,
		RuleType:     governanceRuleTypeFaultDetect,
		Namespace:    rule.Namespace,
		Name:         rule.Name,
		Enable:       1,
		DstService:   rule.DstService,
		DstNamespace: rule.DstNamespace,
		DstMethod:    rule.DstMethod,
		Config:       rule.Rule,
		Rule:         rule.Rule,
		Revision:     rule.Revision,
		Description:  rule.Description,
		Metadata:     marshalMetadata(rule.Metadata),
		Valid:        rule.Valid,
	}
}

func governanceRuleRecordToFaultDetectRule(record *governanceRuleRecord) (*rules.FaultDetectRule, error) {
	if record == nil {
		return nil, nil
	}
	rule := &rules.FaultDetectRule{
		ID:           record.ID,
		Name:         record.Name,
		Namespace:    record.Namespace,
		Description:  record.Description,
		DstService:   record.DstService,
		DstNamespace: record.DstNamespace,
		DstMethod:    record.DstMethod,
		Rule:         record.Rule,
		Revision:     record.Revision,
		Valid:        record.Valid,
		CreateTime:   record.CreateTime,
		ModifyTime:   record.ModifyTime,
	}
	if record.Metadata != "" {
		rule.Metadata = map[string]string{}
		_ = json.Unmarshal([]byte(record.Metadata), &rule.Metadata)
	}
	return rule, nil
}

func faultDetectReleaseToGovernanceReleaseRecord(release *rules.FaultDetectRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeFaultDetect,
		ReleaseName:  release.ReleaseName,
		RuleID:       release.RuleId,
		RuleName:     release.RuleName,
		Description:  release.Description,
		ReleaseType:  string(release.ReleaseType),
		Version:      release.Version,
		Active:       release.Active,
		ClientLabels: marshalClientLabels(release.ClientLabels),
		Valid:        release.Valid,
	}
	if release.Rule != nil {
		record.RuleID = utilsDefaultString(record.RuleID, release.Rule.ID)
		record.RuleName = utilsDefaultString(record.RuleName, release.Rule.Name)
		record.Namespace = release.Rule.Namespace
		record.Service = release.Rule.DstService
		record.Rule = marshalFaultDetectRule(release.Rule)
	}
	return record
}

func governanceRuleReleaseRecordToFaultDetectRelease(record *governanceRuleReleaseRecord) (*rules.FaultDetectRelease, error) {
	if record == nil {
		return nil, nil
	}
	rule := &rules.FaultDetectRule{}
	if record.Rule != "" {
		if err := json.Unmarshal([]byte(record.Rule), rule); err != nil {
			return nil, err
		}
	}
	return &rules.FaultDetectRelease{
		RuleRelease: rules.RuleRelease{
			Id:          record.ID,
			ReleaseName: record.ReleaseName,
			RuleId:      record.RuleID,
			RuleName:    record.RuleName,
			Description: record.Description,
			Resource:    apimodel.RuleRelease_FaultDetectRules,
			ReleaseType: rules.ReleaseType(record.ReleaseType),
			Active:      record.Active,
			Version:     record.Version,
			Valid:       record.Valid,
			Ctime:       record.CreateTime,
			Mtime:       record.ModifyTime,
		},
		Rule: rule,
	}, nil
}

func marshalRouterConfig(conf *rules.RouterConfig) string {
	if conf == nil {
		return "{}"
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func marshalLosslessRule(rule *rules.LosslessRule) string {
	if rule == nil || rule.Proto == nil {
		return "{}"
	}
	data, err := protojson.Marshal(rule.Proto)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func marshalRateLimit(limit *rules.RateLimit) string {
	if limit == nil {
		return "{}"
	}
	data, err := json.Marshal(limit)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func marshalCircuitBreakerRule(rule *rules.CircuitBreakerRule) string {
	if rule == nil {
		return "{}"
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func marshalFaultDetectRule(rule *rules.FaultDetectRule) string {
	if rule == nil {
		return "{}"
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return "{}"
	}
	return string(data)
}
