/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, Tencent. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package paramcheck

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestCheckLaneGroupParamRejectsMoreThanTwentyLaneRules(t *testing.T) {
	group := &apitraffic.LaneGroup{
		Name:  "lane-group-a",
		Rules: make([]*apitraffic.LaneRule, 0, 21),
	}
	for i := 0; i < 21; i++ {
		group.Rules = append(group.Rules, &apitraffic.LaneRule{
			Name: fmt.Sprintf("lane-%d", i),
		})
	}

	resp := checkLaneGroupParam(group, false)

	require.NotNil(t, resp)
	require.Equal(t, uint32(apimodel.Code_InvalidParameter), resp.Code)
	require.Contains(t, resp.Info, "lane_rule size must be <= 20")
}
