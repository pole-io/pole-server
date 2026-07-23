package workloadcredential

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/pkg/common/utils"
)

func TestConfigParsesDurationAndExternalKeyReferences(t *testing.T) {
	var root struct {
		WorkloadCredential Config `yaml:"workloadCredential"`
	}
	err := utils.ParseYamlContent(`
workloadCredential:
  enabled: true
  issuer: pole-control-plane
  audience: pole-data-plane
  trustDomain: pole.local
  ttl: 5m
  clockSkew: 30s
  bundleSequence: 2
  bundleTTL: 10m
  keys:
    - id: active-v2
      state: ACTIVE
      privateKeyFile: /var/run/secrets/pole/active.pem
`, &root)

	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, root.WorkloadCredential.TTL)
	assert.Equal(t, 30*time.Second, root.WorkloadCredential.ClockSkew)
	assert.Equal(t, uint64(2), root.WorkloadCredential.BundleSequence)
	assert.Equal(t, "/var/run/secrets/pole/active.pem", root.WorkloadCredential.Keys[0].PrivateKeyFile)
}
