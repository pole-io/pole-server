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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestLaneGroupGovernanceRecordKeepsLaneRulesInAggregateJSON(t *testing.T) {
	group := &ruletypes.LaneGroup{}
	require.NoError(t, group.FromSpec(&apitraffic.LaneGroup{
		Id:          "group-1",
		Name:        "lane-group-a",
		Description: "lane group",
		Rules: []*apitraffic.LaneRule{
			{Id: "rule-1", Name: "lane-a", GroupName: "lane-group-a", Enable: true, Priority: 10},
			{Id: "rule-2", Name: "lane-b", GroupName: "lane-group-a", Enable: false, Priority: 20},
		},
	}))
	group.Revision = "rev-1"
	group.Valid = true
	group.CreateTime = time.Unix(100, 0)
	group.ModifyTime = time.Unix(200, 0)

	record := laneGroupToGovernanceRuleRecord(group)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(record.Rule), &raw))
	require.Len(t, raw["rules"], 2)

	out, err := governanceRuleRecordToLaneGroup(&governanceRuleRecord{
		ID:          record.ID,
		RuleType:    record.RuleType,
		Name:        record.Name,
		Rule:        record.Rule,
		Revision:    record.Revision,
		Description: record.Description,
		Valid:       true,
		CreateTime:  group.CreateTime,
		ModifyTime:  group.ModifyTime,
	})
	require.NoError(t, err)
	require.Equal(t, "group-1", out.ID)
	require.Len(t, out.LaneRules, 2)
	require.Equal(t, uint32(10), out.LaneRules["rule-1"].Priority)
	require.False(t, out.LaneRules["rule-2"].Enable)
}
