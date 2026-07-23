package systemsettings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	store "github.com/pole-io/pole-server/console/pkg/observer"
)

const dataEncryptionKeySize = 32

type secretEnvelope struct {
	kek []byte
}

func newSecretEnvelope(encodedMasterKey string) (*secretEnvelope, error) {
	if encodedMasterKey == "" {
		return nil, errors.New("Pole system secret master key is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(encodedMasterKey)
	if err != nil || len(key) != dataEncryptionKeySize {
		return nil, errors.New("Pole system secret master key must be a base64 encoded 32-byte key")
	}
	return &secretEnvelope{kek: key}, nil
}

func (e *secretEnvelope) encrypt(component, domain, purpose, actor string, plaintext []byte) (*store.EncryptedSecretVersion, error) {
	dek := make([]byte, dataEncryptionKeySize)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generate data encryption key: %w", err)
	}
	dataNonce, ciphertext, err := seal(dek, plaintext, []byte(component+"/"+domain+"/"+purpose))
	if err != nil {
		return nil, err
	}
	wrapNonce, wrappedDEK, err := seal(e.kek, dek, []byte("pole-system-secret/dek"))
	if err != nil {
		return nil, err
	}
	fingerprint := sha256.Sum256(plaintext)
	return &store.EncryptedSecretVersion{
		Component: component, Domain: domain, Purpose: purpose,
		Ciphertext: ciphertext, DataNonce: dataNonce, WrappedDEK: wrappedDEK, WrapNonce: wrapNonce,
		Fingerprint: base64.RawURLEncoding.EncodeToString(fingerprint[:12]), CreatedBy: actor,
	}, nil
}

func (e *secretEnvelope) decrypt(version *store.EncryptedSecretVersion) ([]byte, error) {
	dek, err := open(e.kek, version.WrapNonce, version.WrappedDEK, []byte("pole-system-secret/dek"))
	if err != nil {
		return nil, errors.New("unwrap Pole system secret failed")
	}
	plaintext, err := open(dek, version.DataNonce, version.Ciphertext,
		[]byte(version.Component+"/"+version.Domain+"/"+version.Purpose))
	if err != nil {
		return nil, errors.New("decrypt Pole system secret failed")
	}
	return plaintext, nil
}

func seal(key, plaintext, aad []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return nonce, gcm.Seal(nil, nonce, plaintext, aad), nil
}

func open(key, nonce, ciphertext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, aad)
}
