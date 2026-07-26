package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/config"
)

type configDiscoverStub struct {
	config.ConfigCenterServer
	revision string
	labels   map[string]string
}

func (s *configDiscoverStub) GetConfigFileWithCache(
	_ context.Context, file *apiconfig.ConfigFile,
) *apiconfig.ConfigDiscoverResponse {
	s.revision = file.GetId()
	s.labels = file.GetLabels()
	resp := api.NewConfigDiscoverResponse(apimodel.Code_DataNoChange)
	resp.Revision = s.revision
	return resp
}

func TestConfigDiscoverUsesTopLevelRevision(t *testing.T) {
	stub := &configDiscoverStub{}
	server := &ConfigGRPCServer{configServer: stub}

	resp := server.handleDiscoverRequest(context.Background(), &apiconfig.ConfigDiscoverRequest{
		Type:     apiconfig.ConfigDiscoverRequest_CONFIG_FILE,
		Revision: "snapshot-revision",
		File: &apiconfig.ConfigFile{
			Id: "legacy-file-id", Namespace: "default", Group: "group-a", Name: "app.yaml",
		},
		Filter: &apiconfig.ConfigDiscoverFilter{Caller: &apimodel.Caller{
			Labels: []*apimodel.ClientLabel{{
				Key: "env", Value: &apimodel.MatchString{Type: apimodel.MatchString_EXACT, Value: "canary"},
			}},
		}},
	})

	require.Equal(t, "snapshot-revision", stub.revision)
	require.Equal(t, "snapshot-revision", resp.GetRevision())
	require.Equal(t, "canary", stub.labels["env"])
}
