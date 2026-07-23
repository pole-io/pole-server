package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"

	"github.com/pole-io/pole-server/apis/pkg/types"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

type serviceIdentityStoreStub struct {
	storeapi.Store
	service *svctypes.Service
	err     error
	token   string
}

func TestCreateServiceModelAssignsStableInternalIdentity(t *testing.T) {
	server := &Server{}
	model := server.createServiceModel(&apiservice.Service{Name: "orders", Namespace: "default"})

	require.NotNil(t, model.Identity)
	assert.Equal(t, model.ID, model.Identity.ServiceID)
	assert.Contains(t, model.Identity.Subject, "pole://service/")
	assert.NotEmpty(t, model.Identity.Revision)
	assert.NotContains(t, model.ToSpec().GetMetadata(), "identity")
}

func (s *serviceIdentityStoreStub) GetOrCreateServiceIdentityByToken(token string) (*svctypes.Service, error) {
	s.token = token
	return s.service, s.err
}

func TestGetServiceIdentityRequiresMetadataToken(t *testing.T) {
	server := &Server{storage: &serviceIdentityStoreStub{}}

	resp := server.GetServiceIdentity(context.Background(), &apiservice.Service{
		Name:      "orders",
		Namespace: "default",
		Token:     "body-token-must-not-be-used",
	})

	assert.Equal(t, uint32(apimodel.Code_EmptyAutToken), resp.GetCode())
	assert.Equal(t, apiservice.DiscoverResponse_SERVICE_IDENTITY, resp.GetType())
	assert.Nil(t, resp.GetServiceIdentity())
}

func TestGetServiceIdentityReturnsIdentityResolvedFromToken(t *testing.T) {
	storage := &serviceIdentityStoreStub{service: &svctypes.Service{
		ID:        "service-id",
		Name:      "orders",
		Namespace: "default",
		Identity: &svctypes.ServiceIdentity{
			ServiceID: "service-id",
			Subject:   "pole://service/identity-id",
			Revision:  "identity-revision",
		},
	}}
	server := &Server{storage: storage}

	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "metadata-token")
	resp := server.GetServiceIdentity(ctx, &apiservice.Service{Name: "orders", Namespace: "default"})

	require.Equal(t, uint32(apimodel.Code_ExecuteSuccess), resp.GetCode())
	require.NotNil(t, resp.GetServiceIdentity())
	assert.Equal(t, "pole://service/identity-id", resp.GetServiceIdentity().GetSubject())
	assert.Equal(t, "orders", resp.GetServiceIdentity().GetService())
	assert.Equal(t, "default", resp.GetServiceIdentity().GetNamespace())
	assert.Equal(t, "identity-revision", resp.GetServiceIdentity().GetRevision())
	assert.Empty(t, resp.GetServiceIdentity().GetCredentialMode())
	assert.Empty(t, resp.GetServiceIdentity().GetCredentialEndpoint())
	assert.Equal(t, "metadata-token", storage.token)
}

func TestGetServiceIdentityRejectsClaimedServiceMismatch(t *testing.T) {
	storage := &serviceIdentityStoreStub{service: &svctypes.Service{
		Name:      "orders",
		Namespace: "default",
		Identity:  &svctypes.ServiceIdentity{Subject: "pole://service/id", Revision: "rev"},
	}}
	server := &Server{storage: storage}

	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "metadata-token")
	resp := server.GetServiceIdentity(ctx, &apiservice.Service{Name: "payments", Namespace: "default"})

	assert.Equal(t, uint32(apimodel.Code_NotAllowedAccess), resp.GetCode())
	assert.Nil(t, resp.GetServiceIdentity())
}

func TestGetServiceIdentityReturnsDataNoChangeByIdentityRevision(t *testing.T) {
	storage := &serviceIdentityStoreStub{service: &svctypes.Service{
		Name:      "orders",
		Namespace: "default",
		Identity:  &svctypes.ServiceIdentity{Subject: "pole://service/id", Revision: "rev"},
	}}
	server := &Server{storage: storage}

	ctx := context.WithValue(context.Background(), types.ContextAuthTokenKey, "metadata-token")
	resp := server.GetServiceIdentity(ctx, &apiservice.Service{
		Name: "orders", Namespace: "default", Revision: "rev",
	})

	assert.Equal(t, uint32(apimodel.Code_DataNoChange), resp.GetCode())
	assert.Nil(t, resp.GetServiceIdentity())
}
