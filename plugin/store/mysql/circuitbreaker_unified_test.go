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

func TestCircuitBreakerStoreCreateUsesGovernanceRuleRepository(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &circuitBreakerStore{master: db, slave: db, governanceRuleRepository: newGovernanceRuleRepository(db, db)}
	rule := &rules.CircuitBreakerRule{
		ID:           "circuit-1",
		Name:         "circuit-a",
		Namespace:    "default",
		Description:  "desc",
		Level:        1,
		SrcService:   "src",
		SrcNamespace: "default",
		DstService:   "dst",
		DstNamespace: "default",
		DstMethod:    "GET",
		Rule:         `{"id":"circuit-1","name":"circuit-a"}`,
		Revision:     "rev-1",
		Enable:       true,
		Valid:        true,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			rule.ID,
			string(governanceRuleTypeCircuitBreaker),
			rule.Namespace,
			rule.Name,
			"",
			"",
			"",
			0,
			1,
			0,
			"1",
			rule.SrcService,
			rule.SrcNamespace,
			rule.DstService,
			rule.DstNamespace,
			rule.DstMethod,
			"",
			"",
			rule.Rule,
			rule.Rule,
			rule.Revision,
			rule.Description,
			"{}",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, store.CreateCircuitBreakerRule(rule))
	require.NoError(t, mock.ExpectationsWereMet())
}
