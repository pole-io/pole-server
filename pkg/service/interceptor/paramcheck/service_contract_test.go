package paramcheck

import (
	"testing"

	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestCheckBaseServiceContractAcceptsTypeWithoutDeprecatedName(t *testing.T) {
	resp := checkBaseServiceContract(&apiservice.ServiceContract{
		Namespace: "default",
		Service:   "payments",
		Protocol:  "grpc",
		Version:   "v1",
		Type:      "protobuf",
	})

	require.Nil(t, resp)
}

func TestCheckPublishServiceContractAllowsProtocolDefaultType(t *testing.T) {
	resp := checkPublishServiceContract(&apiservice.ServiceContract{
		Namespace: "default",
		Service:   "payments",
		Protocol:  "http",
		Version:   "v1",
		Content:   `{"openapi":"3.0.3","paths":{}}`,
	})

	require.Nil(t, resp)
	require.Equal(t, uint32(apimodel.Code_EmptyRequest), checkPublishServiceContract(nil).GetCode())
}
