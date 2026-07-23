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
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
)

func TestRouterRuleStoreCreateUsesGovernanceRuleRepository(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &routerRuleStore{master: db, slave: db, governanceRuleRepository: newGovernanceRuleRepository(db, db)}
	conf := &rules.RouterConfig{
		ID:          "route-1",
		Namespace:   "default",
		Name:        "route-a",
		Policy:      "RulePolicy",
		Config:      `{"rules":[]}`,
		Enable:      true,
		Priority:    10,
		Revision:    "rev-1",
		Description: "route rule",
		Valid:       true,
	}

	mock.ExpectBegin()
	expectGovernanceRuleNameAvailable(mock, governanceRuleTypeRoute, conf.Namespace, conf.Name)
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			conf.ID,
			string(governanceRuleTypeRoute),
			conf.Namespace,
			conf.Name,
			"",
			"",
			"",
			int(conf.Priority),
			1,
			0,
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			conf.Policy,
			conf.Config,
			routerConfigRuleJSONMatcher(conf),
			conf.Revision,
			conf.Description,
			"{}",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, store.CreateRoutingConfig(conf))
	require.NoError(t, mock.ExpectationsWereMet())
}

func routerConfigRuleJSONMatcher(conf *rules.RouterConfig) sqlmock.Argument {
	return jsonContainsFields(map[string]interface{}{
		"id":          conf.ID,
		"name":        conf.Name,
		"namespace":   conf.Namespace,
		"policy":      conf.Policy,
		"config":      conf.Config,
		"enable":      conf.Enable,
		"priority":    float64(conf.Priority),
		"revision":    conf.Revision,
		"description": conf.Description,
	})
}
