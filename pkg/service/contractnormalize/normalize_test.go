package contractnormalize_test

import (
	"errors"
	"testing"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/pkg/service/contractnormalize"
)

func TestNormalizeCanonicalizesProtocolAndContractType(t *testing.T) {
	input := &apiservice.ServiceContract{
		Service:  "payments",
		Version:  "v1",
		Protocol: " HTTP ",
		Name:     " OpenAPI ",
		Content:  `{"openapi":"3.0.3","paths":{"/health":{"get":{}}}}`,
		Interfaces: []*apiservice.InterfaceDescriptor{{
			Path: "/health", Method: "GET",
		}},
	}

	got, err := contractnormalize.Normalize(input)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got.Protocol != "http" {
		t.Fatalf("Protocol = %q, want %q", got.Protocol, "http")
	}
	if got.Type != "openapi" || got.Name != "openapi" {
		t.Fatalf("Type/Name = %q/%q, want openapi/openapi", got.Type, got.Name)
	}
	if input.Protocol != " HTTP " || input.Name != " OpenAPI " || input.Type != "" {
		t.Fatalf("Normalize() mutated input: %#v", input)
	}
}

func TestNormalizeValidatesSupportedProtocols(t *testing.T) {
	for _, protocol := range []string{"http", "GRPC", " Dubbo ", "thrift"} {
		t.Run(protocol, func(t *testing.T) {
			input := &apiservice.ServiceContract{
				Service:    "payments",
				Version:    "v1",
				Protocol:   protocol,
				Content:    `{"openapi":"3.0.3","paths":{"/health":{"get":{}}}}`,
				Interfaces: []*apiservice.InterfaceDescriptor{{Path: "example.Service", Method: "Call"}},
			}
			if protocol != "http" {
				input.Content = "raw contract"
			}
			got, err := contractnormalize.Normalize(input)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got.Protocol == "" {
				t.Fatal("Protocol is empty")
			}
		})
	}

	_, err := contractnormalize.Normalize(&apiservice.ServiceContract{Protocol: "graphql"})
	if !errors.Is(err, contractnormalize.ErrUnsupportedProtocol) {
		t.Fatalf("Normalize() error = %v, want ErrUnsupportedProtocol", err)
	}
}

func TestNormalizeRequiresAndPreservesStructuredRPCInterfaces(t *testing.T) {
	for _, protocol := range []string{"grpc", "dubbo", "thrift"} {
		t.Run(protocol, func(t *testing.T) {
			_, err := contractnormalize.Normalize(&apiservice.ServiceContract{
				Service: "payments", Version: "v1", Protocol: protocol, Content: "raw contract",
			})
			if !errors.Is(err, contractnormalize.ErrInterfacesRequired) {
				t.Fatalf("Normalize() error = %v, want ErrInterfacesRequired", err)
			}

			input := &apiservice.ServiceContract{
				Service:  "payments",
				Version:  "v1",
				Protocol: protocol,
				Content:  "raw contract",
				Interfaces: []*apiservice.InterfaceDescriptor{{
					Path:    "example.v1.Greeter",
					Method:  "SayHello",
					Content: `{"requestType":"HelloRequest"}`,
				}},
			}
			got, err := contractnormalize.Normalize(input)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if len(got.Interfaces) != 1 || got.Interfaces[0].Content != input.Interfaces[0].Content {
				t.Fatalf("Interfaces = %#v, want reported interface preserved", got.Interfaces)
			}
			if got.Interfaces[0] == input.Interfaces[0] {
				t.Fatal("Normalize() returned caller-owned interface pointer")
			}
		})
	}
}

