/*
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
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

func TestConfigTemplateTables(t *testing.T) {
	tables := configTemplateTables()
	require.Len(t, tables, 5)

	seen := make(map[string]bool, len(tables))
	for _, table := range tables {
		seen[table.name] = true
		require.Contains(t, table.ddl, table.name)
	}
	require.True(t, seen["config_template_release"])
	require.True(t, seen["namespace_template_values"])
	require.True(t, seen["namespace_template_value_release"])
	require.True(t, seen["config_file_template_binding_release"])
	require.True(t, seen["namespace_config_template_draft"])
}

func TestEnsureConfigTemplateSchemaCreatesMissingTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	for _, table := range configTemplateTables() {
		mock.ExpectQuery(regexp.QuoteMeta(`
		select count(*)
		from information_schema.tables
		where table_schema = database()
		  and table_name = ?`)).
			WithArgs(table.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(regexp.QuoteMeta(table.ddl)).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	expectMissingConfigTemplateDraftColumns(mock)

	require.NoError(t, ensureConfigTemplateSchema(db))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureConfigTemplateSchemaSkipsExistingTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	for _, table := range configTemplateTables() {
		mock.ExpectQuery(regexp.QuoteMeta(`
		select count(*)
		from information_schema.tables
		where table_schema = database()
		  and table_name = ?`)).
			WithArgs(table.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	expectExistingConfigTemplateDraftColumns(mock)

	require.NoError(t, ensureConfigTemplateSchema(db))
	require.NoError(t, mock.ExpectationsWereMet())
}

func expectMissingConfigTemplateDraftColumns(mock sqlmock.Sqlmock) {
	for _, column := range configTemplateDraftColumns() {
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(column.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(regexp.QuoteMeta(column.ddl)).WillReturnResult(sqlmock.NewResult(0, 1))
	}
}

func expectExistingConfigTemplateDraftColumns(mock sqlmock.Sqlmock) {
	for _, column := range configTemplateDraftColumns() {
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(column.name).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
}
