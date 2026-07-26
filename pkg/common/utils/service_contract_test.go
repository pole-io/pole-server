package utils

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestCheckContractInterfaceTetradDistinguishesOverloadType(t *testing.T) {
	first, response := CheckContractInterfaceTetrad("contract-1", apiservice.InterfaceDescriptor_Client,
		&apiservice.InterfaceDescriptor{Path: "com.example.PaymentService", Method: "pay", Type: "pay(string)"})
	require.Nil(t, response)
	second, response := CheckContractInterfaceTetrad("contract-1", apiservice.InterfaceDescriptor_Client,
		&apiservice.InterfaceDescriptor{Path: "com.example.PaymentService", Method: "pay", Type: "pay(long)"})
	require.Nil(t, response)

	require.NotEqual(t, first, second)
}
