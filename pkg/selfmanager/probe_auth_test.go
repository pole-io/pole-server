package selfmanager

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCapabilityProbeSignatureIsBoundAndShortLived(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	key, err := DeriveCapabilityProbeKey(masterKey)
	require.NoError(t, err)
	require.NotEqual(t, masterKey, key)
	now := time.Unix(1000, 0)
	body := []byte(`{"allowlist":["list_namespaces"]}`)
	timestamp, signature, err := SignCapabilityProbe(key, "POST",
		"/ai/mcp/v1/self-capabilities/probe", body, now)
	require.NoError(t, err)
	require.NoError(t, VerifyCapabilityProbe(key, "POST",
		"/ai/mcp/v1/self-capabilities/probe", body, timestamp, signature, now))
	require.Error(t, VerifyCapabilityProbe(key, "POST",
		"/ai/mcp/v1/self-capabilities/probe", []byte(`{}`), timestamp, signature, now))
	require.ErrorContains(t, VerifyCapabilityProbe(key, "POST",
		"/ai/mcp/v1/self-capabilities/probe", body, timestamp, signature,
		now.Add(2*time.Minute)), "expired")
}
