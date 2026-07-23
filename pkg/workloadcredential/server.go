package workloadcredential

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/protobuf/types/known/timestamppb"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	identityProtocolVersion uint32 = 1
	workloadJWTType                = "pole-workload+jwt"
	serviceTokenBinding            = "SERVICE_TOKEN"
	credentialEndpoint             = "v1.WorkloadCredentialService"
)

var errExpiredCredential = errors.New("workload credential expired")

type ServiceIdentityResolver interface {
	GetOrCreateServiceIdentityByToken(token string) (*svctypes.Service, error)
}

type Clock func() time.Time

// Server owns the complete signing and verification policy behind a small
// interface. It never accepts a service selector from a request body.
type Server struct {
	cfg           Config
	resolver      ServiceIdentityResolver
	keyring       *keyRing
	clock         Clock
	bundleVersion string
}

var (
	defaultServerMu sync.RWMutex
	defaultServer   *Server
)

func NewServer(cfg Config, resolver ServiceIdentityResolver, clock Clock) (*Server, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("workload credential issuer is disabled")
	}
	cfg = cfg.withDefaults()
	if resolver == nil {
		return nil, fmt.Errorf("workload credential service identity resolver is nil")
	}
	if cfg.Issuer == "" || cfg.Audience == "" || cfg.TrustDomain == "" {
		return nil, fmt.Errorf("workload credential issuer, audience and trustDomain are required")
	}
	if cfg.BundleSequence == 0 {
		return nil, fmt.Errorf("workload credential bundleSequence must be greater than zero")
	}
	if cfg.TTL <= 0 || cfg.ClockSkew < 0 || cfg.BundleTTL <= 0 {
		return nil, fmt.Errorf("workload credential ttl, clockSkew and bundleTTL are invalid")
	}
	ring, err := loadKeyRing(cfg)
	if err != nil {
		return nil, err
	}
	if clock == nil {
		clock = time.Now
	}
	return &Server{
		cfg:           cfg,
		resolver:      resolver,
		keyring:       ring,
		clock:         clock,
		bundleVersion: computeBundleVersion(cfg, ring),
	}, nil
}

func Initialize(cfg Config, resolver ServiceIdentityResolver) error {
	if !cfg.Enabled {
		defaultServerMu.Lock()
		defaultServer = nil
		defaultServerMu.Unlock()
		return nil
	}
	server, err := NewServer(cfg, resolver, nil)
	if err != nil {
		return err
	}
	defaultServerMu.Lock()
	defaultServer = server
	defaultServerMu.Unlock()
	return nil
}

func GetServer() *Server {
	defaultServerMu.RLock()
	defer defaultServerMu.RUnlock()
	return defaultServer
}

type credentialClaims struct {
	jwt.RegisteredClaims
	ProtocolVersion  uint32 `json:"pole_ver"`
	TrustDomain      string `json:"pole_trust_domain"`
	Namespace        string `json:"pole_namespace"`
	Service          string `json:"pole_service"`
	IdentityRevision string `json:"pole_identity_revision"`
	BindingType      string `json:"pole_binding_type"`
}

func (s *Server) Issue(ctx context.Context, req *apisecurity.WorkloadCredentialIssueRequest) *apisecurity.WorkloadCredentialResponse {
	if code := validateIssueRequest(req); code != apimodel.Code_ExecuteSuccess {
		return credentialResponse(code)
	}
	service, code := s.resolvePrincipal(ctx)
	if code != apimodel.Code_ExecuteSuccess {
		return credentialResponse(code)
	}
	if expected := req.GetExpectedIdentityRevision(); expected != "" && expected != s.DescriptorRevision(service.Identity.Revision) {
		return credentialResponse(apimodel.Code_StaleServiceIdentityRevision)
	}
	credential, err := s.sign(service)
	if err != nil {
		return credentialResponse(apimodel.Code_WorkloadCredentialIssuerUnavailable)
	}
	resp := credentialResponse(apimodel.Code_ExecuteSuccess)
	resp.Credential = credential
	return resp
}

func (s *Server) Renew(ctx context.Context, req *apisecurity.WorkloadCredentialRenewRequest) *apisecurity.WorkloadCredentialResponse {
	if code := validateRenewRequest(req); code != apimodel.Code_ExecuteSuccess {
		return credentialResponse(code)
	}
	service, code := s.resolvePrincipal(ctx)
	if code != apimodel.Code_ExecuteSuccess {
		return credentialResponse(code)
	}
	claims, err := s.verify(req.GetCurrentCredential())
	if errors.Is(err, errExpiredCredential) {
		return credentialResponse(apimodel.Code_ExpiredWorkloadCredential)
	}
	if err != nil {
		return credentialResponse(apimodel.Code_InvalidWorkloadCredential)
	}
	if claims.Subject != service.Identity.Subject ||
		claims.IdentityRevision != s.DescriptorRevision(service.Identity.Revision) ||
		claims.Namespace != service.Namespace || claims.Service != service.Name {
		return credentialResponse(apimodel.Code_WorkloadCredentialIssueForbidden)
	}
	credential, err := s.sign(service)
	if err != nil {
		return credentialResponse(apimodel.Code_WorkloadCredentialIssuerUnavailable)
	}
	resp := credentialResponse(apimodel.Code_ExecuteSuccess)
	resp.Credential = credential
	return resp
}

