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
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

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

func TestGovernanceRuleRepositoryQueryReleaseVersionsByRuleName(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repo := newGovernanceRuleRepository(db, db)
	filter := map[string]string{"rule_name": "rate-limit-a"}
	countSQL := "SELECT COUNT(*) FROM governance_rule_release WHERE rule_type = ? AND rule_name = ? AND flag = 0"
	querySQL := "SELECT " + ReleaseVersionColumns + " FROM governance_rule_release WHERE rule_type = ? AND rule_name = ? AND flag = 0 ORDER BY version DESC LIMIT ?, ?"

	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).
		WithArgs(string(governanceRuleTypeRateLimit), filter["rule_name"]).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(uint64(1)))
	mock.ExpectQuery(regexp.QuoteMeta(querySQL)).
		WithArgs(string(governanceRuleTypeRateLimit), filter["rule_name"], uint32(0), uint32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "rule_id", "rule_name", "flag", "active", "version", "description", "release_type", "ctime", "mtime",
		}).AddRow("release-1", "normal", "rate-1", "rate-limit-a", 0, 1, uint64(3), "desc", "normal", int64(100), int64(101)))

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

	mock.ExpectQuery(regexp.QuoteMeta(selectGovernanceRuleReleaseSQL)).
		WithArgs(string(governanceRuleTypeTrafficSecurity), "", "rule-1", "release-1", "normal").
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
