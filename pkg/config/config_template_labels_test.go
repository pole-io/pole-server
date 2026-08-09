package config

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	storemock "github.com/pole-io/pole-server/plugin/store/mock"
)

func TestConfigTemplateLabelsArePersistedSeparatelyAndReturned(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	template := &conftypes.ConfigFileTemplate{Id: 7, Name: "application", Labels: map[string]string{"team": "platform"}}

	storage.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(template, nil)
	storage.EXPECT().SaveConfigFileTemplate(gomock.Any()).DoAndReturn(
		func(saved *conftypes.ConfigFileTemplate) (*conftypes.ConfigFileTemplate, error) {
			require.Equal(t, map[string]string{"team": "infra", "scene": "production"}, saved.Labels)
			return saved, nil
		})
	response := server.SaveConfigTemplateLabels(context.Background(), 7,
		map[string]string{"team": "infra", "scene": "production"})
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())

	storage.EXPECT().GetConfigFileTemplateByID(uint64(7)).Return(template, nil)
	response = server.GetConfigTemplateLabels(context.Background(), 7)
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
	data := &structpb.Struct{}
	require.NoError(t, anypb.UnmarshalTo(response.GetData(), data, proto.UnmarshalOptions{}))
	require.Equal(t, "infra", data.GetFields()["labels"].GetStructValue().GetFields()["team"].GetStringValue())
}

func TestTemplateContentUpdatePreservesLabels(t *testing.T) {
	controller := gomock.NewController(t)
	storage := storemock.NewMockStore(controller)
	server := &Server{storage: storage}
	storage.EXPECT().GetConfigFileTemplate("application").Return(&conftypes.ConfigFileTemplate{
		Id: 7, Name: "application", Labels: map[string]string{"team": "platform"},
	}, nil)
	storage.EXPECT().SaveConfigFileTemplate(gomock.Any()).DoAndReturn(
		func(saved *conftypes.ConfigFileTemplate) (*conftypes.ConfigFileTemplate, error) {
			require.Equal(t, map[string]string{"team": "platform"}, saved.Labels)
			return saved, nil
		})

	response := server.UpdateConfigFileTemplate(context.Background(), &apiconfig.ConfigFileTemplate{
		Id: 7, Name: "application", Content: "key=value", Format: "text",
		Engine: &apiconfig.ConfigTemplateEngine{Name: "pole-mustache", Version: "v1"},
	})
	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), response.GetCode())
}