func TestNormalizeProjectsDubboNativeMetadataWithoutDiscardingSourceKeys(t *testing.T) {
	input := &apiservice.ServiceContract{
		Service:  "order-provider",
		Version:  "v1",
		Protocol: "dubbo",
		Content:  `{"app":"order-provider","revision":"app-revision","services":{}}`,
		Metadata: map[string]string{
			"application":          "order-provider",
			"interface":            "com.demo.order.api.OrderService",
			"group":                "shop",
			"version":              "1.2.0",
			"side":                 "provider",
			"metadata-type":        "remote",
			"metadata-revision":    "app-revision",
			"mapping-applications": "order-provider,order-provider-canary",
		},
		Interfaces: []*apiservice.InterfaceDescriptor{{
			Path:   "com.demo.order.api.OrderService",
			Method: "getOrder",
			Type:   "getOrder(java.lang.String)",
		}},
	}

	got, err := contractnormalize.Normalize(input)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	for key, want := range map[string]string{
		"dubbo.application":          "order-provider",
		"dubbo.interface":            "com.demo.order.api.OrderService",
		"dubbo.group":                "shop",
		"dubbo.version":              "1.2.0",
		"dubbo.side":                 "provider",
		"dubbo.metadata-type":        "remote",
		"dubbo.metadata-revision":    "app-revision",
		"dubbo.mapping-applications": "order-provider,order-provider-canary",
	} {
		if got.Metadata[key] != want {
			t.Errorf("Metadata[%q] = %q, want %q", key, got.Metadata[key], want)
		}
	}
	if got.Metadata["application"] != "order-provider" {
		t.Fatal("Normalize() discarded the original Dubbo metadata key")
	}
	if _, exists := input.Metadata["dubbo.application"]; exists {
		t.Fatal("Normalize() mutated caller-owned metadata")
	}
}

