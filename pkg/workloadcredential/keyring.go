package workloadcredential

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

type signingKey struct {
	id      string
	state   KeyState
	private ed25519.PrivateKey
	public  ed25519.PublicKey
}

type keyRing struct {
	active *signingKey
	keys   map[string]*signingKey
}

func loadKeyRing(cfg Config) (*keyRing, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("workload credential is disabled")
	}
	ring := &keyRing{keys: make(map[string]*signingKey, len(cfg.Keys))}
	activeCount := 0
	for _, item := range cfg.Keys {
		if strings.TrimSpace(item.ID) == "" {
			return nil, fmt.Errorf("workload credential signing key id is empty")
		}
		if _, exists := ring.keys[item.ID]; exists {
			return nil, fmt.Errorf("duplicate workload credential signing key id %q", item.ID)
		}
		key := &signingKey{id: item.ID, state: item.State}
		switch item.State {
		case KeyStateActive:
			activeCount++
			privateKey, err := readPrivateKey(item.PrivateKeyFile)
			if err != nil {
				return nil, fmt.Errorf("load ACTIVE signing key %q: %w", item.ID, err)
			}
			key.private = privateKey
			key.public = privateKey.Public().(ed25519.PublicKey)
			ring.active = key
		case KeyStateVerifyOnly:
			publicKey, err := readPublicKey(item.PublicKeyFile)
			if err != nil {
				return nil, fmt.Errorf("load VERIFY_ONLY signing key %q: %w", item.ID, err)
			}
			key.public = publicKey
		default:
			return nil, fmt.Errorf("workload credential signing key %q has invalid state %q", item.ID, item.State)
		}
		ring.keys[item.ID] = key
	}
	if activeCount != 1 {
		return nil, fmt.Errorf("workload credential keyring requires exactly one ACTIVE key, got %d", activeCount)
	}
	return ring, nil
}

func readPrivateKey(path string) (ed25519.PrivateKey, error) {
	if path == "" {
		return nil, fmt.Errorf("private key file is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat private key file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("private key file is not a regular file")
	}
	// Kubernetes Secret volumes are commonly mounted 0444. Reject writes by
	// group/other while allowing those read-only mounts.
	if info.Mode().Perm()&0o022 != 0 {
		return nil, fmt.Errorf("private key file permissions %04o allow group/other writes", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("private key file is not PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not Ed25519")
	}
	return key, nil
}

func readPublicKey(path string) (ed25519.PublicKey, error) {
	if path == "" {
		return nil, fmt.Errorf("public key file is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("public key file is not PEM")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKIX public key: %w", err)
	}
	key, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not Ed25519")
	}
	return key, nil
}
