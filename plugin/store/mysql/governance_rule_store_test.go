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
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func expectGovernanceRuleNameAvailable(
	mock sqlmock.Sqlmock, ruleType governanceRuleType, namespace, name string,
) {
	mock.ExpectQuery(regexp.QuoteMeta(lockGovernanceRuleSQL)).
		WithArgs(string(ruleType), "", namespace, name).
		WillReturnRows(sqlmock.NewRows(governanceRuleColumnsForTest()))
}

func TestGovernanceRuleRepositoryCreateRuleUsesRuleType(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	record := &governanceRuleRecord{
		ID:          "rule-1",
		RuleType:    governanceRuleTypeLaneGroup,
		Namespace:   "default",
		Name:        "lane-a",
		Service:     "svc-a",
		Rule:        `{"name":"lane-a"}`,
		Revision:    "rev-1",
		Description: "lane group",
		Valid:       true,
	}

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)

	mock.ExpectQuery(regexp.QuoteMeta(lockGovernanceRuleSQL)).
		WithArgs(string(record.RuleType), "", record.Namespace, record.Name).
		WillReturnRows(sqlmock.NewRows(governanceRuleColumnsForTest()))
	mock.ExpectExec(regexp.QuoteMeta(insertGovernanceRuleSQL)).
		WithArgs(
			record.ID,
			string(record.RuleType),
			record.Namespace,
			record.Name,
			record.ServiceID,
			record.Service,
			record.Method,
			record.Priority,
			record.Enable,
			record.Disable,
			record.Level,
			record.SrcService,
			record.SrcNamespace,
			record.DstService,
			record.DstNamespace,
			record.DstMethod,
			record.Labels,
			record.Policy,
			record.Config,
			record.Rule,
			record.Revision,
			record.Description,
			record.Metadata,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.CreateRule(storeTx, record))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryCreateRuleRejectsDuplicateWithinOwnerNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	record := &governanceRuleRecord{ID: "route-2", RuleType: governanceRuleTypeRoute, Namespace: "prod", Name: "checkout-route"}
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)
	mock.ExpectQuery(regexp.QuoteMeta(lockGovernanceRuleSQL)).
		WithArgs(string(record.RuleType), "", record.Namespace, record.Name).
		WillReturnRows(sqlmock.NewRows(governanceRuleColumnsForTest()).
			AddRow("route-1", string(record.RuleType), "prod", "checkout-route", "", "", "", 0, 1, 0, "", "", "", "", "", "", "", "", "", `{}`, "rev-1", "", "{}", 0, int64(100), int64(100), int64(100)))

	err = repo.CreateRule(storeTx, record)
	require.Equal(t, store.DuplicateEntryErr, store.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryActiveReleaseScopesByRuleTypeRuleAndReleaseType(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	release := &governanceRuleReleaseRecord{
		ID:          "rel-1",
		RuleType:    governanceRuleTypeLaneGroup,
		ReleaseName: "normal",
		RuleID:      "rule-1",
		RuleName:    "lane-a",
		ReleaseType: "normal",
	}

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)

	mock.ExpectExec(regexp.QuoteMeta(inactiveGovernanceRuleReleaseSQL)).
		WithArgs(string(release.RuleType), release.RuleID, release.ReleaseType).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(selectMaxGovernanceRuleReleaseVersionSQL)).
		WithArgs(string(release.RuleType), release.RuleID).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(uint64(3)))
	mock.ExpectExec(regexp.QuoteMeta(activeGovernanceRuleReleaseSQL)).
		WithArgs(uint64(4), string(release.RuleType), release.RuleID, release.ReleaseName, release.ReleaseType).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.ActiveRelease(storeTx, release))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryGetMoreRulesReturnsMultipleRuleTypes(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	mtime := time.Unix(100, 0)

	rows := sqlmock.NewRows(governanceRuleColumnsForTest()).
		AddRow("route-1", string(governanceRuleTypeRoute), "default", "route-a", "", "svc-a", "", 0, 1, 0, "", "", "", "", "", "", "", "nearby", "", `{"id":"route-1"}`, "rev-1", "route", "", 0, int64(90), int64(101), int64(101)).
		AddRow("lane-1", string(governanceRuleTypeLaneGroup), "", "lane-a", "", "", "", 0, 1, 0, "", "", "", "", "", "", "", "", "", `{"id":"lane-1"}`, "rev-2", "lane", "", 0, int64(91), int64(102), int64(102))

	mock.ExpectQuery(regexp.QuoteMeta(selectMoreGovernanceRulesSQL)).
		WithArgs(timeToTimestamp(mtime)).
		WillReturnRows(rows)

	out, err := repo.GetMoreRules(mtime, false)
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, governanceRuleTypeRoute, out[0].RuleType)
	require.Equal(t, governanceRuleTypeLaneGroup, out[1].RuleType)
	require.Equal(t, time.Unix(102, 0), out[1].ModifyTime)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryGetRuleByNameScopesOwnerNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	rows := sqlmock.NewRows(governanceRuleColumnsForTest()).
		AddRow("route-prod", string(governanceRuleTypeRoute), "prod", "checkout-route", "", "", "", 0, 1, 0, "", "", "", "", "", "", "", "", "", `{}`, "rev-1", "", "{}", 0, int64(100), int64(100), int64(100))

	mock.ExpectQuery(regexp.QuoteMeta(selectGovernanceRuleByNameSQL)).
		WithArgs(string(governanceRuleTypeRoute), "prod", "checkout-route").
		WillReturnRows(rows)

	rule, err := repo.GetRuleByName(governanceRuleTypeRoute, "prod", "checkout-route")
	require.NoError(t, err)
	require.Equal(t, "route-prod", rule.ID)
	require.Equal(t, "prod", rule.Namespace)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryLockRuleScopesNameByOwnerNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)
	rows := sqlmock.NewRows(governanceRuleColumnsForTest()).
		AddRow("route-prod", string(governanceRuleTypeRoute), "prod", "checkout-route", "", "", "", 0, 1, 0, "", "", "", "", "", "", "", "", "", `{}`, "rev-1", "", "{}", 0, int64(100), int64(100), int64(100))

	mock.ExpectQuery(regexp.QuoteMeta(lockGovernanceRuleSQL)).
		WithArgs(string(governanceRuleTypeRoute), "route-prod", "prod", "checkout-route").
		WillReturnRows(rows)

	rule, err := repo.LockRule(storeTx, governanceRuleTypeRoute, "route-prod", "prod", "checkout-route")
	require.NoError(t, err)
	require.Equal(t, "route-prod", rule.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryQueryRulesFiltersOwnerNamespaceExactly(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	countSQL := "SELECT COUNT(*) FROM governance_rule WHERE rule_type = ? AND flag = 0 AND namespace = ?"
	querySQL := "SELECT " + governanceRuleSelectColumns + " FROM governance_rule WHERE rule_type = ? AND flag = 0 AND namespace = ? ORDER BY mtime DESC LIMIT ?, ?"
	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).
		WithArgs(string(governanceRuleTypeLaneGroup), "prod").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(uint32(1)))
	mock.ExpectQuery(regexp.QuoteMeta(querySQL)).
		WithArgs(string(governanceRuleTypeLaneGroup), "prod", uint32(0), uint32(20)).
		WillReturnRows(sqlmock.NewRows(governanceRuleColumnsForTest()).
			AddRow("lane-prod", string(governanceRuleTypeLaneGroup), "prod", "gray-lane", "", "", "", 0, 1, 0, "", "", "", "", "", "", "", "", "", `{}`, "rev-1", "", "{}", 0, int64(100), int64(100), int64(100)))

	total, rules, err := repo.QueryRules(context.Background(), governanceRuleTypeLaneGroup, map[string]string{"namespace": "prod"}, 0, 20)
	require.NoError(t, err)
	require.Equal(t, uint32(1), total)
	require.Len(t, rules, 1)
	require.Equal(t, "prod", rules[0].Namespace)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryCountRulesByOwnerNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	mock.ExpectQuery(regexp.QuoteMeta(countGovernanceRulesByNamespaceSQL)).
		WithArgs("prod").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(uint64(4)))

	total, err := repo.CountRulesByNamespace("prod")
	require.NoError(t, err)
	require.Equal(t, uint64(4), total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryQueryReleaseVersionsByRuleName(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	filter := map[string]string{"namespace": "prod", "rule_name": "rate-limit-a"}
	countSQL := "SELECT COUNT(*) FROM governance_rule_release WHERE rule_type = ? AND namespace = ? AND rule_name = ? AND flag = 0"
	querySQL := "SELECT " + ReleaseVersionColumns + " FROM governance_rule_release WHERE rule_type = ? AND namespace = ? AND rule_name = ? AND flag = 0 ORDER BY version DESC LIMIT ?, ?"

	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).
		WithArgs(string(governanceRuleTypeRateLimit), filter["namespace"], filter["rule_name"]).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(uint64(1)))
	mock.ExpectQuery(regexp.QuoteMeta(querySQL)).
		WithArgs(string(governanceRuleTypeRateLimit), filter["namespace"], filter["rule_name"], uint32(0), uint32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "namespace", "name", "rule_id", "rule_name", "flag", "active", "version", "description", "release_type", "ctime", "mtime",
		}).AddRow("release-1", "prod", "normal", "rate-1", "rate-limit-a", 0, 1, uint64(3), "desc", "normal", int64(100), int64(101)))

	total, releases, err := repo.QueryReleaseVersions(
		context.Background(),
		governanceRuleTypeRateLimit,
		apimodel.RuleRelease_RateLimitRules,
		filter,
		0,
		10,
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), total)
	require.Len(t, releases, 1)
	require.Equal(t, "rate-limit-a", releases[0].RuleName)
	require.Equal(t, "prod", releases[0].Namespace)
	require.Equal(t, apimodel.RuleRelease_RateLimitRules, releases[0].Resource)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryGetReleaseReturnsNilWhenMissing(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)

	mock.ExpectQuery(regexp.QuoteMeta(selectGovernanceRuleReleaseByRuleIDSQL)).
		WithArgs(string(governanceRuleTypeTrafficSecurity), "rule-1", "release-1", "normal").
		WillReturnRows(sqlmock.NewRows(governanceRuleReleaseColumnsForTest()))

	got, err := repo.GetRelease(storeTx, governanceRuleTypeTrafficSecurity, &rules.RuleRelease{
		RuleId:      "rule-1",
		ReleaseName: "release-1",
		ReleaseType: "normal",
	})
	require.NoError(t, err)
	require.Nil(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryGetReleaseByNameScopesOwnerNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)
	mock.ExpectQuery(regexp.QuoteMeta(selectGovernanceRuleReleaseSQL)).
		WithArgs(string(governanceRuleTypeRoute), "prod", "checkout-route", "normal", "normal").
		WillReturnRows(sqlmock.NewRows(governanceRuleReleaseColumnsForTest()))

	got, err := repo.GetRelease(storeTx, governanceRuleTypeRoute, &rules.RuleRelease{
		Namespace:   "prod",
		RuleName:    "checkout-route",
		ReleaseName: "normal",
		ReleaseType: "normal",
	})
	require.NoError(t, err)
	require.Nil(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGovernanceRuleRepositoryGetActiveReleaseReturnsNilWhenMissing(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	storeTx := NewSqlDBTx(tx)

	expectedSQL := `SELECT id, rule_type, name, rule_id, rule_name, namespace, service, rule,
	version, active, description, release_type, client_labels, metadata, flag,
	UNIX_TIMESTAMP\(ctime\), UNIX_TIMESTAMP\(mtime\)
FROM governance_rule_release
WHERE rule_type = \? AND rule_id = \? AND release_type = \? AND active = 1 AND flag = 0
ORDER BY version DESC
LIMIT 1`
	mock.ExpectQuery(expectedSQL).
		WithArgs(string(governanceRuleTypeTrafficSecurity), "rule-1", "gray").
		WillReturnRows(sqlmock.NewRows(governanceRuleReleaseColumnsForTest()))

	got, err := repo.GetActiveRelease(storeTx, &governanceRuleReleaseRecord{
		RuleType:    governanceRuleTypeTrafficSecurity,
		RuleID:      "rule-1",
		ReleaseType: "gray",
	})
	require.NoError(t, err)
	require.Nil(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func governanceRuleColumnsForTest() []string {
	return []string{
		"id", "rule_type", "namespace", "name", "service_id", "service", "method", "priority",
		"enable", "disable", "level", "src_service", "src_namespace", "dst_service", "dst_namespace",
		"dst_method", "labels", "policy", "config", "rule", "revision", "description", "metadata",
		"flag", "ctime", "etime", "mtime",
	}
}

func governanceRuleReleaseColumnsForTest() []string {
	return []string{
		"id", "rule_type", "name", "rule_id", "rule_name", "namespace", "service", "rule",
		"version", "active", "description", "release_type", "client_labels", "metadata", "flag",
		"ctime", "mtime",
	}
}
