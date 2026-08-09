package config

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storeapi "github.com/pole-io/pole-server/apis/store"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestSelectTemplateValueReleaseUsesFirstMatchedGrayThenNormal(t *testing.T) {
	releases := []*conftypes.NamespaceTemplateValueRelease{
		{
			ID: "gray-priority-1", TemplateReleaseID: "tpl-r1",
			ReleaseType: conftypes.TemplateValueReleaseTypeGray, Active: true,
			BetaLabels: []*apimodel.ClientLabel{clientLabel("region", "shanghai")},
		},
		{
			ID: "gray-priority-2", TemplateReleaseID: "tpl-r1",
			ReleaseType: conftypes.TemplateValueReleaseTypeGray, Active: true,
			BetaLabels: []*apimodel.ClientLabel{clientLabel("region", "shanghai")},
		},
		{
			ID: "normal", TemplateReleaseID: "tpl-r1",
			ReleaseType: conftypes.TemplateValueReleaseTypeNormal, Active: true,
		},
	}

	matched := selectTemplateValueRelease(releases, map[string]string{"region": "shanghai"})
	require.Equal(t, "gray-priority-1", matched.ID)

	fallback := selectTemplateValueRelease(releases, map[string]string{"region": "beijing"})
	require.Equal(t, "normal", fallback.ID)
}

func TestUpdateConfigFileAttributeDoesNotReportUnchangedContent(t *testing.T) {
	server := &Server{}
	saved := &conftypes.ConfigFile{
		Content: "same", Comment: "different comment", Format: "yaml",
		Metadata: map[string]string{"owner": "team-a"},
	}
	update := &conftypes.ConfigFile{
		Content: "same", Comment: "different comment", Format: "yaml",
		Metadata: map[string]string{"owner": "team-a"},
	}

	_, changed := server.updateConfigFileAttribute(saved, update)

	require.False(t, changed)
}

func TestSelectTemplateValueReleaseCanSwitchTemplateSnapshotAsOneEnvironmentRelease(t *testing.T) {
	releases := []*conftypes.NamespaceTemplateValueRelease{
		{
			ID: "wrong-template", TemplateReleaseID: "tpl-r2",
			ReleaseType: conftypes.TemplateValueReleaseTypeGray, Active: true,
			BetaLabels: []*apimodel.ClientLabel{clientLabel("region", "shanghai")},
		},
		{
			ID: "normal-r1", TemplateReleaseID: "tpl-r1",
			ReleaseType: conftypes.TemplateValueReleaseTypeNormal, Active: true,
		},
	}

	matched := selectTemplateValueRelease(releases, map[string]string{"region": "shanghai"})
	require.Equal(t, "wrong-template", matched.ID)
}

