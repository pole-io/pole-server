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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

var _ store.LaneStore = (*laneStore)(nil)

type laneStore struct {
	master *BaseDB
	slave  *BaseDB
	*governanceRuleRepository
}

// AddLaneGroup 添加泳道组
func (l *laneStore) AddLaneGroup(tx store.Tx, item *ruletypes.LaneGroup) error {
	if err := l.repo().CreateRule(tx, laneGroupToGovernanceRuleRecord(item)); err != nil {
		log.Error("[Store][Lane] add lane group", zap.String("id", item.ID),
			zap.String("name", item.Name), zap.Error(err))
		return err
	}
	return nil
}

// UpdateLaneGroup 更新泳道组
func (l *laneStore) UpdateLaneGroup(tx store.Tx, item *ruletypes.LaneGroup) error {
	if err := l.repo().UpdateRule(tx, laneGroupToGovernanceRuleRecord(item)); err != nil {
		log.Error("[Store][Lane] update lane group", zap.String("id", item.ID),
			zap.String("name", item.Name), zap.Error(err))
		return err
	}
	return nil
}

// GetLaneGroup 查询泳道组
func (l *laneStore) GetLaneGroup(namespace, name string) (*ruletypes.LaneGroup, error) {
	record, err := l.repo().GetRuleByName(governanceRuleTypeLaneGroup, namespace, name)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToLaneGroup(record)
}

// GetLaneGroupByID .
func (l *laneStore) GetLaneGroupByID(id string) (*ruletypes.LaneGroup, error) {
	record, err := l.repo().GetRuleByID(governanceRuleTypeLaneGroup, id)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleRecordToLaneGroup(record)
}

func (l *laneStore) LockLaneGroup(tx store.Tx, namespace, keyword string) (*ruletypes.LaneGroup, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if keyword == "" {
		return nil, ErrorMissingParams
	}
	record, err := l.repo().LockRule(tx, governanceRuleTypeLaneGroup, keyword, namespace, keyword)
	if err != nil {
		log.Error("[Store][Lane] lock one lane group", zap.String("keyword", keyword), zap.Error(err))
		return nil, store.Error(err)
	}
	return governanceRuleRecordToLaneGroup(record)
}

// GetLaneGroups 查询泳道组
func (l *laneStore) GetLaneGroups(filter map[string]string, offset, limit uint32) (uint32, []*ruletypes.LaneGroup, error) {
	count, records, err := l.repo().QueryRules(context.Background(), governanceRuleTypeLaneGroup, filter, offset, limit)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	result := make([]*ruletypes.LaneGroup, 0, len(records))
	for i := range records {
		group, err := governanceRuleRecordToLaneGroup(records[i])
		if err != nil {
			return 0, nil, store.Error(err)
		}
		if group != nil {
			result = append(result, group)
		}
	}
	return count, result, nil
}

// DeleteLaneGroup 删除泳道组
func (l *laneStore) DeleteLaneGroup(id string) error {
	err := l.master.processWithTransaction("DeleteLaneGroup", func(tx *BaseTx) error {
		if err := l.repo().DeleteRule(NewSqlDBTx(tx), governanceRuleTypeLaneGroup, id); err != nil {
			log.Error("[Store][Lane] delete lane group", zap.String("id", id), zap.Error(err))
			return err
		}
		return tx.Commit()
	})
	return store.Error(err)
}

// GetMoreLaneGroups 获取泳道规则列表到缓存层
func (l *laneStore) GetMoreLaneGroups(mtime time.Time, firstUpdate bool) (map[string]*ruletypes.LaneGroup, error) {
	if firstUpdate {
		mtime = time.Unix(0, 1)
	}
	records, err := l.repo().GetMoreRulesByType(governanceRuleTypeLaneGroup, mtime, firstUpdate)
	if err != nil {
		return nil, store.Error(err)
	}
	deltaGroups := make(map[string]*ruletypes.LaneGroup, len(records))
	for i := range records {
		group, err := governanceRuleRecordToLaneGroup(records[i])
		if err != nil {
			return nil, store.Error(err)
		}
		if group != nil {
			deltaGroups[group.Name] = group
		}
	}
	return deltaGroups, nil
}

// ActiveLaneGroup implements store.LaneStore.
func (l *laneStore) ActiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) error {
	return l.repo().ActiveRelease(tx, laneGroupReleaseToGovernanceReleaseRecord(release))
}

// GetLaneGroupVersions .
func (l *laneStore) GetLaneGroupVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	return l.repo().QueryReleaseVersions(ctx, governanceRuleTypeLaneGroup, model.RuleRelease_LaneRules, filter, offset, limit)
}

func (l *laneStore) GetReleaseLaneGroupRule(tx store.Tx, release *rules.RuleRelease) (*rules.LaneGroupRelease, error) {
	record, err := l.repo().GetRelease(tx, governanceRuleTypeLaneGroup, release)
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToLaneGroupRelease(record)
}

