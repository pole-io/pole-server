package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"

	apicrypto "github.com/pole-io/pole-server/apis/crypto"
	cryptoaes "github.com/pole-io/pole-server/plugin/crypto/aes"
)

type templateValueCryptoManager struct {
	cryptor apicrypto.Crypto
}

func (m *templateValueCryptoManager) Name() string                 { return "test-template-value-crypto" }
func (m *templateValueCryptoManager) Initialize() error            { return nil }
func (m *templateValueCryptoManager) Destroy() error               { return nil }
func (m *templateValueCryptoManager) GetCryptoAlgoNames() []string { return []string{"AES"} }
func (m *templateValueCryptoManager) GetCrypto(name string) (apicrypto.Crypto, error) {
	return m.cryptor, nil
}

func TestTemplateSensitiveValueIsEncryptedAtRestAndDecryptedForUse(t *testing.T) {
	server := &Server{cryptoManager: &templateValueCryptoManager{cryptor: &cryptoaes.AESCrypto{}}}
	values := map[string]*apiconfig.ConfigTemplateValue{
		"database.password": {
			Value: &apiconfig.ConfigTemplateValue_StringValue{StringValue: "secret-value"},
		},
		"database.host": {
			Value: &apiconfig.ConfigTemplateValue_StringValue{StringValue: "db.internal"},
		},
	}
	schema := []*apiconfig.ConfigTemplateParameterSchema{
		{Name: "database.password", Sensitive: true},
		{Name: "database.host"},
	}

	payload, err := server.encodeTemplateValuesForStorage(values, schema)
	require.NoError(t, err)
	require.NotContains(t, payload, "secret-value")
	require.Contains(t, payload, "db.internal")
	require.Contains(t, payload, `"encrypted":true`)
	require.Contains(t, payload, `"algorithm":"AES"`)

	decoded, err := server.decodeTemplateValuesFromStorage(payload)
	require.NoError(t, err)
	require.Equal(t, "secret-value", decoded["database.password"].GetStringValue())
	require.Equal(t, "db.internal", decoded["database.host"].GetStringValue())
	require.False(t, strings.Contains(payload, decoded["database.password"].GetStringValue()))
}
