package selfmanager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	ProbeTimestampHeader = "X-Pole-Self-Manager-Timestamp"
	ProbeSignatureHeader = "X-Pole-Self-Manager-Signature"
	probeKeySize         = 32
	probeClockSkew       = time.Minute
)

func ValidateCapabilityProbeKey(encodedProbeKey string) error {
	_, err := decodeProbeKey(encodedProbeKey)
	return err
}

func DeriveCapabilityProbeKey(encodedMasterKey string) (string, error) {
	masterKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedMasterKey))
	if err != nil || len(masterKey) != probeKeySize {
		return "", errors.New("Pole system secret master key must be a base64 encoded 32-byte key")
	}
	deriver := hmac.New(sha256.New, masterKey)
	_, _ = deriver.Write([]byte("pole-self-manager-capability-probe-key-v1"))
	return base64.StdEncoding.EncodeToString(deriver.Sum(nil)), nil
}

func ResolveCapabilityProbeKey(encodedProbeKey, encodedMasterKey string) (string, error) {
	if strings.TrimSpace(encodedProbeKey) != "" {
		if err := ValidateCapabilityProbeKey(encodedProbeKey); err != nil {
			return "", err
		}
		return strings.TrimSpace(encodedProbeKey), nil
	}
	return DeriveCapabilityProbeKey(encodedMasterKey)
}

func SignCapabilityProbe(encodedProbeKey, method, path string, body []byte,
	now time.Time) (string, string, error) {
	key, err := decodeProbeKey(encodedProbeKey)
	if err != nil {
		return "", "", err
	}
	timestamp := strconv.FormatInt(now.UTC().Unix(), 10)
	signature := capabilityProbeSignature(key, method, path, body, timestamp)
	return timestamp, base64.RawURLEncoding.EncodeToString(signature), nil
}

func VerifyCapabilityProbe(encodedProbeKey, method, path string, body []byte,
	timestamp, encodedSignature string, now time.Time) error {
	key, err := decodeProbeKey(encodedProbeKey)
	if err != nil {
		return err
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(timestamp), 10, 64)
	if err != nil {
		return errors.New("invalid probe timestamp")
	}
	signedAt := time.Unix(seconds, 0)
	if signedAt.Before(now.Add(-probeClockSkew)) || signedAt.After(now.Add(probeClockSkew)) {
		return errors.New("expired probe signature")
	}
	signature, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encodedSignature))
	if err != nil {
		return errors.New("invalid probe signature")
	}
	expected := capabilityProbeSignature(key, method, path, body, timestamp)
	if !hmac.Equal(signature, expected) {
		return errors.New("invalid probe signature")
	}
	return nil
}

func decodeProbeKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(key) != probeKeySize {
		return nil, errors.New("Pole self-manager probe key must be a base64 encoded 32-byte key")
	}
	return key, nil
}

func capabilityProbeSignature(key []byte, method, path string, body []byte, timestamp string) []byte {
	bodyHash := sha256.Sum256(body)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("pole-self-manager-capability-probe-v1\n"))
	_, _ = mac.Write([]byte(strings.ToUpper(method)))
	_, _ = mac.Write([]byte("\n"))
	_, _ = mac.Write([]byte(path))
	_, _ = mac.Write([]byte("\n"))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("\n"))
	_, _ = mac.Write(bodyHash[:])
	return mac.Sum(nil)
}
