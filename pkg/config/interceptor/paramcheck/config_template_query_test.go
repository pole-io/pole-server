package paramcheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestTemplateManagementQueriesRejectInvalidIdentity(t *testing.T) {
	server := &Server{}
	ctx := context.Background()

	require.Equal(t, uint32(apimodel.Code_BadRequest),
		server.ListConfigTemplateReleases(ctx, 0).GetCode())
	require.Equal(t, uint32(apimodel.Code_BadRequest),
		server.GetNamespaceTemplateValues(ctx, "", 7).GetCode())
	require.Equal(t, uint32(apimodel.Code_BadRequest),
		server.ListNamespaceTemplateValueReleases(ctx, "prod", 0).GetCode())
	require.Equal(t, uint32(apimodel.Code_BadRequest),
		server.ListConfigTemplateBindings(ctx, "prod", "", "application.yaml").GetCode())
}
