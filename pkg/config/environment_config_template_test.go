package config

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	basetypes "github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/store"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

type namespaceDraftTestStore struct {
	store.Store
	draft *conftypes.NamespaceConfigTemplateDraft
}

func (s *namespaceDraftTestStore) GetNamespaceConfigTemplateDraft(
	_ string, _ uint64) (*conftypes.NamespaceConfigTemplateDraft, error) {
	return s.draft, nil
}

func (s *namespaceDraftTestStore) SaveNamespaceConfigTemplateDraft(
	draft *conftypes.NamespaceConfigTemplateDraft, expectedVersion uint64) error {
	if s.draft != nil && s.draft.DraftVersion != expectedVersion {
		return store.ErrNamespaceConfigTemplateDraftConflict
	}
	draft.DraftVersion = expectedVersion + 1
	copy := *draft
	s.draft = &copy
	return nil
}

func TestEnvironmentTemplateDraftInitializesFromActiveEnvironmentRelease(t *testing.T) {
	controller := gomock.NewController(t)
	base := storemock.NewMockStore(controller)
	storage := &namespaceDraftTestStore{Store: base}
	server := &Server{storage: storage}
	base.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Comment: "global identity", Content: "legacy-global",
		Format: "text", Engine: "pole-mustache", EngineVersion: "v1",
	}, nil)
	base.EXPECT().ListNamespaceTemplateValueReleases("prod", uint64(7)).Return(
		[]*conftypes.NamespaceTemplateValueRelease{{
			ID: "env-v3", TemplateReleaseID: "tpl-prod-v3", Version: 3, Active: true,
			ReleaseType: conftypes.TemplateValueReleaseTypeNormal,
		}}, nil)
	base.EXPECT().GetConfigTemplateRelease("tpl-prod-v3").Return(&conftypes.ConfigTemplateRelease{
		ID: "tpl-prod-v3", TemplateID: 7, Content: "prod={{{region}}}", Format: "yaml",
		Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: "[]",
	}, nil)

	draft, err := server.getOrInitializeNamespaceConfigTemplateDraft("prod", 7)

	require.NoError(t, err)
	require.Equal(t, "prod={{{region}}}", draft.Content)
	require.Equal(t, "environment-release:env-v3", draft.InitializedFrom)
	require.Equal(t, "application", draft.Name)
}

func TestPublishEnvironmentReleaseUsesNamespaceScopedTemplateDraft(t *testing.T) {
	controller := gomock.NewController(t)
	base := storemock.NewMockStore(controller)
	storage := &namespaceDraftTestStore{Store: base, draft: &conftypes.NamespaceConfigTemplateDraft{
		Namespace: "dev", TemplateID: 7, Content: "dev-only", Format: "text",
		Engine: "pole-mustache", EngineVersion: "v1", ParameterSchema: "[]", DraftVersion: 2,
	}}
	server := &Server{storage: storage}
	base.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Comment: "identity",
	}, nil)
	base.EXPECT().ListConfigTemplateReleases(uint64(7)).Return(nil, nil)
	base.EXPECT().ListNamespaceTemplateValueReleases("dev", uint64(7)).Return(nil, nil)
	base.EXPECT().CreateConfigTemplateEnvironmentRelease(gomock.Any(), gomock.Any()).DoAndReturn(
		func(template *conftypes.ConfigTemplateRelease, _ *conftypes.NamespaceTemplateValueRelease) error {
			require.Equal(t, "dev-only", template.Content)
			return nil
		})

	response := server.PublishNamespaceTemplateValueRelease(context.Background(),
		&apiconfig.NamespaceTemplateValueRelease{
			Namespace: "dev", TemplateId: 7, Active: true,
			ReleaseType: apiconfig.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_NORMAL,
		})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestSaveNamespaceTemplateDraftRejectsSystemNamespace(t *testing.T) {
	controller := gomock.NewController(t)
	base := storemock.NewMockStore(controller)
	server := &Server{storage: &namespaceDraftTestStore{Store: base}}
	base.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(&conftypes.ConfigFileTemplate{Id: 7}, nil)
	base.EXPECT().GetNamespace("pole-system").Return(&basetypes.Namespace{
		Name: "pole-system", Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
	}, nil)

	response := server.SaveNamespaceConfigTemplateDraft(context.Background(),
		&conftypes.NamespaceConfigTemplateDraft{
			Namespace: "pole-system", TemplateID: 7, Content: "not-allowed",
		})

	require.Equal(t, uint32(apimodel.Code_BadRequest), response.GetCode())
}
