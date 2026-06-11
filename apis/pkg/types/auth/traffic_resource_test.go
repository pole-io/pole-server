package auth

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
)

func TestTrafficGovernanceResourceFieldMappings(t *testing.T) {
	require.Equal(t, apisecurity.ResourceType_MirrorRules, ResourceFieldNames["mirror_rules"])
	require.Equal(t, apisecurity.ResourceType_SecurityRules, ResourceFieldNames["security_rules"])
	require.Equal(t, apisecurity.ResourceType_MockRules, ResourceFieldNames["mock_rules"])

	resources := &apisecurity.StrategyResources{
		MirrorRules:   []*apisecurity.StrategyResourceEntry{},
		SecurityRules: []*apisecurity.StrategyResourceEntry{},
		MockRules:     []*apisecurity.StrategyResourceEntry{},
	}
	for _, typ := range []apisecurity.ResourceType{
		apisecurity.ResourceType_MirrorRules,
		apisecurity.ResourceType_SecurityRules,
		apisecurity.ResourceType_MockRules,
	} {
		getter := ResourceFieldPointerGetters[typ]
		require.NotNil(t, getter)
		ptr := getter(resources)
		require.Equal(t, reflect.Ptr, ptr.Kind())
		ptr.Elem().Set(reflect.Append(ptr.Elem(), reflect.ValueOf(&apisecurity.StrategyResourceEntry{Id: "rule-1"})))
	}

	require.Equal(t, "rule-1", resources.GetMirrorRules()[0].GetId())
	require.Equal(t, "rule-1", resources.GetSecurityRules()[0].GetId())
	require.Equal(t, "rule-1", resources.GetMockRules()[0].GetId())
}
