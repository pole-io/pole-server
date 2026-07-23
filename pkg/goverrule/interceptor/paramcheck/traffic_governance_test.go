package paramcheck

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func TestCustomHeaderAuthenticationHashesAndClearsPlaintext(t *testing.T) {
	rule := &apisecurity.TrafficSecurityRule{
		Authentication: &apisecurity.TrafficSecurityAuthentication{
			Mode: apisecurity.TrafficSecurityAuthMode_CUSTOM_HEADER,
			CustomHeader: &apisecurity.CustomHeaderAuthentication{
				HeaderName:  " x-api-key ",
				Value:       "plain-secret",
				ValueSha256: "caller-controlled",
			},
		},
		Policies: []*apisecurity.TrafficSecurityPolicy{{
			Action: apisecurity.TrafficSecurityAction_TRAFFIC_SECURITY_ALLOW,
		}},
	}

	rsp := validateAndNormalizeTrafficSecurityRule(rule)

	assert.Nil(t, rsp)
	custom := rule.GetAuthentication().GetCustomHeader()
	assert.Equal(t, "x-api-key", custom.GetHeaderName())
	assert.Empty(t, custom.GetValue())
	digest := sha256.Sum256([]byte("plain-secret"))
	assert.Equal(t, hex.EncodeToString(digest[:]), custom.GetValueSha256())
}

func TestManagedIdentityAuthenticationRejectsAmbiguousCallerSelector(t *testing.T) {
	rule := &apisecurity.TrafficSecurityRule{
		Authentication: &apisecurity.TrafficSecurityAuthentication{
			Mode:            apisecurity.TrafficSecurityAuthMode_MANAGED_IDENTITY,
			ManagedIdentity: &apisecurity.ManagedIdentityAuthentication{},
		},
		Policies: []*apisecurity.TrafficSecurityPolicy{{
			ManagedCaller: &apisecurity.ManagedCallerSelector{
				AnyAuthenticated: true,
				Callers: []*apitraffic.SourceService{{
					Namespace: "default",
					Service:   "orders",
				}},
			},
		}},
	}

	rsp := validateAndNormalizeTrafficSecurityRule(rule)

	require.NotNil(t, rsp)
	assert.Equal(t, uint32(apimodel.Code_InvalidParameter), rsp.GetCode())
}

func TestLegacyTrafficSecurityRuleWithoutAuthenticationRemainsValid(t *testing.T) {
	assert.Nil(t, validateAndNormalizeTrafficSecurityRule(&apisecurity.TrafficSecurityRule{}))
}

func TestAuthenticationModesRejectFieldsOwnedByAnotherMode(t *testing.T) {
	tests := []struct {
		name string
		rule *apisecurity.TrafficSecurityRule
	}{
		{
			name: "legacy managed caller",
			rule: &apisecurity.TrafficSecurityRule{
				Authentication: &apisecurity.TrafficSecurityAuthentication{Mode: apisecurity.TrafficSecurityAuthMode_LEGACY_REQUEST_MATCH},
				Policies:       []*apisecurity.TrafficSecurityPolicy{{ManagedCaller: &apisecurity.ManagedCallerSelector{AnyAuthenticated: true}}},
			},
		},
		{
			name: "managed request matcher",
			rule: &apisecurity.TrafficSecurityRule{
				Authentication: &apisecurity.TrafficSecurityAuthentication{
					Mode:            apisecurity.TrafficSecurityAuthMode_MANAGED_IDENTITY,
					ManagedIdentity: &apisecurity.ManagedIdentityAuthentication{},
				},
				Policies: []*apisecurity.TrafficSecurityPolicy{{
					ManagedCaller:    &apisecurity.ManagedCallerSelector{AnyAuthenticated: true},
					TrafficMatchRule: &apitraffic.TrafficMatchRule{},
				}},
			},
		},
		{
			name: "custom request matcher",
			rule: &apisecurity.TrafficSecurityRule{
				Authentication: &apisecurity.TrafficSecurityAuthentication{
					Mode:         apisecurity.TrafficSecurityAuthMode_CUSTOM_HEADER,
					CustomHeader: &apisecurity.CustomHeaderAuthentication{HeaderName: "x-api-key", Value: "secret"},
				},
				Policies: []*apisecurity.TrafficSecurityPolicy{{TrafficMatchRule: &apitraffic.TrafficMatchRule{}}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rsp := validateAndNormalizeTrafficSecurityRule(test.rule)
			require.NotNil(t, rsp)
			assert.Equal(t, uint32(apimodel.Code_InvalidParameter), rsp.GetCode())
		})
	}
}