func (s *Server) resolvePrincipal(ctx context.Context) (*svctypes.Service, apimodel.Code) {
	token := utils.ParseAuthToken(ctx)
	if token == "" {
		return nil, apimodel.Code_EmptyAutToken
	}
	service, err := s.resolver.GetOrCreateServiceIdentityByToken(token)
	if err != nil {
		return nil, storeapi.StoreCode2APICode(err)
	}
	if service == nil || service.Identity == nil {
		return nil, apimodel.Code_TokenNotExisted
	}
	return service, apimodel.Code_ExecuteSuccess
}

func validateIssueRequest(req *apisecurity.WorkloadCredentialIssueRequest) apimodel.Code {
	if req == nil || req.GetProtocolVersion() != identityProtocolVersion || req.GetEvidence() != nil ||
		!supportsOnlyJWT(req.GetAcceptedFormats()) {
		return apimodel.Code_InvalidWorkloadCredentialRequest
	}
	return apimodel.Code_ExecuteSuccess
}

func validateRenewRequest(req *apisecurity.WorkloadCredentialRenewRequest) apimodel.Code {
	if req == nil || req.GetProtocolVersion() != identityProtocolVersion || req.GetCurrentCredential() == "" ||
		req.GetEvidence() != nil || !supportsOnlyJWT(req.GetAcceptedFormats()) {
		return apimodel.Code_InvalidWorkloadCredentialRequest
	}
	return apimodel.Code_ExecuteSuccess
}

func supportsOnlyJWT(formats []apisecurity.WorkloadCredentialFormat) bool {
	for _, format := range formats {
		if format != apisecurity.WorkloadCredentialFormat_WORKLOAD_CREDENTIAL_FORMAT_JWT_ED25519 {
			return false
		}
	}
	return true
}

func (s *Server) sign(service *svctypes.Service) (*apisecurity.WorkloadCredential, error) {
	now := s.clock().UTC()
	expiresAt := now.Add(s.cfg.TTL)
	notBefore := now.Add(-s.cfg.ClockSkew)
	id := utils.NewUUID()
	claims := &credentialClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.cfg.Issuer, Subject: service.Identity.Subject,
			Audience: jwt.ClaimStrings{s.cfg.Audience}, ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(notBefore), IssuedAt: jwt.NewNumericDate(now), ID: id,
		},
		ProtocolVersion: identityProtocolVersion, TrustDomain: s.cfg.TrustDomain,
		Namespace: service.Namespace, Service: service.Name,
		IdentityRevision: s.DescriptorRevision(service.Identity.Revision), BindingType: serviceTokenBinding,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["typ"] = workloadJWTType
	token.Header["kid"] = s.keyring.active.id
	serialized, err := token.SignedString(s.keyring.active.private)
	if err != nil {
		return nil, err
	}
	return &apisecurity.WorkloadCredential{
		Format:             apisecurity.WorkloadCredentialFormat_WORKLOAD_CREDENTIAL_FORMAT_JWT_ED25519,
		Serialized:         serialized,
		CredentialId:       id,
		KeyId:              s.keyring.active.id,
		TrustBundleVersion: s.bundleVersion,
		IssuedAt:           timestamppb.New(now),
		NotBefore:          timestamppb.New(notBefore),
		ExpiresAt:          timestamppb.New(expiresAt),
		RenewAfter:         timestamppb.New(now.Add(s.cfg.TTL * 2 / 3)),
		BindingType:        apisecurity.WorkloadBindingType_WORKLOAD_BINDING_SERVICE_TOKEN,
	}, nil
}

