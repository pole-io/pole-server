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
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestLosslessStoreCreateUsesGovernanceRuleRepository(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &losslessStore{master: db, slave: db, governanceRuleRepository: newGovernanceRuleRepository(db, db)}
	rule := &rules.LosslessRule{
		ID:          "lossless-1",
		Namespace:   "default",
		Service:     "svc-a",
		Revision:    "rev-1",
		Description: "lossless rule",
		Valid:       true,
		Proto: &apitraffic.LosslessRule{
			Id: "lossless-1",
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			rule.ID,
			string(governanceRuleTypeLossless),
			rule.Namespace,
			rule.Service,
			"",
			rule.Service,
			"",
			0,
			1,
			0,
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			rule.Revision,
			rule.Description,
			"{}",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, store.CreateLossLessRule(rule))
	require.NoError(t, mock.ExpectationsWereMet())
}