// GetActiveLaneGroup implements store.LaneStore.
func (l *laneStore) GetActiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) (*ruletypes.LaneGroupRelease, error) {
	record, err := l.repo().GetActiveRelease(tx, laneGroupReleaseToGovernanceReleaseRecord(release))
	if err != nil {
		return nil, store.Error(err)
	}
	return governanceRuleReleaseRecordToLaneGroupRelease(record)
}

// InactiveLaneGroup implements store.LaneStore.
func (l *laneStore) InactiveLaneGroup(tx store.Tx, release *ruletypes.LaneGroupRelease) error {
	return l.repo().InactiveRelease(tx, laneGroupReleaseToGovernanceReleaseRecord(release))
}

// PublishLaneGroup implements store.LaneStore.
func (l *laneStore) PublishLaneGroup(tx store.Tx, rule *ruletypes.LaneGroupRelease) error {
	if rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][lane] publish lane group missing some params")
	}
	return l.repo().PublishRelease(tx, laneGroupReleaseToGovernanceReleaseRecord(rule))
}

// GetMoreLaneGroupReleases implements store.LaneStore.
func (l *laneStore) GetMoreLaneGroupReleases(firstUpdate bool, mtime time.Time) ([]*ruletypes.LaneGroupRelease, error) {
	records, err := l.repo().GetMoreReleasesByType(governanceRuleTypeLaneGroup, firstUpdate, mtime)
	if err != nil {
		return nil, store.Error(err)
	}
	out := make([]*ruletypes.LaneGroupRelease, 0, len(records))
	for i := range records {
		release, err := governanceRuleReleaseRecordToLaneGroupRelease(records[i])
		if err != nil {
			return nil, err
		}
		out = append(out, release)
	}
	return out, nil
}

// GetLaneRule 查询泳道规则
func (l *laneStore) GetLaneRule(id string) (*ruletypes.LaneRule, error) {
	records, err := l.repo().GetMoreRulesByType(governanceRuleTypeLaneGroup, time.Unix(0, 1), true)
	if err != nil {
		return nil, store.Error(err)
	}
	for i := range records {
		group, err := governanceRuleRecordToLaneGroup(records[i])
		if err != nil {
			return nil, store.Error(err)
		}
		if group == nil {
			continue
		}
		if rule, ok := group.LaneRules[id]; ok {
			return rule, nil
		}
	}
	return nil, nil
}

// AddLaneRules 添加泳道规则
func (l *laneStore) AddLaneRules(tx store.Tx, rules []*ruletypes.LaneRule) error {
	if len(rules) == 0 {
		return nil
	}
	for i := range rules {
		item := rules[i]
		if err := l.updateLaneRuleAggregate(tx, item.Namespace, item.LaneGroup, func(group *ruletypes.LaneGroup) error {
			if group.LaneRules == nil {
				group.LaneRules = map[string]*ruletypes.LaneRule{}
			}
			group.LaneRules[item.ID] = item
			return nil
		}); err != nil {
			log.Error("[Store][Lane] add lane rule", zap.String("id", item.ID),
				zap.String("group", item.LaneGroup), zap.String("name", item.Name), zap.Error(err))
			return store.Error(err)
		}
	}
	return nil
}

// UpdateLaneRules 更新泳道规则
func (l *laneStore) UpdateLaneRules(tx store.Tx, rules []*ruletypes.LaneRule) error {
	if len(rules) == 0 {
		return nil
	}
	for i := range rules {
		item := rules[i]
		if err := l.updateLaneRuleAggregate(tx, item.Namespace, item.LaneGroup, func(group *ruletypes.LaneGroup) error {
			if group.LaneRules == nil {
				group.LaneRules = map[string]*ruletypes.LaneRule{}
			}
			oldRule := group.LaneRules[item.ID]
			if oldRule != nil && item.IsChangeEnable() {
				item.EnableTime = time.Now()
			} else if oldRule != nil {
				item.EnableTime = oldRule.EnableTime
			}
			group.LaneRules[item.ID] = item
			return nil
		}); err != nil {
			log.Error("[Store][Lane] update lane rule", zap.String("id", item.ID),
				zap.String("group", item.LaneGroup), zap.String("name", item.Name), zap.Error(err))
			return store.Error(err)
		}
	}
	return nil
}

// DeleteLaneRules 删除泳道规则
func (l *laneStore) DeleteLaneRules(tx store.Tx, namespace, group string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	deleteSet := make(map[string]struct{}, len(ids))
	for i := range ids {
		deleteSet[ids[i]] = struct{}{}
	}
	if err := l.updateLaneRuleAggregate(tx, namespace, group, func(group *ruletypes.LaneGroup) error {
		for id := range deleteSet {
			delete(group.LaneRules, id)
		}
		return nil
	}); err != nil {
		log.Error("[Store][Lane] delete lane rule", zap.String("group", group), zap.Strings("ids", ids), zap.Error(err))
		return store.Error(err)
	}
	return nil
}

func (l *laneStore) DeleteLaneGroupReleases(tx store.Tx, rule *rules.LaneGroupRelease) error {
	return l.repo().DeleteRelease(tx, governanceRuleTypeLaneGroup, rule.Id)
}

