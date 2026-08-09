package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
)

const templateValueEncryptionAlgorithm = "AES"

type persistedEncryptedTemplateValue struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	DataKey   string `json:"data_key,omitempty"`
}

func (s *Server) encodeTemplateValuesForStorage(values map[string]*apiconfig.ConfigTemplateValue,
	schema []*apiconfig.ConfigTemplateParameterSchema) (string, error) {
	plain, err := conftypes.EncodeTemplateValues(values)
	if err != nil {
		return "", err
	}
	sensitive := make(map[string]bool, len(schema))
	for _, parameter := range schema {
		if parameter != nil && parameter.GetSensitive() {
			sensitive[parameter.GetName()] = true
		}
	}
	if len(sensitive) == 0 {
		return plain, nil
	}
	if s.cryptoManager == nil {
		return "", fmt.Errorf("template Value encryption manager is unavailable")
	}
	cryptor, err := s.cryptoManager.GetCrypto(templateValueEncryptionAlgorithm)
	if err != nil {
		return "", err
	}
	persisted := map[string]persistedEncryptedTemplateValue{}
	if err := json.Unmarshal([]byte(plain), &persisted); err != nil {
		return "", err
	}
	for name, item := range persisted {
		if !sensitive[name] {
			continue
		}
		key, err := cryptor.GenerateKey()
		if err != nil {
			return "", fmt.Errorf("generate encryption key for template Value %q: %w", name, err)
		}
		ciphertext, err := cryptor.Encrypt(item.Value, key)
		if err != nil {
			return "", fmt.Errorf("encrypt template Value %q: %w", name, err)
		}
		item.Value = ciphertext
		item.Encrypted = true
		item.Algorithm = templateValueEncryptionAlgorithm
		item.DataKey = base64.StdEncoding.EncodeToString(key)
		persisted[name] = item
	}
	encoded, err := json.Marshal(persisted)
	return string(encoded), err
}

func (s *Server) decodeTemplateValuesFromStorage(payload string) (map[string]*apiconfig.ConfigTemplateValue, error) {
	if payload == "" {
		return map[string]*apiconfig.ConfigTemplateValue{}, nil
	}
	persisted := map[string]persistedEncryptedTemplateValue{}
	if err := json.Unmarshal([]byte(payload), &persisted); err != nil {
		return nil, err
	}
	for name, item := range persisted {
		if !item.Encrypted {
			continue
		}
		if s.cryptoManager == nil {
			return nil, fmt.Errorf("template Value encryption manager is unavailable")
		}
		cryptor, err := s.cryptoManager.GetCrypto(item.Algorithm)
		if err != nil {
			return nil, fmt.Errorf("resolve encryption algorithm for template Value %q: %w", name, err)
		}
		key, err := base64.StdEncoding.DecodeString(item.DataKey)
		if err != nil {
			return nil, fmt.Errorf("decode encryption key for template Value %q: %w", name, err)
		}
		plaintext, err := cryptor.Decrypt(item.Value, key)
		if err != nil {
			return nil, fmt.Errorf("decrypt template Value %q: %w", name, err)
		}
		item.Value, item.Encrypted, item.Algorithm, item.DataKey = plaintext, false, "", ""
		persisted[name] = item
	}
	plain, err := json.Marshal(persisted)
	if err != nil {
		return nil, err
	}
	return conftypes.DecodeTemplateValues(string(plain))
}

func (s *Server) templateParameterSchema(namespace string,
	templateID uint64) ([]*apiconfig.ConfigTemplateParameterSchema, error) {
	item, err := s.getOrInitializeNamespaceConfigTemplateDraft(namespace, templateID)
	if err != nil {
		return nil, err
	}
	return conftypes.DecodeTemplateParameterSchema(item.ParameterSchema)
}
