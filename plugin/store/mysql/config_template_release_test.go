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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

func TestCreateActiveNormalNamespaceTemplateValueReleaseDeactivatesPreviousRelease(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &namespaceTemplateValuesStore{master: db, slave: db}
	release := &conftypes.NamespaceTemplateValueRelease{
		ID:                "value-release-2",
		ValuesID:          "values-1",
		Namespace:         "dev",
		TemplateID:        7,
		TemplateReleaseID: "template-release-1",
		Values:            `{"database.host":"mysql-dev"}`,
		ReleaseType:       conftypes.TemplateValueReleaseTypeNormal,
		BetaLabels:        []*apimodel.ClientLabel{},
		Active:            true,
		Version:           2,
		Revision:          "value-revision-2",
		CreateBy:          "tester",
		ModifyBy:          "tester",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE namespace_template_value_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND template_id = ? AND release_type = ? AND active = 1`)).
		WithArgs("dev", uint64(7), conftypes.TemplateValueReleaseTypeNormal).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO namespace_template_value_release").
		WithArgs("value-release-2", "values-1", "dev", uint64(7), "template-release-1",
			`{"database.host":"mysql-dev"}`, conftypes.TemplateValueReleaseTypeNormal,
			"[]", int32(0), true, uint64(2), "value-revision-2", "", "tester", "tester").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.CreateNamespaceTemplateValueRelease(release))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConfigTemplateEnvironmentReleaseUsesOneTransaction(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &stableStore{
		configTemplateReleaseStore:   &configTemplateReleaseStore{master: db, slave: db},
		namespaceTemplateValuesStore: &namespaceTemplateValuesStore{master: db, slave: db},
	}
	template := &conftypes.ConfigTemplateRelease{
		ID: "template-snapshot-2", TemplateID: 7, Name: "application", Content: "plain",
		Format: "text", Engine: "pole-mustache", EngineVersion: "v1", Version: 2,
		ContentSHA256: "sha", CreateBy: "tester",
	}
	release := &conftypes.NamespaceTemplateValueRelease{
		ID: "environment-release-2", ValuesID: "prod@7", Namespace: "prod", TemplateID: 7,
		TemplateReleaseID: template.ID, Values: `{}`, ReleaseType: conftypes.TemplateValueReleaseTypeNormal,
		BetaLabels: []*apimodel.ClientLabel{}, Active: true, Version: 2, Revision: "revision-2",
		CreateBy: "tester", ModifyBy: "tester",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO config_template_release").
		WithArgs(template.ID, uint64(7), template.Name, template.Content, template.Format,
			template.ParameterSchema, template.Engine, template.EngineVersion, template.Version,
			template.ContentSHA256, template.Comment, template.CreateBy).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE namespace_template_value_release").
		WithArgs("prod", uint64(7), conftypes.TemplateValueReleaseTypeNormal).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO namespace_template_value_release").
		WithArgs(release.ID, release.ValuesID, release.Namespace, release.TemplateID,
			release.TemplateReleaseID, release.Values, release.ReleaseType, "[]", int32(0), true,
			uint64(2), release.Revision, "", "tester", "tester").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.CreateConfigTemplateEnvironmentRelease(template, release))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNamespaceTemplateValueReleaseRestoresBetaLabels(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &namespaceTemplateValuesStore{master: db, slave: db}
	mock.ExpectQuery(regexp.QuoteMeta(namespaceTemplateValueReleaseSelect + ` WHERE id = ?`)).
		WithArgs("gray-release").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "values_id", "namespace", "template_id", "template_release_id",
			"values_content", "release_type", "beta_labels", "priority", "active",
			"version", "revision", "comment", "create_by", "modify_by", "ctime", "mtime",
		}).AddRow("gray-release", "values-1", "prod", 9, "template-release-3",
			`{"feature.enabled":true}`, "gray",
			`[{"key":"region","value":{"type":0,"value":"ap-shanghai"}}]`, 100,
			true, 3, "gray-revision", "gray rollout", "creator", "modifier", 1000, 1001))

	release, err := repository.GetNamespaceTemplateValueRelease("gray-release")
	require.NoError(t, err)
	require.Len(t, release.BetaLabels, 1)
	require.Equal(t, "region", release.BetaLabels[0].GetKey())
	require.Equal(t, "ap-shanghai", release.BetaLabels[0].GetValue().GetValue())
	require.Equal(t, apimodel.MatchString_EXACT, release.BetaLabels[0].GetValue().GetType())
	require.Equal(t, conftypes.TemplateValueReleaseTypeGray, release.ReleaseType)
	require.Equal(t, int32(100), release.Priority)
	require.Equal(t, "gray-revision", release.Revision)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMoreNamespaceTemplateValueReleasesReturnsActiveAndDeactivatedRows(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &namespaceTemplateValuesStore{master: db, slave: db}
	query := namespaceTemplateValueReleaseSelect +
		` WHERE mtime > FROM_UNIXTIME(?) ORDER BY mtime ASC`
	rows := sqlmock.NewRows([]string{
		"id", "values_id", "namespace", "template_id", "template_release_id",
		"values_content", "release_type", "beta_labels", "priority", "active",
		"version", "revision", "comment", "create_by", "modify_by", "ctime", "mtime",
	}).
		AddRow("normal-old", "values-1", "prod", 7, "template-release-1",
			`{"region":"old"}`, "normal", "[]", 0, false, 1, "old", "", "", "", 100, 200).
		AddRow("normal-new", "values-1", "prod", 7, "template-release-1",
			`{"region":"new"}`, "normal", "[]", 0, true, 2, "new", "", "", "", 201, 201)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(150)).WillReturnRows(rows)

	releases, err := repository.GetMoreNamespaceTemplateValueReleases(false, time.Unix(150, 0))

	require.NoError(t, err)
	require.Len(t, releases, 2)
	require.False(t, releases[0].Active)
	require.True(t, releases[1].Active)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListNamespaceTemplateValueReleasesUsesGrayMatchOrder(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &namespaceTemplateValuesStore{master: db, slave: db}
	mock.ExpectQuery(regexp.QuoteMeta(namespaceTemplateValueReleaseSelect+
		` WHERE namespace = ? AND template_id = ?
		ORDER BY priority ASC, version DESC, mtime DESC`)).
		WithArgs("prod", uint64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "values_id", "namespace", "template_id", "template_release_id",
			"values_content", "release_type", "beta_labels", "priority", "active",
			"version", "revision", "comment", "create_by", "modify_by", "ctime", "mtime",
		}))

	releases, err := repository.ListNamespaceTemplateValueReleases("prod", 9)
	require.NoError(t, err)
	require.Empty(t, releases)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateActiveConfigTemplateBindingDeactivatesPreviousBinding(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	db := &BaseDB{DB: rawDB}
	repository := &configTemplateBindingStore{master: db, slave: db}
	binding := &conftypes.ConfigTemplateBinding{
		BindingReleaseID:  "binding-2",
		Namespace:         "prod",
		Group:             "application",
		FileName:          "server.yaml",
		TemplateID:        7,
		TemplateReleaseID: "template-release-2",
		Active:            true,
		Version:           2,
		CreateBy:          "tester",
		ModifyBy:          "tester",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE config_file_template_binding_release
			SET active = 0, mtime = sysdate()
			WHERE namespace = ? AND config_group = ? AND file_name = ? AND active = 1`)).
		WithArgs("prod", "application", "server.yaml").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO config_file_template_binding_release").
		WithArgs("binding-2", "prod", "application", "server.yaml", uint64(7),
			"template-release-2", true, uint64(2), "", "tester", "tester").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repository.CreateConfigTemplateBinding(binding))
	require.NoError(t, mock.ExpectationsWereMet())
}
