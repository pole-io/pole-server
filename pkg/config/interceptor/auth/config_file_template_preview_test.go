package config_auth

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	"github.com/pole-io/pole-server/pkg/config"
	authmock "github.com/pole-io/pole-server/plugin/access_control/auth/mock"
)

type templatePreviewStrategyServer struct {
	authapi.StrategyServer
	checker authapi.AuthChecker
}

func (s *templatePreviewStrategyServer) GetAuthChecker() authapi.AuthChecker {
	return s.checker
}

type templatePreviewNextServer struct {
	config.ConfigCenterServer
	called bool
}

func (s *templatePreviewNextServer) PreviewConfigTemplate(
	context.Context, *apiconfig.RenderPreviewRequest) *apiconfig.RenderPreview {
	s.called = true
	return &apiconfig.RenderPreview{Code: uint32(apimodel.Code_ExecuteSuccess)}
}

func TestPreviewConfigTemplateUsesStandardAuthorizationCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	checker := authmock.NewMockAuthChecker(ctrl)
	checker.EXPECT().CheckConsolePermission(gomock.Any()).
		Return(false, errors.New("permission denied"))
	next := &templatePreviewNextServer{}
	server := &Server{
		nextServer: next,
		policySvr:  &templatePreviewStrategyServer{checker: checker},
	}

	preview := server.PreviewConfigTemplate(context.Background(), &apiconfig.RenderPreviewRequest{})

	if preview.GetCode() != uint32(apimodel.Code_NotAllowedAccess) {
		t.Fatalf("preview code = %d, want %d", preview.GetCode(), apimodel.Code_NotAllowedAccess)
	}
	if len(preview.GetDiagnostics()) != 0 {
		t.Fatalf("authorization failure must not be a render diagnostic: %+v", preview.GetDiagnostics())
	}
	if next.called {
		t.Fatal("authorization failure must not call the renderer")
	}
}
