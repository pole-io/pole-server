package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceEnvironmentBindingToSpecKeepsOrphanSnapshot(t *testing.T) {
	binding := (&ServiceEnvironmentBinding{
		LogicalServiceID: "logical-1",
		ServiceID:        "missing-service",
		Namespace:        "prod",
		ServiceName:      "checkout-prod",
	}).ToSpec(nil)

	require.Equal(t, "logical-1", binding.GetLogicalServiceId())
	require.Equal(t, "missing-service", binding.GetServiceId())
	require.Equal(t, "prod", binding.GetNamespace())
	require.Equal(t, "checkout-prod", binding.GetServiceName())
	require.Nil(t, binding.GetService())
}
