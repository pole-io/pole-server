package rules

import (
	"testing"

	apifault "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCircuitBreakerBlockConfigAcceptsRegexSeparate(t *testing.T) {
	var rule apifault.CircuitBreakerRule
	payload := []byte(`{
		"name": "cb-regex-separate",
		"block_configs": [{
			"block_config": {
				"name": "api-regex",
				"regex_separate": true,
				"apis": [{
					"protocol": "HTTP",
					"method": "GET",
					"path": { "type": "REGEX", "value": "/v1/items/.*" }
				}],
				"error_conditions": [{
					"input_type": "RET_CODE",
					"condition": { "type": "EXACT", "value": "500" }
				}],
				"trigger_conditions": [{
					"trigger_type": "ERROR_RATE",
					"error_percent": 50,
					"interval": 60,
					"minimum_request": 10
				}]
			}
		}]
	}`)

	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(payload, &rule); err != nil {
		t.Fatalf("unmarshal circuit breaker with block_config.regex_separate: %v", err)
	}
	blockConfig := rule.GetBlockConfigs()[0].GetBlockConfig()
	field := blockConfig.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name("regex_separate"))
	if field == nil {
		t.Fatal("BlockConfig descriptor does not expose regex_separate")
	}
	if got := blockConfig.ProtoReflect().Get(field).Bool(); !got {
		t.Fatal("regex_separate = false, want true")
	}
}