func (s *Server) verify(serialized string) (*credentialClaims, error) {
	claims := &credentialClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(serialized, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Header["typ"] != workloadJWTType {
			return nil, fmt.Errorf("invalid credential type")
		}
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, fmt.Errorf("missing credential key id")
		}
		key := s.keyring.keys[kid]
		if key == nil {
			return nil, fmt.Errorf("unknown credential key id")
		}
		return key.public, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("verify workload credential")
	}
	now := s.clock().UTC()
	if claims.ExpiresAt == nil || !now.Before(claims.ExpiresAt.Time) {
		return nil, errExpiredCredential
	}
	if claims.NotBefore == nil || claims.NotBefore.Time.After(now.Add(s.cfg.ClockSkew)) ||
		claims.IssuedAt == nil || claims.IssuedAt.Time.After(now.Add(s.cfg.ClockSkew)) ||
		claims.ID == "" || claims.Subject == "" || claims.ProtocolVersion != identityProtocolVersion ||
		claims.Issuer != s.cfg.Issuer || !claims.VerifyAudience(s.cfg.Audience, true) ||
		claims.TrustDomain != s.cfg.TrustDomain || claims.Namespace == "" || claims.Service == "" ||
		claims.IdentityRevision == "" || claims.BindingType != serviceTokenBinding {
		return nil, fmt.Errorf("invalid workload credential claims")
	}
	return claims, nil
}

func credentialResponse(code apimodel.Code) *apisecurity.WorkloadCredentialResponse {
	base := api.NewResponse(code)
	return &apisecurity.WorkloadCredentialResponse{Code: base.Code, Info: base.Info}
}

func computeBundleVersion(cfg Config, ring *keyRing) string {
	ids := make([]string, 0, len(ring.keys))
	for id := range ring.keys {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%d\x00", cfg.TrustDomain, cfg.Issuer, cfg.BundleSequence)
	for _, id := range ids {
		key := ring.keys[id]
		_, _ = hash.Write([]byte(id))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(key.public)
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(key.state))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Server) DescriptorFields() (mode, trustBundleVersion, trustDomain, audience, endpoint string, protocolVersion uint32,
	formats []apisecurity.WorkloadCredentialFormat) {
	return "JWT_ED25519", s.bundleVersion, s.cfg.TrustDomain, s.cfg.Audience, credentialEndpoint,
		identityProtocolVersion, []apisecurity.WorkloadCredentialFormat{
			apisecurity.WorkloadCredentialFormat_WORKLOAD_CREDENTIAL_FORMAT_JWT_ED25519,
		}
}

// DescriptorRevision changes whenever either the stable identity or the
// credential bootstrap descriptor changes. Trust-bundle rotation therefore
// cannot be hidden behind an identity-only DataNoChange response.
func (s *Server) DescriptorRevision(identityRevision string) string {
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\x00%s\x00%d", identityRevision, s.bundleVersion,
		s.cfg.Audience, credentialEndpoint, identityProtocolVersion)
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Server) TrustBundle() *apisecurity.WorkloadTrustBundle {
	now := s.clock().UTC()
	bundleExpiresAt := now.Add(s.cfg.BundleTTL)
	keyNotBefore := now.Add(-s.cfg.ClockSkew)
	ids := make([]string, 0, len(s.keyring.keys))
	for id := range s.keyring.keys {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	keys := make([]*apisecurity.WorkloadVerificationKey, 0, len(ids))
	for _, id := range ids {
		key := s.keyring.keys[id]
		state := apisecurity.VerificationKeyState_VERIFICATION_KEY_STATE_RETIRING
		if key.state == KeyStateActive {
			state = apisecurity.VerificationKeyState_VERIFICATION_KEY_STATE_ACTIVE
		}
		keys = append(keys, &apisecurity.WorkloadVerificationKey{
			KeyId:     id,
			Algorithm: apisecurity.WorkloadSigningAlgorithm_WORKLOAD_SIGNING_ALGORITHM_ED25519,
			PublicKey: append([]byte(nil), key.public...),
			State:     state,
			NotBefore: timestamppb.New(keyNotBefore),
			NotAfter:  timestamppb.New(bundleExpiresAt),
		})
	}
	return &apisecurity.WorkloadTrustBundle{
		SchemaVersion: 1,
		TrustDomain:   s.cfg.TrustDomain,
		Issuer:        s.cfg.Issuer,
		Version:       s.bundleVersion,
		Sequence:      s.cfg.BundleSequence,
		IssuedAt:      timestamppb.New(now),
		ExpiresAt:     timestamppb.New(bundleExpiresAt),
		Keys:          keys,
	}
}

func (s *Server) DiscoverTrustBundle(ctx context.Context, query *apisecurity.TrustBundleQuery) (
	*apisecurity.WorkloadTrustBundle, apimodel.Code) {
	if _, code := s.resolvePrincipal(ctx); code != apimodel.Code_ExecuteSuccess {
		return nil, code
	}
	if query != nil {
		if query.GetKnownSequence() > s.cfg.BundleSequence {
			return nil, apimodel.Code_InvalidWorkloadCredentialRequest
		}
		if query.GetKnownVersion() == s.bundleVersion && query.GetKnownSequence() == s.cfg.BundleSequence {
			return nil, apimodel.Code_DataNoChange
		}
	}
	return s.TrustBundle(), apimodel.Code_ExecuteSuccess
}
