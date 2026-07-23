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

func TestFaultDetectStoreCreateUsesGovernanceRuleRepository(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	store := &faultDetectRuleStore{master: db, slave: db, governanceRuleRepository: newGovernanceRuleRepository(db, db)}
	rule := &rules.FaultDetectRule{
		ID:           "fault-1",
		Name:         "fault-a",
		Namespace:    "default",
		Description:  "desc",
		DstService:   "dst",
		DstNamespace: "default",
		DstMethod:    "GET",
		Rule:         `{"id":"fault-1","name":"fault-a"}`,
		Revision:     "rev-1",
		Metadata:     map[string]string{"env": "test"},
		Valid:        true,
	}

	mock.ExpectBegin()
	expectGovernanceRuleNameAvailable(mock, governanceRuleTypeFaultDetect, rule.Namespace, rule.Name)
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			rule.ID,
			string(governanceRuleTypeFaultDetect),
			rule.Namespace,
			rule.Name,
			"",
			"",
			"",
			0,
			1,
			0,
			"",
			"",
			"",
			rule.DstService,
			rule.DstNamespace,
			rule.DstMethod,
			"",
			"",
			rule.Rule,
			rule.Rule,
			rule.Revision,
			rule.Description,
			`{"env":"test"}`,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, store.CreateFaultDetectRule(rule))
	require.NoError(t, mock.ExpectationsWereMet())
}
