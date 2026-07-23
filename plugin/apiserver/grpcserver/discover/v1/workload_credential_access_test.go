package v1

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/workloadcredential"
)

type credentialResolverStub struct {
	token string
}

func TestDisabledWorkloadCredentialRPCFailsClosed(t *testing.T) {
	server := &DiscoverGRPCServer{}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "service-token"))

	resp, err := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})

	require.NoError(t, err)
	assert.Equal(t, uint32(apimodel.Code_WorkloadCredentialIssuerUnavailable), resp.GetCode())
	assert.Nil(t, resp.GetCredential())
}

func (r *credentialResolverStub) GetOrCreateServiceIdentityByToken(token string) (*svctypes.Service, error) {
	r.token = token
	if token == "" {
		return nil, nil
	}
	return &svctypes.Service{
		ID: "service-id", Namespace: "default", Name: "orders",
		Identity: &svctypes.ServiceIdentity{
			ServiceID: "service-id", Subject: "pole://service/identity-id", Revision: "identity-revision",
		},
	}, nil
}

func TestIssueUsesAuthorizationMetadataOverTLS(t *testing.T) {
	resolver := &credentialResolverStub{}
	server := &DiscoverGRPCServer{workloadCredentialServer: newCredentialServer(t, resolver)}
	ctx := secureIncomingContext("service-token")

	resp, err := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})

	require.NoError(t, err)
	assert.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	assert.Equal(t, "service-token", resolver.token)
}

func TestEnabledWorkloadCredentialRPCAcceptsPlaintextWhenDeploymentChoosesIt(t *testing.T) {
	resolver := &credentialResolverStub{}
	server := &DiscoverGRPCServer{workloadCredentialServer: newCredentialServer(t, resolver)}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "service-token"))

	resp, err := server.Issue(ctx, &apisecurity.WorkloadCredentialIssueRequest{ProtocolVersion: 1})

	require.NoError(t, err)
	assert.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	assert.Equal(t, "service-token", resolver.token)
}

func TestServiceIdentityBundleAllowsNilServiceAndDoesNotUseBodyToken(t *testing.T) {
	resolver := &credentialResolverStub{}
	server := &DiscoverGRPCServer{workloadCredentialServer: newCredentialServer(t, resolver)}
	// handleDiscoverRequest receives the already converted stream context.
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "metadata-token"))
	ctx = utils.ConvertGRPCContext(ctx)

	resp := server.handleDiscoverRequest(ctx, &apiservice.DiscoverRequest{
		Type: apiservice.DiscoverRequest_SERVICE_IDENTITY_BUNDLE,
		// The bundle request deliberately has no Service message.
	})

	assert.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	assert.Equal(t, apiservice.DiscoverResponse_SERVICE_IDENTITY_BUNDLE, resp.GetType())
	assert.NotNil(t, resp.GetServiceIdentityBundle())
	assert.Equal(t, "metadata-token", resolver.token)
}

func TestServiceIdentityBundleDoesNotAllowBodyTokenOverride(t *testing.T) {
	resolver := &credentialResolverStub{}
	server := &DiscoverGRPCServer{workloadCredentialServer: newCredentialServer(t, resolver)}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "metadata-token"))
	ctx = utils.ConvertGRPCContext(ctx)

	resp := server.handleDiscoverRequest(ctx, &apiservice.DiscoverRequest{
		Type:    apiservice.DiscoverRequest_SERVICE_IDENTITY_BUNDLE,
		Service: &apiservice.Service{Token: "body-token"},
	})

	assert.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	assert.Equal(t, "metadata-token", resolver.token)
}

func newCredentialServer(t *testing.T, resolver workloadcredential.ServiceIdentityResolver) *workloadcredential.Server {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "active.pem")
	require.NoError(t, os.WriteFile(path,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), 0o600))
	server, err := workloadcredential.NewServer(workloadcredential.Config{
		Enabled: true, Issuer: "pole-control-plane", Audience: "pole-data-plane",
		TrustDomain: "pole.local", BundleSequence: 1,
		Keys: []workloadcredential.KeyConfig{{
			ID: "active-v1", State: workloadcredential.KeyStateActive, PrivateKeyFile: path,
		}},
	}, resolver, func() time.Time { return time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC) })
	require.NoError(t, err)
	return server
}

func secureIncomingContext(token string) context.Context {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", token))
	return peer.NewContext(ctx, &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}})
}
