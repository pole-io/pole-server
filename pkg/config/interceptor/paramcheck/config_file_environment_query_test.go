package paramcheck

import (
	"context"
	"testing"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/config"
)

type configFileEnvironmentQueryServer struct {
	config.ConfigCenterServer
	filter map[string]string
}

func (s *configFileEnvironmentQueryServer) SearchConfigFiles(
	_ context.Context,
	filter map[string]string,
) *apimodel.BatchQueryResponse {
	s.filter = filter
	return api.NewConfigBatchQueryResponse(apimodel.Code_ExecuteSuccess)
}

func TestSearchConfigFilesForwardsBriefSummaryFlag(t *testing.T) {
	next := &configFileEnvironmentQueryServer{}
	server := &Server{nextServer: next}

	server.SearchConfigFiles(context.Background(), map[string]string{
		"offset": "0",
		"limit":  "100",
		"group":  "application",
		"name":   "app.yaml",
		"brief":  "true",
	})

	if next.filter["brief"] != "true" {
		t.Fatalf("brief flag not forwarded: %#v", next.filter)
	}
}