func TestPublishConfigTemplateReleaseUsesPersistedDraftSnapshot(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	schema, err := conftypes.EncodeTemplateParameterSchema(
		[]*apiconfig.ConfigTemplateParameterSchema{{
			Name: "region", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING,
			Required: true,
		}})
	require.NoError(t, err)
	storage.EXPECT().GetConfigFileTemplate("application").Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "region={{{region}}}", Comment: "saved",
		Format: "text", Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: schema,
	}, nil)
	storage.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(nil, nil)
	storage.EXPECT().CreateConfigTemplateRelease(gomock.Any()).DoAndReturn(
		func(release *conftypes.ConfigTemplateRelease) error {
			require.Equal(t, "region={{{region}}}", release.Content)
			require.Equal(t, "release-note", release.Comment)
			require.Equal(t, uint64(1), release.Version)
			return nil
		})

	response := server.PublishConfigTemplateRelease(context.Background(), &apiconfig.ConfigTemplateRelease{
		TemplateId: 7,
		Name:       "application",
		Content:    "untrusted request content",
		Comment:    "release-note",
		Engine:     &apiconfig.ConfigTemplateEngine{Name: "go-template", Version: "v1"},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestPublishConfigTemplateReleaseFallsBackToDraftComment(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	storage.EXPECT().GetConfigFileTemplate("application").Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "plain text", Comment: "saved template description",
		Format: "text", Engine: "pole-mustache", EngineVersion: "v1",
	}, nil)
	storage.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(nil, nil)
	storage.EXPECT().CreateConfigTemplateRelease(gomock.Any()).DoAndReturn(
		func(release *conftypes.ConfigTemplateRelease) error {
			require.Equal(t, "saved template description", release.Comment)
			return nil
		})

	response := server.PublishConfigTemplateRelease(context.Background(), &apiconfig.ConfigTemplateRelease{
		TemplateId: 7,
		Name:       "application",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestUpdateLegacyTemplateEndpointPreservesDefinitionFields(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	stored := &conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "prod-safe", Format: "yaml",
		Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: "[]",
	}
	storage.EXPECT().GetConfigFileTemplate("application").Return(stored, nil)
	storage.EXPECT().SaveConfigFileTemplate(gomock.Any()).DoAndReturn(
		func(updated *conftypes.ConfigFileTemplate) (*conftypes.ConfigFileTemplate, error) {
			require.Equal(t, "prod-safe", updated.Content)
			require.Equal(t, "yaml", updated.Format)
			require.Equal(t, "new identity description", updated.Comment)
			return updated, nil
		})

	response := server.UpdateConfigFileTemplate(context.Background(), &apiconfig.ConfigFileTemplate{
		Name: "application", Content: "dev-must-not-leak", Format: "text",
		Comment: "new identity description",
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestPublishEnvironmentReleaseCreatesTemplateAndValueSnapshotsAtomically(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	schema, err := conftypes.EncodeTemplateParameterSchema(
		[]*apiconfig.ConfigTemplateParameterSchema{{
			Name: "region", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING,
			Required: true,
		}})
	require.NoError(t, err)
	storage.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "region={{{region}}}", Comment: "draft",
		Format: "text", Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: schema,
	}, nil)
	storage.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(nil, nil)
	storage.EXPECT().ListNamespaceTemplateValueReleases("prod", uint64(7)).Return(nil, nil)
	storage.EXPECT().CreateConfigTemplateEnvironmentRelease(gomock.Any(), gomock.Any()).DoAndReturn(
		func(template *conftypes.ConfigTemplateRelease, release *conftypes.NamespaceTemplateValueRelease) error {
			require.NotNil(t, template)
			require.Equal(t, uint64(1), template.Version)
			require.Equal(t, template.ID, release.TemplateReleaseID)
			require.Equal(t, uint64(1), release.Version)
			require.Equal(t, "prod", release.Namespace)
			return nil
		})

	response := server.PublishNamespaceTemplateValueRelease(context.Background(),
		&apiconfig.NamespaceTemplateValueRelease{
			Namespace: "prod", TemplateId: 7, Active: true,
			ReleaseType: apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_NORMAL,
			Values: map[string]*apiconfig.ConfigTemplateValue{
				"region": {Value: &apiconfig.ConfigTemplateValue_StringValue{StringValue: "shanghai"}},
			},
		})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestPublishEnvironmentReleaseReusesIdenticalTemplateSnapshot(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	schema, err := conftypes.EncodeTemplateParameterSchema(nil)
	require.NoError(t, err)
	draft := &conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "plain", Format: "text",
		Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: schema,
	}
	existing := &conftypes.ConfigTemplateRelease{
		ID: "template-snapshot-1", TemplateID: 7, Name: draft.Name, Content: draft.Content,
		Format: draft.Format, Engine: draft.Engine, EngineVersion: draft.EngineVersion,
		ParameterSchema: schema, Version: 1,
	}
	storage.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(draft, nil)
	storage.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(
		[]*conftypes.ConfigTemplateRelease{existing}, nil)
	storage.EXPECT().ListNamespaceTemplateValueReleases("prod", uint64(7)).Return(nil, nil)
	storage.EXPECT().CreateConfigTemplateEnvironmentRelease(nil, gomock.Any()).DoAndReturn(
		func(_ *conftypes.ConfigTemplateRelease, release *conftypes.NamespaceTemplateValueRelease) error {
			require.Equal(t, existing.ID, release.TemplateReleaseID)
			return nil
		})

	response := server.PublishNamespaceTemplateValueRelease(context.Background(),
		&apiconfig.NamespaceTemplateValueRelease{
			Namespace: "prod", TemplateId: 7, Active: true,
			ReleaseType: apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_NORMAL,
		})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestBindConfigFileTemplateUpdatesDraftAndBindingAtomically(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	tx := storemock.NewMockTx(controller)
	server := &Server{storage: storage, chains: newConfigChains(nil, nil)}
	storage.EXPECT().GetConfigTemplateRelease("tpl-r1").Return(&conftypes.ConfigTemplateRelease{
		ID: "tpl-r1", TemplateID: 7,
	}, nil)
	storage.EXPECT().ListConfigTemplateBindings(&conftypes.ConfigFileKey{
		Namespace: "prod", Group: "app", Name: "application.yaml",
	}).Return(nil, nil)
	storage.EXPECT().StartTx().Return(tx, nil)
	storage.EXPECT().LockConfigFile(tx, &conftypes.ConfigFileKey{
		Namespace: "prod", Group: "app", Name: "application.yaml",
	}).Return(&conftypes.ConfigFile{
		Name: "application.yaml", Namespace: "prod", Group: "app",
		Content: "old", Metadata: map[string]string{"owner": "team-a"},
	}, nil)
	storage.EXPECT().CreateConfigTemplateBindingTx(tx, gomock.Any()).Return(nil)
	storage.EXPECT().UpdateConfigFileTx(tx, gomock.Any()).DoAndReturn(
		func(_ storeapi.Tx, file *conftypes.ConfigFile) error {
			require.Equal(t, "new", file.Content)
			apiFile := conftypes.ToConfigFileAPI(file)
			require.Equal(t, apiconfig.ConfigFile_CONFIG_TEMPLATE, apiFile.GetConfigType())
			require.Equal(t, "tpl-r1", apiFile.GetTemplateBinding().GetTemplateReleaseId())
			require.NotEmpty(t, apiFile.GetTemplateBinding().GetBindingReleaseId())
			return nil
		})
	tx.EXPECT().Commit().Return(nil)
	tx.EXPECT().Rollback().Return(nil)

	response := server.BindConfigFileTemplate(context.Background(), &apiconfig.ConfigFile{
		Name: "application.yaml", Namespace: "prod", Group: "app", Content: "new", Format: "yaml",
		Labels:     map[string]string{"owner": "team-a"},
		ConfigType: apiconfig.ConfigFile_CONFIG_TEMPLATE,
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId: 7, TemplateReleaseId: "tpl-r1",
		},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestResolveTemplateSnapshotUsesMatchedEnvironmentReleaseAcrossLegacyTemplatePin(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	file := &conftypes.ConfigFileKey{Namespace: "prod", Group: "app", Name: "application.yaml"}

	templateSpec := &apiconfig.ConfigTemplateRelease{
		Id: "tpl-r2", TemplateId: 7, Content: "region={{{region}}}", Format: "text",
		Engine: &apiconfig.ConfigTemplateEngine{Name: "pole-mustache", Version: "v1"},
		ParameterSchema: []*apiconfig.ConfigTemplateParameterSchema{{
			Name: "region", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING, Required: true,
		}},
	}
	templatePayload, err := conftypes.EncodeTemplateParameterSchema(templateSpec.GetParameterSchema())
	require.NoError(t, err)
	valueSpec := &apiconfig.NamespaceTemplateValueRelease{
		Id: "value-gray", ValuesId: "values-1", Namespace: "prod", TemplateId: 7,
		TemplateReleaseId: "tpl-r2",
		Values: map[string]*apiconfig.ConfigTemplateValue{
			"region": {Value: &apiconfig.ConfigTemplateValue_StringValue{StringValue: "shanghai"}},
		},
	}
	valuePayload, err := conftypes.EncodeTemplateValues(valueSpec.GetValues())
	require.NoError(t, err)

	storage.EXPECT().GetConfigTemplateRelease("tpl-r2").Return(&conftypes.ConfigTemplateRelease{
		ID: "tpl-r2", TemplateID: 7, Content: templateSpec.Content, Format: "text",
		Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: templatePayload,
	}, nil)
	storage.EXPECT().ListNamespaceTemplateValueReleases("prod", uint64(7)).Return(
		[]*conftypes.NamespaceTemplateValueRelease{
			{
				ID: "value-gray", ValuesID: "values-1", Namespace: "prod", TemplateID: 7,
				TemplateReleaseID: "tpl-r2", Values: valuePayload,
				ReleaseType: conftypes.TemplateValueReleaseTypeGray, Active: true,
				BetaLabels: []*apimodel.ClientLabel{clientLabel("region", "shanghai")},
			},
			{
				ID: "value-normal", Namespace: "prod", TemplateID: 7,
				TemplateReleaseID: "tpl-r2", ReleaseType: conftypes.TemplateValueReleaseTypeNormal, Active: true,
			},
		}, nil)

	snapshot, err := server.resolveTemplateSnapshot(
		context.Background(), file, &apiconfig.ConfigTemplateBinding{
			BindingReleaseId:  "binding-r1",
			TemplateId:        7,
			TemplateReleaseId: "legacy-pinned-template",
		}, map[string]string{"region": "shanghai"})
	require.NoError(t, err)
	require.Equal(t, "value-gray", snapshot.GetValueRelease().GetId())
	require.Equal(t, "tpl-r2", snapshot.GetTemplateRelease().GetId())
	require.Equal(t, "tpl-r2", snapshot.GetTemplateBinding().GetTemplateReleaseId())
	require.Equal(t, "shanghai", snapshot.GetValueRelease().GetValues()["region"].GetStringValue())
	require.Empty(t, snapshot.GetValueRelease().GetBetaLabels())
	require.NotEmpty(t, snapshot.GetExpectedRenderedSha256())
	require.NotEmpty(t, snapshot.GetRevision())
}

func TestTemplateManagementQueriesReturnSpecResponses(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	schema, err := conftypes.EncodeTemplateParameterSchema(
		[]*apiconfig.ConfigTemplateParameterSchema{{
			Name: "region", Type: apiconfig.ConfigTemplateParameterType_TEMPLATE_PARAMETER_STRING,
		}})
	require.NoError(t, err)
	values, err := conftypes.EncodeTemplateValues(map[string]*apiconfig.ConfigTemplateValue{
		"region": {Value: &apiconfig.ConfigTemplateValue_StringValue{StringValue: "shanghai"}},
	})
	require.NoError(t, err)
	now := time.Unix(1000, 0)

	storage.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(
		[]*conftypes.ConfigTemplateRelease{{
			ID: "tpl-r1", TemplateID: 7, Content: "region={{{region}}}", Format: "text",
			ParameterSchema: schema, Engine: "pole-mustache", EngineVersion: "v1",
			Version: 1, CreateTime: now,
		}}, nil)
	storage.EXPECT().GetNamespaceTemplateValues("prod", uint64(7)).Return(
		&conftypes.NamespaceTemplateValues{
			ID: "values-1", Namespace: "prod", TemplateID: 7, Values: values,
			Revision: "draft-r1", CreateTime: now, ModifyTime: now,
		}, nil)
	storage.EXPECT().ListNamespaceTemplateValueReleases("prod", uint64(7)).Return(
		[]*conftypes.NamespaceTemplateValueRelease{{
			ID: "value-r1", ValuesID: "values-1", Namespace: "prod", TemplateID: 7,
			TemplateReleaseID: "tpl-r1", Values: values,
			ReleaseType: conftypes.TemplateValueReleaseTypeNormal, Active: true,
			Version: 1, Revision: "value-revision", CreateTime: now,
		}}, nil)
	storage.EXPECT().ListConfigTemplateBindings(&conftypes.ConfigFileKey{
		Namespace: "prod", Group: "app", Name: "application.yaml",
	}).Return([]*conftypes.ConfigTemplateBinding{{
		BindingReleaseID: "binding-r1", TemplateID: 7, TemplateReleaseID: "tpl-r1",
	}}, nil)

	templateResp := server.ListConfigTemplateReleases(context.Background(), 7)
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), templateResp.GetCode())
	require.Equal(t, uint32(1), templateResp.GetAmount())
	templateRelease := &apiconfig.ConfigTemplateRelease{}
	require.NoError(t, anypb.UnmarshalTo(templateResp.GetData()[0], templateRelease, proto.UnmarshalOptions{}))
	require.Equal(t, "tpl-r1", templateRelease.GetId())
	require.NotEmpty(t, templateRelease.GetCtime())

	valuesResp := server.GetNamespaceTemplateValues(context.Background(), "prod", 7)
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), valuesResp.GetCode())
	valuesSpec := &apiconfig.NamespaceTemplateValues{}
	require.NoError(t, anypb.UnmarshalTo(valuesResp.GetData(), valuesSpec, proto.UnmarshalOptions{}))
	require.Equal(t, "shanghai", valuesSpec.GetValues()["region"].GetStringValue())

	valueReleasesResp := server.ListNamespaceTemplateValueReleases(context.Background(), "prod", 7)
	require.Equal(t, uint32(1), valueReleasesResp.GetSize())
	valueRelease := &apiconfig.NamespaceTemplateValueRelease{}
	require.NoError(t, anypb.UnmarshalTo(
		valueReleasesResp.GetData()[0], valueRelease, proto.UnmarshalOptions{}))
	require.Equal(t, "value-r1", valueRelease.GetId())

	bindingsResp := server.ListConfigTemplateBindings(
		context.Background(), "prod", "app", "application.yaml")
	require.Equal(t, uint32(1), bindingsResp.GetSize())
	binding := &apiconfig.ConfigTemplateBinding{}
	require.NoError(t, anypb.UnmarshalTo(bindingsResp.GetData()[0], binding, proto.UnmarshalOptions{}))
	require.Equal(t, "binding-r1", binding.GetBindingReleaseId())
}

func clientLabel(key, value string) *apimodel.ClientLabel {
	return &apimodel.ClientLabel{
		Key: key,
		Value: &apimodel.MatchString{
			Type: apimodel.MatchString_EXACT, Value: value,
		},
	}
}
