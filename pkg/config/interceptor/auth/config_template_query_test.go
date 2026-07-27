package config_auth

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/config"
	authmock "github.com/pole-io/pole-server/plugin/access_control/auth/mock"
)

type templateQueryNextServer struct {
	config.ConfigCenterServer
	called bool
}

func (s *templateQueryNextServer) ListConfigTemplateReleases(
	context.Context, uint64) *apimodel.BatchQueryResponse {
	s.called = true
	return &apimodel.BatchQueryResponse{Code: uint32(apimodel.Code_ExecuteSuccess)}
}

func TestTemplateManagementQueryChecksConsolePermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	checker := authmock.NewMockAuthChecker(ctrl)
	checker.EXPECT().CheckConsolePermission(gomock.Any()).
		Return(false, errors.New("permission denied"))
	next := &templateQueryNextServer{}
	server := &Server{
		nextServer: next,
		policySvr: &templatePreviewStrategyServer{
			checker: authapi.AuthChecker(checker),
		},
	}

	response := server.ListConfigTemplateReleases(context.Background(), 7)

	require.Equal(t, uint32(apimodel.Code_NotAllowedAccess), response.GetCode())
	require.False(t, next.called)
}
