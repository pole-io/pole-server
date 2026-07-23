package workloadcredential

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
)

type resolverStub struct {
	token   string
	service *svctypes.Service
	err     error
}

func (r *resolverStub) GetOrCreateServiceIdentityByToken(token string) (*svctypes.Service, error) {
	r.token = token
	return r.service, r.err
}

func TestIssueDerivesPrincipalFromMetadataTokenAndSignsEdDSAJWT(t *testing.T) {
	now := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	resolver := &resolverStub{service: testServiceIdentity()}
	server := newTestServer(t, resolver, now)
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "service-token")

	resp := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	require.NotNil(t, resp.GetCredential())
	assert.Equal(t, "service-token", resolver.token)
	assert.Equal(t, apisecurity.WorkloadCredentialFormat_WORKLOAD_CREDENTIAL_FORMAT_JWT_ED25519, resp.Credential.Format)
	assert.Equal(t, apisecurity.WorkloadBindingType_WORKLOAD_BINDING_SERVICE_TOKEN, resp.Credential.BindingType)
	assert.Equal(t, "active-v1", resp.Credential.KeyId)

	claims := &credentialClaims{}
	token, _, err := new(jwt.Parser).ParseUnverified(resp.Credential.Serialized, claims)
	require.NoError(t, err)
	assert.Equal(t, "EdDSA", token.Header["alg"])
	assert.Equal(t, workloadJWTType, token.Header["typ"])
	assert.Equal(t, "active-v1", token.Header["kid"])
	assert.Equal(t, "pole://service/identity-id", claims.Subject)
	assert.Equal(t, "default", claims.Namespace)
	assert.Equal(t, "orders", claims.Service)
	assert.Equal(t, "SERVICE_TOKEN", claims.BindingType)
	assert.True(t, now.Add(DefaultCredentialTTL).Equal(claims.ExpiresAt.Time))
	assert.NotEmpty(t, claims.ID)
	assert.NotContains(t, resp.Credential.Serialized, "service-token")
}

func TestIssueRejectsEvidenceAndUnsupportedFormats(t *testing.T) {
	server := newTestServer(t, &resolverStub{service: testServiceIdentity()}, time.Now())
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "service-token")

	evidenceResp := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{
		ProtocolVersion: 1,
		Evidence: &apisecurity.WorkloadEvidence{
			Evidence: &apisecurity.WorkloadEvidence_CloudIdentityToken{CloudIdentityToken: "secret"},
		},
	})
	formatResp := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{
		ProtocolVersion: 1,
		AcceptedFormats: []apisecurity.WorkloadCredentialFormat{
			apisecurity.WorkloadCredentialFormat_WORKLOAD_CREDENTIAL_FORMAT_UNSPECIFIED,
		},
	})

	assert.Equal(t, uint32(apimodel.Code_InvalidWorkloadCredentialRequest), evidenceResp.GetCode())
	assert.Equal(t, uint32(apimodel.Code_InvalidWorkloadCredentialRequest), formatResp.GetCode())
}

func TestRenewRequiresCurrentUnexpiredCredentialForSameSubject(t *testing.T) {
	now := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	resolver := &resolverStub{service: testServiceIdentity()}
	server := newTestServer(t, resolver, now)
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "service-token")
	issued := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})
	require.NotNil(t, issued.GetCredential())
	issuedClaims, err := server.verify(issued.Credential.Serialized)
	require.NoError(t, err)
	require.Equal(t, resolver.service.Identity.Subject, issuedClaims.Subject)
	require.Equal(t, server.DescriptorRevision(resolver.service.Identity.Revision), issuedClaims.IdentityRevision)
	require.Equal(t, resolver.service.Namespace, issuedClaims.Namespace)
	require.Equal(t, resolver.service.Name, issuedClaims.Service)

	renewed := server.Renew(ctx, &apisecurity.WorkloadCredentialRenewRequest{
		ProtocolVersion:   1,
		CurrentCredential: issued.Credential.Serialized,
	})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), renewed.GetCode())
	require.NotNil(t, renewed.GetCredential())
	assert.NotEqual(t, issued.Credential.CredentialId, renewed.Credential.CredentialId)

	server.clock = func() time.Time { return now.Add(DefaultCredentialTTL) }
	expired := server.Renew(ctx, &apisecurity.WorkloadCredentialRenewRequest{
		ProtocolVersion:   1,
		CurrentCredential: issued.Credential.Serialized,
	})
	assert.Equal(t, uint32(apimodel.Code_ExpiredWorkloadCredential), expired.GetCode())

	resolver.service.Identity.Subject = "pole://service/other"
	server.clock = func() time.Time { return now.Add(time.Minute) }
	mismatch := server.Renew(ctx, &apisecurity.WorkloadCredentialRenewRequest{
		ProtocolVersion:   1,
		CurrentCredential: issued.Credential.Serialized,
	})
	assert.Equal(t, uint32(apimodel.Code_WorkloadCredentialIssueForbidden), mismatch.GetCode())

	resolver.service.Identity.Subject = claimsSubject(t, issued.Credential.Serialized)
	resolver.service.Identity.Revision = "identity-revision-2"
	revisionMismatch := server.Renew(ctx, &apisecurity.WorkloadCredentialRenewRequest{
		ProtocolVersion:   1,
		CurrentCredential: issued.Credential.Serialized,
	})
	assert.Equal(t, uint32(apimodel.Code_WorkloadCredentialIssueForbidden), revisionMismatch.GetCode())
}

func claimsSubject(t *testing.T, serialized string) string {
	t.Helper()
	claims := &credentialClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(serialized, claims)
	require.NoError(t, err)
	return claims.Subject
}