func TestNormalizeExtractsDubboProviderDefinition(t *testing.T) {
	got, err := contractnormalize.Normalize(&apiservice.ServiceContract{
		Service:  "order-provider",
		Version:  "v1",
		Protocol: "dubbo",
		Content: `{
			"parameters": {
				"side": "provider",
				"application": "order-provider",
				"interface": "com.demo.order.api.OrderService",
				"group": "shop",
				"version": "1.2.0"
			},
			"canonicalName": "com.demo.order.api.OrderService",
			"methods": [
				{"name": "getOrder", "parameterTypes": ["java.lang.String"], "returnType": "com.demo.Order"},
				{"name": "getOrder", "parameterTypes": ["long"], "returnType": "com.demo.Order"}
			],
			"types": []
		}`,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(got.Interfaces) != 2 {
		t.Fatalf("len(Interfaces) = %d, want 2", len(got.Interfaces))
	}
	if got.Interfaces[0].Path != "com.demo.order.api.OrderService" ||
		got.Interfaces[0].Type != "getOrder(java.lang.String)" {
		t.Fatalf("Interfaces[0] = %#v", got.Interfaces[0])
	}
	if got.Interfaces[1].Type != "getOrder(long)" {
		t.Fatalf("Interfaces[1].Type = %q, want getOrder(long)", got.Interfaces[1].Type)
	}
	if got.Metadata["dubbo.application"] != "order-provider" ||
		got.Metadata["dubbo.service-key"] != "shop/com.demo.order.api.OrderService:1.2.0" {
		t.Fatalf("Metadata = %#v", got.Metadata)
	}
}

func TestNormalizeExtractsDubboApplicationRevisionSnapshot(t *testing.T) {
	got, err := contractnormalize.Normalize(&apiservice.ServiceContract{
		Service:  "order-provider",
		Version:  "v1",
		Protocol: "dubbo",
		Content: `{
			"app": "order-provider",
			"revision": "da3be833baa2088c5f6776fb7ab1a436",
			"services": {
				"com.demo.order.api.OrderService:dubbo": {
					"name": "com.demo.order.api.OrderService",
					"protocol": "dubbo",
					"path": "com.demo.order.api.OrderService",
					"params": {
						"side": "provider",
						"methods": "createOrder,getOrder",
						"group": "shop",
						"version": "1.2.0",
						"metadata-type": "remote"
					}
				}
			}
		}`,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(got.Interfaces) != 2 ||
		got.Interfaces[0].Method != "createOrder" ||
		got.Interfaces[1].Method != "getOrder" {
		t.Fatalf("Interfaces = %#v", got.Interfaces)
	}
	if got.Metadata["dubbo.application"] != "order-provider" ||
		got.Metadata["dubbo.metadata-revision"] != "da3be833baa2088c5f6776fb7ab1a436" ||
		got.Metadata["dubbo.metadata-type"] != "remote" ||
		got.Metadata["dubbo.service-key"] != "shop/com.demo.order.api.OrderService:1.2.0" ||
		got.Metadata["dubbo.snapshot-service-key"] != "com.demo.order.api.OrderService:dubbo" {
		t.Fatalf("Metadata = %#v", got.Metadata)
	}
}

func TestNormalizeExtractsDubboConsumerMetadata(t *testing.T) {
	got, err := contractnormalize.Normalize(&apiservice.ServiceContract{
		Service:  "order-consumer",
		Version:  "v1",
		Protocol: "dubbo",
		Content: `{
			"valid": "true",
			"side": "consumer",
			"application": "order-consumer",
			"methods": "createOrder,getOrder",
			"interface": "com.demo.order.api.OrderService",
			"version": "1.2.0",
			"group": "shop"
		}`,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(got.Interfaces) != 2 || got.Interfaces[1].Method != "getOrder" {
		t.Fatalf("Interfaces = %#v", got.Interfaces)
	}
	if got.Metadata["dubbo.side"] != "consumer" ||
		got.Metadata["dubbo.application"] != "order-consumer" ||
		got.Metadata["dubbo.service-key"] != "shop/com.demo.order.api.OrderService:1.2.0" {
		t.Fatalf("Metadata = %#v", got.Metadata)
	}
}

func TestNormalizeExtractsStableHTTPInterfacesFromOpenAPIJSON(t *testing.T) {
	input := &apiservice.ServiceContract{
		Service:  "payments",
		Version:  "v1",
		Protocol: "http",
		Content: `{
			"openapi": "3.0.3",
			"paths": {
				"/users/{id}": {
					"parameters": [],
					"delete": {"operationId": "deleteUser"}
				},
				"/pets": {
					"post": {"operationId": "createPet"},
					"get": {"operationId": "listPets"}
				}
			}
		}`,
	}

	got, err := contractnormalize.Normalize(input)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got.Type != "openapi" || got.Name != "openapi" {
		t.Fatalf("Type/Name = %q/%q, want openapi/openapi", got.Type, got.Name)
	}

	want := [][2]string{
		{"/pets", "GET"},
		{"/pets", "POST"},
		{"/users/{id}", "DELETE"},
	}
	if len(got.Interfaces) != len(want) {
		t.Fatalf("len(Interfaces) = %d, want %d: %#v", len(got.Interfaces), len(want), got.Interfaces)
	}
	for i, pair := range want {
		if got.Interfaces[i].Path != pair[0] || got.Interfaces[i].Method != pair[1] {
			t.Errorf("Interfaces[%d] = %s %s, want %s %s",
				i, got.Interfaces[i].Method, got.Interfaces[i].Path, pair[1], pair[0])
		}
	}
	if got.Interfaces[0].Content != `{"operationId":"listPets"}` {
		t.Errorf("Interfaces[0].Content = %q, want original OpenAPI operation", got.Interfaces[0].Content)
	}
}

func TestNormalizeExtractsHTTPInterfacesFromOpenAPIYAML(t *testing.T) {
	input := &apiservice.ServiceContract{
		Service:  "payments",
		Version:  "v1",
		Protocol: "HTTP",
		Content: `openapi: 3.1.0
paths:
  /orders:
    post:
      operationId: createOrder
    get:
      operationId: listOrders
`,
	}

	got, err := contractnormalize.Normalize(input)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(got.Interfaces) != 2 {
		t.Fatalf("len(Interfaces) = %d, want 2", len(got.Interfaces))
	}
	if got.Interfaces[0].Path != "/orders" || got.Interfaces[0].Method != "GET" {
		t.Errorf("Interfaces[0] = %#v, want GET /orders", got.Interfaces[0])
	}
	if got.Interfaces[1].Path != "/orders" || got.Interfaces[1].Method != "POST" {
		t.Errorf("Interfaces[1] = %#v, want POST /orders", got.Interfaces[1])
	}
}

func TestNormalizeRejectsInvalidOrNonV3OpenAPIContent(t *testing.T) {
	tests := map[string]string{
		"swagger v2": `{"swagger":"2.0","paths":{"/pets":{"get":{}}}}`,
		"no paths":   `{"openapi":"3.0.3","info":{"title":"empty","version":"1"}}`,
		"malformed":  `openapi: [`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := contractnormalize.Normalize(&apiservice.ServiceContract{
				Service: "payments", Version: "v1", Protocol: "http", Content: content,
			})
			if !errors.Is(err, contractnormalize.ErrInvalidOpenAPI) {
				t.Fatalf("Normalize() error = %v, want ErrInvalidOpenAPI", err)
			}
		})
	}
}

func TestNormalizeRejectsInvalidOpenAPIEvenWithReportedInterfaces(t *testing.T) {
	_, err := contractnormalize.Normalize(&apiservice.ServiceContract{
		Service: "payments", Version: "v1", Protocol: "http", Content: `{"swagger":"2.0"}`,
		Interfaces: []*apiservice.InterfaceDescriptor{{Path: "/payments", Method: "GET"}},
	})

	if !errors.Is(err, contractnormalize.ErrInvalidOpenAPI) {
		t.Fatalf("Normalize() error = %v, want ErrInvalidOpenAPI", err)
	}
}

func TestNormalizeRejectsIncompleteStructuredInterfaces(t *testing.T) {
	for name, descriptor := range map[string]*apiservice.InterfaceDescriptor{
		"missing path":   {Method: "Call"},
		"missing method": {Path: "example.Service"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := contractnormalize.Normalize(&apiservice.ServiceContract{
				Service:    "payments",
				Version:    "v1",
				Protocol:   "grpc",
				Content:    "raw contract",
				Interfaces: []*apiservice.InterfaceDescriptor{descriptor},
			})
			if !errors.Is(err, contractnormalize.ErrInvalidInterface) {
				t.Fatalf("Normalize() error = %v, want ErrInvalidInterface", err)
			}
		})
	}
}

func TestNormalizeRejectsNilContract(t *testing.T) {
	_, err := contractnormalize.Normalize(nil)
	if !errors.Is(err, contractnormalize.ErrNilContract) {
		t.Fatalf("Normalize() error = %v, want ErrNilContract", err)
	}
}

func TestNormalizeRequiresIdentityAndRawContent(t *testing.T) {
	for name, contract := range map[string]*apiservice.ServiceContract{
		"service": {Protocol: "grpc", Version: "v1", Content: "proto",
			Interfaces: []*apiservice.InterfaceDescriptor{{Path: "example.Service", Method: "Call"}}},
		"version": {Protocol: "grpc", Service: "payments", Content: "proto",
			Interfaces: []*apiservice.InterfaceDescriptor{{Path: "example.Service", Method: "Call"}}},
		"content": {Protocol: "grpc", Service: "payments", Version: "v1",
			Interfaces: []*apiservice.InterfaceDescriptor{{Path: "example.Service", Method: "Call"}}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := contractnormalize.Normalize(contract)
			if !errors.Is(err, contractnormalize.ErrInvalidContract) {
				t.Fatalf("Normalize() error = %v, want ErrInvalidContract", err)
			}
		})
	}
}
