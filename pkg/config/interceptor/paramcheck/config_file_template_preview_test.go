package paramcheck

import (
	"context"
	"testing"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestPreviewConfigTemplateUsesStandardInvalidParameterCode(t *testing.T) {
	server := &Server{cfg: Config{ContentMaxLength: 1024}}

	preview := server.PreviewConfigTemplate(context.Background(), &apiconfig.RenderPreviewRequest{})

	if preview.GetCode() != uint32(apimodel.Code_InvalidParameter) {
		t.Fatalf("preview code = %d, want %d", preview.GetCode(), apimodel.Code_InvalidParameter)
	}
	if len(preview.GetDiagnostics()) != 0 {
		t.Fatalf("request validation failure must not be a render diagnostic: %+v", preview.GetDiagnostics())
	}
}

func TestPreviewConfigTemplateRejectsUnsupportedEngineAsAPIError(t *testing.T) {
	server := &Server{cfg: Config{ContentMaxLength: 1024}}

	preview := server.PreviewConfigTemplate(context.Background(), &apiconfig.RenderPreviewRequest{
		Input: &apiconfig.ConfigTemplateRenderInput{
			Content: "region={{{region}}}",
			Engine:  &apiconfig.ConfigTemplateEngine{Name: "go-template", Version: "v1"},
		},
	})

	if preview.GetCode() != uint32(apimodel.Code_InvalidParameter) {
		t.Fatalf("preview code = %d, want %d", preview.GetCode(), apimodel.Code_InvalidParameter)
	}
	if len(preview.GetDiagnostics()) != 0 {
		t.Fatalf("unsupported engine must not be a render diagnostic: %+v", preview.GetDiagnostics())
	}
}
