package skillmarketplace

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"time"
)

func SignaturePayload(publisher, name, version, digest string, signedAt time.Time) []byte {
	return []byte(strings.Join([]string{
		publisher + "/" + name,
		version,
		digest,
		signedAt.UTC().Format(time.RFC3339),
	}, "\n"))
}

func VerifyDetachedSignature(publicKey, signature []byte, publisher, name, version, digest string, signedAt time.Time) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("publisher Ed25519 public key must be %d bytes", ed25519.PublicKeySize)
	}
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("detached Ed25519 signature must be %d bytes", ed25519.SignatureSize)
	}
	if signedAt.IsZero() {
		return fmt.Errorf("signedAt is required")
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), SignaturePayload(publisher, name, version, digest, signedAt), signature) {
		return fmt.Errorf("invalid detached Ed25519 signature")
	}
	return nil
}
