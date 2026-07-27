package config

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storeapi "github.com/pole-io/pole-server/apis/store"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestCreateTemplateConfigFileCreatesPinnedBindingAtomically(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	tx := storemock.NewMockTx(controller)
	server := &Server{storage: storage, chains: newConfigChains(nil, nil)}

	storage.EXPECT().GetConfigFileGroup("prod", "application").Return(
		&conftypes.ConfigFileGroup{Namespace: "prod", Name: "application"}, nil)
	storage.EXPECT().GetConfigTemplateRelease("template-release-1").Return(
		&conftypes.ConfigTemplateRelease{ID: "template-release-1", TemplateID: 7}, nil)
	storage.EXPECT().StartTx().Return(tx, nil)
	storage.EXPECT().GetConfigFileTx(tx, "prod", "application", "application.yaml").Return(nil, nil)
	storage.EXPECT().CreateConfigFileTx(tx, gomock.Any()).DoAndReturn(
		func(_ storeapi.Tx, file *conftypes.ConfigFile) error {
			require.Equal(t, conftypes.ConfigFileTypeTemplate, file.ConfigType)
			apiFile := conftypes.ToConfigFileAPI(file)
			require.Equal(t, uint64(7), apiFile.GetTemplateBinding().GetTemplateId())
			require.Equal(t, "template-release-1", apiFile.GetTemplateBinding().GetTemplateReleaseId())
			require.NotEmpty(t, apiFile.GetTemplateBinding().GetBindingReleaseId())
			return nil
		})
	storage.EXPECT().CreateConfigTemplateBindingTx(tx, gomock.Any()).DoAndReturn(
		func(_ storeapi.Tx, binding *conftypes.ConfigTemplateBinding) error {
			require.Equal(t, "prod", binding.Namespace)
			require.Equal(t, "application", binding.Group)
			require.Equal(t, "application.yaml", binding.FileName)
			require.Equal(t, uint64(7), binding.TemplateID)
			require.Equal(t, "template-release-1", binding.TemplateReleaseID)
			require.NotEmpty(t, binding.BindingReleaseID)
			require.Equal(t, uint64(1), binding.Version)
			require.False(t, binding.Active)
			return nil
		})
	tx.EXPECT().Commit().Return(nil)
	tx.EXPECT().Rollback().Return(nil)

	response := server.CreateConfigFile(context.Background(), &apiconfig.ConfigFile{
		Namespace:  "prod",
		Group:      "application",
		Name:       "application.yaml",
		Format:     "yaml",
		ConfigType: apiconfig.ConfigFile_CONFIG_TEMPLATE,
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId:        7,
			TemplateReleaseId: "template-release-1",
		},
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}

func TestCreateTemplateConfigFileRejectsUnknownTemplateReleaseBeforeStartingTransaction(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage, chains: newConfigChains(nil, nil)}

	storage.EXPECT().GetConfigFileGroup("prod", "application").Return(
		&conftypes.ConfigFileGroup{Namespace: "prod", Name: "application"}, nil)
	storage.EXPECT().GetConfigTemplateRelease("missing-release").Return(nil, nil)

	response := server.CreateConfigFile(context.Background(), &apiconfig.ConfigFile{
		Namespace:  "prod",
		Group:      "application",
		Name:       "application.yaml",
		ConfigType: apiconfig.ConfigFile_CONFIG_TEMPLATE,
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId:        7,
			TemplateReleaseId: "missing-release",
		},
	})

	require.Equal(t, uint32(apimodel.Code_NotFoundResource), response.GetCode())
}

func TestCreatePlainConfigFileRejectsTemplateBindingBeforeStartingTransaction(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage, chains: newConfigChains(nil, nil)}

	storage.EXPECT().GetConfigFileGroup("prod", "application").Return(
		&conftypes.ConfigFileGroup{Namespace: "prod", Name: "application"}, nil)

	response := server.CreateConfigFile(context.Background(), &apiconfig.ConfigFile{
		Namespace:  "prod",
		Group:      "application",
		Name:       "application.yaml",
		Format:     "yaml",
		ConfigType: apiconfig.ConfigFile_CONFIG_FILE,
		TemplateBinding: &apiconfig.ConfigTemplateBinding{
			TemplateId:        7,
			TemplateReleaseId: "template-release-1",
		},
	})

	require.Equal(t, uint32(apimodel.Code_InvalidParameter), response.GetCode())
}
