// Package contractnormalize normalizes reported service contracts before they
// enter the service contract persistence workflow.
package contractnormalize

import (
	"encoding/json"
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
	if normalized.Protocol == "dubbo" && len(normalized.GetInterfaces()) == 0 {
		extractDubboNativeMetadata(normalized)
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
	if normalized.Protocol == "dubbo" {
		projectDubboMetadata(normalized)
	}
	return normalized, nil
}

type dubboNativeMethod struct {
	Name           string   `json:"name"`
	ParameterTypes []string `json:"parameterTypes"`
	ReturnType     string   `json:"returnType"`
}

type dubboNativeService struct {
	Name     string         `json:"name"`
	Protocol string         `json:"protocol"`
	Path     string         `json:"path"`
	Params   map[string]any `json:"params"`
}

func extractDubboNativeMetadata(contract *apiservice.ServiceContract) {
	var raw map[string]any
	var document struct {
		App           string                        `json:"app"`
		Revision      string                        `json:"revision"`
		Services      map[string]dubboNativeService `json:"services"`
		Parameters    map[string]any                `json:"parameters"`
		CanonicalName string                        `json:"canonicalName"`
		Methods       json.RawMessage               `json:"methods"`
	}
	if json.Unmarshal([]byte(contract.GetContent()), &document) != nil {
		return
	}
	_ = json.Unmarshal([]byte(contract.GetContent()), &raw)
	if contract.Metadata == nil {
		contract.Metadata = map[string]string{}
	}

	if document.App != "" {
		setMetadataDefault(contract.Metadata, "application", document.App)
	}
	if document.Revision != "" {
		setMetadataDefault(contract.Metadata, "metadata-revision", document.Revision)
	}

	if document.CanonicalName != "" {
		mergeScalarMetadata(contract.Metadata, document.Parameters)
		setMetadataDefault(contract.Metadata, "interface", document.CanonicalName)
		var methods []dubboNativeMethod
		_ = json.Unmarshal(document.Methods, &methods)
		for _, method := range methods {
			if strings.TrimSpace(method.Name) == "" {
				continue
			}
			methodJSON, _ := json.Marshal(map[string]any{
				"canonicalName": document.CanonicalName,
				"method":        method,
				"parameters":    document.Parameters,
			})
			contract.Interfaces = append(contract.Interfaces, &apiservice.InterfaceDescriptor{
				Path:    document.CanonicalName,
				Method:  method.Name,
				Type:    method.Name + "(" + strings.Join(method.ParameterTypes, ",") + ")",
				Content: string(methodJSON),
			})
		}
		if len(contract.Interfaces) == 0 {
			appendDubboMethodNames(contract, document.CanonicalName, scalarString(document.Parameters["methods"]), "")
		}
	}

	serviceKeys := make([]string, 0, len(document.Services))
	for key := range document.Services {
		serviceKeys = append(serviceKeys, key)
	}
	sort.Strings(serviceKeys)
	for _, key := range serviceKeys {
		service := document.Services[key]
		if strings.ToLower(strings.TrimSpace(service.Protocol)) != "" &&
			strings.ToLower(strings.TrimSpace(service.Protocol)) != "dubbo" {
			continue
		}
		path := strings.TrimSpace(service.Path)
		if path == "" {
			path = strings.TrimSpace(service.Name)
		}
		if path == "" {
			continue
		}
		if len(document.Services) == 1 {
			mergeScalarMetadata(contract.Metadata, service.Params)
			setMetadataDefault(contract.Metadata, "interface", path)
			setMetadataDefault(contract.Metadata, "dubbo.snapshot-service-key", key)
		}
		serviceJSON, _ := json.Marshal(service)
		appendDubboMethodNames(contract, path, scalarString(service.Params["methods"]), string(serviceJSON))
	}
	if len(contract.Interfaces) == 0 {
		path := scalarString(raw["interface"])
		methods := scalarString(raw["methods"])
		if path != "" && methods != "" {
			mergeScalarMetadata(contract.Metadata, raw)
			appendDubboMethodNames(contract, path, methods, contract.GetContent())
		}
	}

	sort.Slice(contract.Interfaces, func(i, j int) bool {
		if contract.Interfaces[i].GetPath() != contract.Interfaces[j].GetPath() {
			return contract.Interfaces[i].GetPath() < contract.Interfaces[j].GetPath()
		}
		if contract.Interfaces[i].GetMethod() != contract.Interfaces[j].GetMethod() {
			return contract.Interfaces[i].GetMethod() < contract.Interfaces[j].GetMethod()
		}
		return contract.Interfaces[i].GetType() < contract.Interfaces[j].GetType()
	})
}

func appendDubboMethodNames(contract *apiservice.ServiceContract, path, methods, content string) {
	for _, method := range strings.Split(methods, ",") {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}
		contract.Interfaces = append(contract.Interfaces, &apiservice.InterfaceDescriptor{
			Path: path, Method: method, Type: method, Content: content,
		})
	}
}

