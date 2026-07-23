package paramcheck

import (
	"testing"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	"github.com/stretchr/testify/require"
)

func TestCheckReviseInstanceAllowsClientTetradWithoutID(t *testing.T) {
	instanceID, resp := checkReviseInstance(&apiservice.Instance{
		Namespace: "default",
		Service:   "checkout",
		Host:      "127.0.0.1",
		Port:      8080,
	})

	require.Nil(t, resp)
	require.NotEmpty(t, instanceID)
}
