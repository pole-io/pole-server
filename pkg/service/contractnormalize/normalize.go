// Package contractnormalize normalizes reported service contracts before they
// enter the service contract persistence workflow.
package contractnormalize

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

var (
	ErrNilContract         = errors.New("service contract is nil")
	ErrUnsupportedProtocol = errors.New("unsupported service contract protocol")
	ErrInterfacesRequired  = errors.New("structured service contract interfaces are required")
	ErrInvalidOpenAPI      = errors.New("invalid OpenAPI 3 service contract")
	ErrInvalidInterface    = errors.New("invalid structured service contract interface")
	ErrInvalidContract     = errors.New("invalid service contract")
)

// Normalize returns a normalized copy of contract.
func Normalize(contract *apiservice.ServiceContract) (*apiservice.ServiceContract, error) {
	if contract == nil {
		return nil, ErrNilContract
	}

	normalized := proto.Clone(contract).(*apiservice.ServiceContract)
	normalized.Protocol = strings.ToLower(strings.TrimSpace(normalized.GetProtocol()))
	switch normalized.Protocol {
	case "http", "grpc", "dubbo", "thrift":
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedProtocol, contract.GetProtocol())
	}

	contractType := normalized.GetType()
	if strings.TrimSpace(contractType) == "" {
		contractType = normalized.GetName()
	}
	if strings.TrimSpace(contractType) == "" {
		contractType = defaultContractType(normalized.Protocol)
	}
	contractType = strings.ToLower(strings.TrimSpace(contractType))
	normalized.Type = contractType
	normalized.Name = contractType

	if strings.TrimSpace(normalized.GetService()) == "" {
		return nil, fmt.Errorf("%w: service is required", ErrInvalidContract)
	}
	if strings.TrimSpace(normalized.GetVersion()) == "" {
		return nil, fmt.Errorf("%w: version is required", ErrInvalidContract)
	}
	if strings.TrimSpace(normalized.GetContent()) == "" {
		return nil, fmt.Errorf("%w: raw contract content is required", ErrInvalidContract)
	}

	if normalized.Protocol == "http" {
		interfaces, err := extractOpenAPIInterfaces(normalized.GetContent())
		if err != nil {
			return nil, err
		}
		normalized.Interfaces = interfaces
	}
	if normalized.Protocol != "http" && len(normalized.GetInterfaces()) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrInterfacesRequired, normalized.Protocol)
	}
	for i, descriptor := range normalized.GetInterfaces() {
		if descriptor == nil || strings.TrimSpace(descriptor.GetPath()) == "" ||
			strings.TrimSpace(descriptor.GetMethod()) == "" {
			return nil, fmt.Errorf("%w at index %d: path and method are required", ErrInvalidInterface, i)
		}
	}
	return normalized, nil
}

func defaultContractType(protocol string) string {
	switch protocol {
	case "http":
		return "openapi"
	case "grpc":
		return "protobuf"
	default:
		return protocol
	}
}

func extractOpenAPIInterfaces(content string) ([]*apiservice.InterfaceDescriptor, error) {
	var document struct {
		OpenAPI string                            `yaml:"openapi"`
		Paths   map[string]map[string]interface{} `yaml:"paths"`
	}
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOpenAPI, err)
	}
	if !strings.HasPrefix(strings.TrimSpace(document.OpenAPI), "3.") {
		return nil, fmt.Errorf("%w: unsupported version %q", ErrInvalidOpenAPI, document.OpenAPI)
	}
	if document.Paths == nil {
		return nil, fmt.Errorf("%w: paths is required", ErrInvalidOpenAPI)
	}

	type operation struct {
		path   string
		method string
	}
	operations := make([]operation, 0)
	for path, pathItem := range document.Paths {
		for method := range pathItem {
			method = strings.ToLower(method)
			if !isHTTPMethod(method) {
				continue
			}
			operations = append(operations, operation{path: path, method: strings.ToUpper(method)})
		}
	}
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].path != operations[j].path {
			return operations[i].path < operations[j].path
		}
		return operations[i].method < operations[j].method
	})

	interfaces := make([]*apiservice.InterfaceDescriptor, 0, len(operations))
	for _, operation := range operations {
		interfaces = append(interfaces, &apiservice.InterfaceDescriptor{
			Path:   operation.path,
			Method: operation.method,
		})
	}
	return interfaces, nil
}

func isHTTPMethod(method string) bool {
	switch method {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}