// cleanSoftDeletedRules .
func (l *laneStore) cleanSoftDeletedRules() error {
	return nil
}

func (l *laneStore) repo() *governanceRuleRepository {
	if l.governanceRuleRepository == nil {
		l.governanceRuleRepository = newGovernanceRuleRepository(l.master, l.slave)
	}
	return l.governanceRuleRepository
}

func (l *laneStore) updateLaneRuleAggregate(
	tx store.Tx, namespace, groupName string, update func(group *ruletypes.LaneGroup) error,
) error {
	group, err := l.LockLaneGroup(tx, namespace, groupName)
	if err != nil {
		return err
	}
	if group == nil {
		return sql.ErrNoRows
	}
	if group.LaneRules == nil {
		group.LaneRules = map[string]*ruletypes.LaneRule{}
	}
	if err := update(group); err != nil {
		return err
	}
	group.ModifyTime = time.Now()
	return l.UpdateLaneGroup(tx, group)
}

func laneGroupToGovernanceRuleRecord(group *ruletypes.LaneGroup) *governanceRuleRecord {
	return &governanceRuleRecord{
		ID:          group.ID,
		RuleType:    governanceRuleTypeLaneGroup,
		Namespace:   group.Namespace,
		Name:        group.Name,
		Rule:        marshalLaneGroupAggregate(group),
		Revision:    group.Revision,
		Description: group.Description,
		Metadata:    marshalMetadata(group.Metadata),
		Valid:       group.Valid,
	}
}

func governanceRuleRecordToLaneGroup(record *governanceRuleRecord) (*ruletypes.LaneGroup, error) {
	if record == nil {
		return nil, nil
	}
	group := &ruletypes.LaneGroup{}
	if err := unmarshalLaneGroupAggregate(record.Rule, group); err != nil {
		return nil, err
	}
	group.ID = record.ID
	group.Namespace = record.Namespace
	for _, laneRule := range group.LaneRules {
		laneRule.Namespace = record.Namespace
	}
	group.Name = record.Name
	group.Revision = record.Revision
	group.Description = record.Description
	group.Valid = record.Valid
	group.CreateTime = record.CreateTime
	group.ModifyTime = record.ModifyTime
	return group, nil
}

func laneGroupReleaseToGovernanceReleaseRecord(release *ruletypes.LaneGroupRelease) *governanceRuleReleaseRecord {
	record := &governanceRuleReleaseRecord{
		ID:           release.Id,
		RuleType:     governanceRuleTypeLaneGroup,
		Namespace:    release.Namespace,
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
		record.Namespace = utilsDefaultString(record.Namespace, release.Rule.Namespace)
		record.Rule = marshalLaneGroupAggregate(release.Rule.LaneGroup)
	}
	return record
}

func governanceRuleReleaseRecordToLaneGroupRelease(record *governanceRuleReleaseRecord) (*ruletypes.LaneGroupRelease, error) {
	if record == nil {
		return nil, nil
	}
	group := &ruletypes.LaneGroup{}
	if err := unmarshalLaneGroupAggregate(record.Rule, group); err != nil {
		return nil, err
	}
	group.Namespace = record.Namespace
	proto, err := group.ToProto()
	if err != nil {
		return nil, err
	}
	return &ruletypes.LaneGroupRelease{
		RuleRelease: ruletypes.RuleRelease{
			Id:           record.ID,
			Namespace:    record.Namespace,
			ReleaseName:  record.ReleaseName,
			RuleId:       record.RuleID,
			RuleName:     record.RuleName,
			Description:  record.Description,
			ReleaseType:  rules.ReleaseType(record.ReleaseType),
			Active:       record.Active,
			Version:      record.Version,
			Valid:        record.Valid,
			Ctime:        record.CreateTime,
			Mtime:        record.ModifyTime,
			ClientLabels: unmarshalClientLabels(record.ClientLabels),
		},
		Rule: proto,
	}, nil
}

func marshalLaneGroupAggregate(group *ruletypes.LaneGroup) string {
	if group == nil {
		return "{}"
	}
	proto, err := group.ToProto()
	if err != nil {
		return group.Rule
	}
	data, err := json.Marshal(proto.Proto)
	if err != nil {
		return group.Rule
	}
	return string(data)
}

func unmarshalLaneGroupAggregate(raw string, group *ruletypes.LaneGroup) error {
	spec := &apitraffic.LaneGroup{}
	if err := json.Unmarshal([]byte(raw), spec); err != nil {
		return err
	}
	return group.FromSpec(spec)
}

func marshalClientLabels(labels []*model.ClientLabel) string {
	if len(labels) == 0 {
		return "[]"
	}
	data, err := json.Marshal(labels)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func unmarshalClientLabels(raw string) []*model.ClientLabel {
	if raw == "" {
		return nil
	}
	labels := []*model.ClientLabel{}
	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		return nil
	}
	return labels
}

func utilsDefaultString(v string, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
