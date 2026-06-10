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

func TestRateLimitStoreCreateUsesGovernanceRuleRepository(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &rateLimitStore{master: db, slave: db, governanceRuleRepository: newGovernanceRuleRepository(db, db)}
	limit := &rules.RateLimit{
		ID:          "limit-1",
		Name:        "limit-a",
		ServiceID:   "svc-id-a",
		Method:      "GET",
		Labels:      `{"$method":{"type":"EXACT","value":"GET"}}`,
		Priority:    100,
		Rule:        `{"id":"limit-1","name":"limit-a"}`,
		Revision:    "rev-1",
		Disable:     false,
		Valid:       true,
		Metadata:    map[string]string{"env": "test"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			limit.ID,
			string(governanceRuleTypeRateLimit),
			"",
			limit.Name,
			limit.ServiceID,
			"",
			limit.Method,
			int(limit.Priority),
			1,
			0,
			"",
			"",
			"",
			"",
			"",
			"",
			limit.Labels,
			"",
			"",
			limit.Rule,
			limit.Revision,
			"",
			`{"env":"test"}`,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, store.CreateRateLimit(limit))
	require.NoError(t, mock.ExpectationsWereMet())
}
