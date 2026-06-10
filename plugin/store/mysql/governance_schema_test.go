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
)

func TestGovernanceRuleTablesUseUnifiedStorage(t *testing.T) {
	tables := governanceRuleTables()
	require.Len(t, tables, 2)

	seen := make(map[string]bool, len(tables))
	for _, table := range tables {
		seen[table.name] = true
		require.Contains(t, table.ddl, table.name)
		require.Contains(t, table.ddl, "rule_type")
	}

	require.True(t, seen["governance_rule"])
	require.True(t, seen["governance_rule_release"])
}

func TestEnsureGovernanceRuleSchemaCreatesMissingUnifiedTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}

	for _, table := range governanceRuleTables() {
		mock.ExpectQuery(regexp.QuoteMeta(`
		select count(*)
		from information_schema.tables
		where table_schema = database()
		  and table_name = ?`)).
			WithArgs(table.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(regexp.QuoteMeta(table.ddl)).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	require.NoError(t, ensureGovernanceRuleSchema(db))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureGovernanceRuleSchemaSkipsExistingUnifiedTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}

	for _, table := range governanceRuleTables() {
		mock.ExpectQuery(regexp.QuoteMeta(`
		select count(*)
		from information_schema.tables
		where table_schema = database()
		  and table_name = ?`)).
			WithArgs(table.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}

	require.NoError(t, ensureGovernanceRuleSchema(db))
	require.NoError(t, mock.ExpectationsWereMet())
}
