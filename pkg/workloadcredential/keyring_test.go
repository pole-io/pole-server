package workloadcredential

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadKeyRingRejectsGroupWritablePrivateKey(t *testing.T) {
	dir := t.TempDir()
	privateKeyFile := writePrivateKeyFile(t, dir, 0o660)

	_, err := loadKeyRing(Config{
		Enabled: true,
		Keys: []KeyConfig{{
			ID:             "active-v1",
			State:          KeyStateActive,
			PrivateKeyFile: privateKeyFile,
		}},
	})

	require.ErrorContains(t, err, "permissions")
}

func TestLoadKeyRingRequiresExactlyOneActiveKey(t *testing.T) {
	dir := t.TempDir()
	first := writePrivateKeyFile(t, filepath.Join(dir, "first"), 0o600)
	second := writePrivateKeyFile(t, filepath.Join(dir, "second"), 0o600)

	_, err := loadKeyRing(Config{
		Enabled: true,
		Keys: []KeyConfig{
			{ID: "active-v1", State: KeyStateActive, PrivateKeyFile: first},
			{ID: "active-v2", State: KeyStateActive, PrivateKeyFile: second},
		},
	})

	require.ErrorContains(t, err, "exactly one ACTIVE")
}

func writePrivateKeyFile(t *testing.T, path string, mode os.FileMode) string {
	t.Helper()
	if filepath.Ext(path) == "" {
		path += ".pem"
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), mode))
	require.NoError(t, os.Chmod(path, mode))
	return path
}
