package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestEnrichServiceContractToSpecPreservesVisualizationFields(t *testing.T) {
	created := time.Unix(100, 0)
	modified := time.Unix(200, 0)
	contract := (&EnrichServiceContract{
		ServiceContract: &ServiceContract{
			ID:            "contract-1",
			Namespace:     "default",
			Service:       "payments",
			Type:          "openapi",
			Protocol:      "http",
			Version:       "v1",
			Content:       `{"openapi":"3.0.0"}`,
			ContentDigest: "contract-digest",
			Metadata:      map[string]string{"format": "openapi-v3"},
			CreateTime:    created,
			ModifyTime:    modified,
		},
		Interfaces: []*InterfaceDescriptor{{
			ID:            "operation-1",
			Path:          "/payments/{id}",
			Method:        "GET",
			Type:          "openapi",
			ContentDigest: "operation-digest",
			Source:        apiservice.InterfaceDescriptor_Client,
			CreateTime:    created,
			ModifyTime:    modified,
		}},
	}).ToSpec()

	require.Equal(t, "contract-digest", contract.GetContentDigest())
	require.Equal(t, "openapi-v3", contract.GetMetadata()["format"])
	require.Len(t, contract.GetInterfaces(), 1)
	require.Equal(t, "operation-digest", contract.GetInterfaces()[0].GetContentDigest())
	require.NotEqual(t, contract.GetInterfaces()[0].GetCtime(), contract.GetInterfaces()[0].GetMtime())
}

func TestEnrichServiceContractFormatUsesManualOverrideAndStableOrder(t *testing.T) {
	contract := &EnrichServiceContract{
		ServiceContract: &ServiceContract{},
		Interfaces: []*InterfaceDescriptor{
			{ID: "client-b", Path: "/b", Method: "POST", Source: apiservice.InterfaceDescriptor_Client},
			{ID: "client-a", Path: "/a", Method: "GET", Source: apiservice.InterfaceDescriptor_Client},
			{ID: "manual-a", Path: "/a", Method: "GET", Source: apiservice.InterfaceDescriptor_Manual},
		},
	}

	contract.Format()

	require.Len(t, contract.Interfaces, 2)
	require.Equal(t, "manual-a", contract.Interfaces[0].ID)
	require.Equal(t, "client-b", contract.Interfaces[1].ID)
}

func TestEnrichServiceContractFormatPreservesDubboOverloads(t *testing.T) {
	contract := &EnrichServiceContract{
		ServiceContract: &ServiceContract{},
		Interfaces: []*InterfaceDescriptor{
			{ID: "pay-string", Path: "com.example.PaymentService", Method: "pay", Type: "pay(string)",
				Source: apiservice.InterfaceDescriptor_Client},
			{ID: "pay-long", Path: "com.example.PaymentService", Method: "pay", Type: "pay(long)",
				Source: apiservice.InterfaceDescriptor_Client},
		},
	}

	contract.Format()

	require.Len(t, contract.Interfaces, 2)
}
