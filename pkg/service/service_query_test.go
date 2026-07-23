package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

func TestBuildServiceQueryItemUsesCredentialSafeListProjection(t *testing.T) {
	got := buildServiceQueryItem(&svctypes.Service{
		ID:         "svc-1",
		Token:      "control-plane-secret",
		Name:       "orders",
		Namespace:  "default",
		Business:   "checkout",
		Department: "platform",
		Comment:    "order service",
		Meta: map[string]string{
			"env": "prod",
		},
	}, 3, 2)

	require.NotNil(t, got)
	assert.Empty(t, got.GetId())
	assert.Empty(t, got.GetToken())
	assert.Equal(t, "orders", got.GetName())
	assert.Equal(t, "default", got.GetNamespace())
	assert.Empty(t, got.GetBusiness())
	assert.Empty(t, got.GetDepartment())
	assert.Empty(t, got.GetComment())
	assert.Equal(t, map[string]string{"env": "prod"}, got.GetMetadata())
	assert.Equal(t, uint32(3), got.GetTotalInstanceCount())
	assert.Equal(t, uint32(2), got.GetHealthyInstanceCount())
	assert.False(t, got.GetEditable())
	assert.False(t, got.GetDeleteable())
}