func mergeScalarMetadata(target map[string]string, source map[string]any) {
	for key, value := range source {
		if scalar := scalarString(value); scalar != "" {
			setMetadataDefault(target, key, scalar)
		}
	}
}

func scalarString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64, bool:
		return fmt.Sprint(typed)
	default:
		return ""
	}
}

func setMetadataDefault(metadata map[string]string, key, value string) {
	if strings.TrimSpace(metadata[key]) == "" && strings.TrimSpace(value) != "" {
		metadata[key] = strings.TrimSpace(value)
	}
}

func projectDubboMetadata(contract *apiservice.ServiceContract) {
	if contract.Metadata == nil {
		contract.Metadata = map[string]string{}
	}
	aliases := []struct {
		source string
		target string
	}{
		{source: "application", target: "dubbo.application"},
		{source: "app", target: "dubbo.application"},
		{source: "interface", target: "dubbo.interface"},
		{source: "group", target: "dubbo.group"},
		{source: "version", target: "dubbo.version"},
		{source: "side", target: "dubbo.side"},
		{source: "metadata-type", target: "dubbo.metadata-type"},
		{source: "metadata-revision", target: "dubbo.metadata-revision"},
		{source: "serialization", target: "dubbo.serialization"},
		{source: "service-key", target: "dubbo.service-key"},
		{source: "mapping-applications", target: "dubbo.mapping-applications"},
		{source: "applications", target: "dubbo.mapping-applications"},
	}
	for _, alias := range aliases {
		if strings.TrimSpace(contract.Metadata[alias.target]) != "" {
			continue
		}
		if value := strings.TrimSpace(contract.Metadata[alias.source]); value != "" {
			contract.Metadata[alias.target] = value
		}
	}

	if contract.Metadata["dubbo.interface"] == "" && len(contract.Interfaces) > 0 {
		path := strings.TrimSpace(contract.Interfaces[0].GetPath())
		sameInterface := path != ""
		for _, descriptor := range contract.Interfaces[1:] {
			if descriptor == nil || strings.TrimSpace(descriptor.GetPath()) != path {
				sameInterface = false
				break
			}
		}
		if sameInterface {
			contract.Metadata["dubbo.interface"] = path
		}
	}
	if contract.Metadata["dubbo.service-key"] == "" {
		serviceKey := contract.Metadata["dubbo.interface"]
		if group := contract.Metadata["dubbo.group"]; group != "" {
			serviceKey = group + "/" + serviceKey
		}
		if version := contract.Metadata["dubbo.version"]; version != "" {
			serviceKey += ":" + version
		}
		if serviceKey != "" {
			contract.Metadata["dubbo.service-key"] = serviceKey
		}
	}
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
		path    string
		method  string
		content string
	}
	operations := make([]operation, 0)
	for path, pathItem := range document.Paths {
		for method, definition := range pathItem {
			method = strings.ToLower(method)
			if !isHTTPMethod(method) {
				continue
			}
			content, _ := json.Marshal(definition)
			operations = append(operations, operation{
				path: path, method: strings.ToUpper(method), content: string(content),
			})
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
			Path:    operation.path,
			Method:  operation.method,
			Content: operation.content,
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