func TestTrustBundleContainsOnlyPublicActiveAndRetiringKeys(t *testing.T) {
	now := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	activePrivate, activePublic := writeTestPrivateKey(t, filepath.Join(dir, "active.pem"))
	_, retiringPrivate, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	retiringPublic := retiringPrivate.Public().(ed25519.PublicKey)
	retiringDER, err := x509.MarshalPKIXPublicKey(retiringPublic)
	require.NoError(t, err)
	retiringPath := filepath.Join(dir, "retiring.pem")
	require.NoError(t, os.WriteFile(retiringPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: retiringDER}), 0o644))

	server, err := NewServer(Config{
		Enabled: true, Issuer: "pole-control-plane", Audience: "pole-data-plane",
		TrustDomain: "pole.local", BundleSequence: 7,
		Keys: []KeyConfig{
			{ID: "active-v2", State: KeyStateActive, PrivateKeyFile: activePrivate},
			{ID: "retiring-v1", State: KeyStateVerifyOnly, PublicKeyFile: retiringPath},
		},
	}, &resolverStub{service: testServiceIdentity()}, func() time.Time { return now })
	require.NoError(t, err)

	bundle := server.TrustBundle()

	require.Len(t, bundle.GetKeys(), 2)
	assert.Equal(t, uint64(7), bundle.GetSequence())
	assert.Equal(t, "pole.local", bundle.GetTrustDomain())
	assert.Equal(t, "pole-control-plane", bundle.GetIssuer())
	assert.True(t, now.Add(DefaultBundleTTL).Equal(bundle.GetExpiresAt().AsTime()))
	keys := map[string]*apisecurity.WorkloadVerificationKey{}
	for _, key := range bundle.GetKeys() {
		keys[key.GetKeyId()] = key
	}
	assert.Equal(t, []byte(activePublic), keys["active-v2"].GetPublicKey())
	assert.Equal(t, apisecurity.VerificationKeyState_VERIFICATION_KEY_STATE_ACTIVE, keys["active-v2"].GetState())
	assert.True(t, now.Add(-DefaultClockSkew).Equal(keys["active-v2"].GetNotBefore().AsTime()))
	assert.True(t, now.Add(DefaultBundleTTL).Equal(keys["active-v2"].GetNotAfter().AsTime()))
	assert.Equal(t, []byte(retiringPublic), keys["retiring-v1"].GetPublicKey())
	assert.Equal(t, apisecurity.VerificationKeyState_VERIFICATION_KEY_STATE_RETIRING, keys["retiring-v1"].GetState())
}

func TestVerifyOnlyRotationKeyValidatesOldCredentialButRevisionChangeRequiresIssue(t *testing.T) {
	now := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	resolver := &resolverStub{service: testServiceIdentity()}
	dir := t.TempDir()
	oldPrivatePath, oldPublic := writeTestPrivateKey(t, filepath.Join(dir, "old-private.pem"))
	oldIssuer, err := NewServer(testConfig(oldPrivatePath, "old-v1"), resolver, func() time.Time { return now })
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "service-token")
	issued := oldIssuer.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})
	require.NotNil(t, issued.GetCredential())

	newPrivatePath, _ := writeTestPrivateKey(t, filepath.Join(dir, "new-private.pem"))
	oldPublicDER, err := x509.MarshalPKIXPublicKey(oldPublic)
	require.NoError(t, err)
	oldPublicPath := filepath.Join(dir, "old-public.pem")
	require.NoError(t, os.WriteFile(oldPublicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: oldPublicDER}), 0o644))
	cfg := testConfig(newPrivatePath, "new-v2")
	cfg.BundleSequence = 2
	cfg.Keys = append(cfg.Keys, KeyConfig{ID: "old-v1", State: KeyStateVerifyOnly, PublicKeyFile: oldPublicPath})
	newIssuer, err := NewServer(cfg, resolver, func() time.Time { return now.Add(time.Minute) })
	require.NoError(t, err)
	_, err = newIssuer.verify(issued.Credential.Serialized)
	require.NoError(t, err)

	renewed := newIssuer.Renew(ctx, &apisecurity.WorkloadCredentialRenewRequest{
		ProtocolVersion: 1, CurrentCredential: issued.Credential.Serialized,
	})

	require.Equal(t, uint32(apimodel.Code_WorkloadCredentialIssueForbidden), renewed.GetCode())
}

func testServiceIdentity() *svctypes.Service {
	return &svctypes.Service{
		ID: "service-id", Namespace: "default", Name: "orders",
		Identity: &svctypes.ServiceIdentity{
			ServiceID: "service-id", Subject: "pole://service/identity-id", Revision: "identity-revision",
		},
	}
}

func newTestServer(t *testing.T, resolver ServiceIdentityResolver, now time.Time) *Server {
	t.Helper()
	dir := t.TempDir()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	path := filepath.Join(dir, "active.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), 0o600))
	server, err := NewServer(Config{
		Enabled: true, Issuer: "pole-control-plane", Audience: "pole-data-plane",
		TrustDomain: "pole.local", BundleSequence: 1,
		Keys: []KeyConfig{{ID: "active-v1", State: KeyStateActive, PrivateKeyFile: path}},
	}, resolver, func() time.Time { return now })
	require.NoError(t, err)
	return server
}

func writeTestPrivateKey(t *testing.T, path string) (string, ed25519.PublicKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), 0o600))
	return path, publicKey
}

func testConfig(privateKeyPath, keyID string) Config {
	return Config{
		Enabled: true, Issuer: "pole-control-plane", Audience: "pole-data-plane",
		TrustDomain: "pole.local", BundleSequence: 1,
		Keys: []KeyConfig{{ID: keyID, State: KeyStateActive, PrivateKeyFile: privateKeyPath}},
	}
}
