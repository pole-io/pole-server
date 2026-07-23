package paramcheck

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

func (s *Server) CreateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := validateAndNormalizeTrafficSecurityRules(req); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficSecurityRules(ctx, req)
}
func (s *Server) UpdateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := validateAndNormalizeTrafficSecurityRules(req); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficSecurityRules(ctx, req)
}
func (s *Server) DeleteTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficSecurityRules(ctx, req)
}
func (s *Server) GetTrafficSecurityRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficSecurityRules(ctx, query)
}
func (s *Server) GetOneTrafficSecurityRule(ctx context.Context, req *apisecurity.TrafficSecurityRule) *apimodel.Response {
	return s.nextSvr.GetOneTrafficSecurityRule(ctx, req)
}

func validateAndNormalizeTrafficSecurityRules(reqs []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if len(reqs) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}
	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range reqs {
		api.Collect(batchRsp, validateAndNormalizeTrafficSecurityRule(reqs[i]))
	}
	if !api.IsSuccess(batchRsp) {
		return batchRsp
	}
	return nil
}

func validateAndNormalizeTrafficSecurityRule(rule *apisecurity.TrafficSecurityRule) *apimodel.Response {
	if rule == nil {
		return api.NewResponse(apimodel.Code_EmptyRequest)
	}
	authn := rule.GetAuthentication()
	if authn == nil {
		// Existing rules predate explicit authentication modes.
		return nil
	}

	switch authn.GetMode() {
	case apisecurity.TrafficSecurityAuthMode_LEGACY_REQUEST_MATCH:
		if authn.GetManagedIdentity() != nil || authn.GetCustomHeader() != nil {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "legacy authentication cannot contain managed or custom-header configuration")
		}
		for _, policy := range rule.GetPolicies() {
			if policy.GetManagedCaller() != nil {
				return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "legacy authentication policy cannot contain managed callers")
			}
		}
	case apisecurity.TrafficSecurityAuthMode_MANAGED_IDENTITY:
		if authn.GetManagedIdentity() == nil || authn.GetCustomHeader() != nil {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "managed identity authentication has inconsistent configuration")
		}
		for _, policy := range rule.GetPolicies() {
			if policy.GetTrafficMatchRule() != nil {
				return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "managed identity policy cannot contain request match authentication")
			}
			selector := policy.GetManagedCaller()
			if selector == nil || (!selector.GetAnyAuthenticated() && len(selector.GetCallers()) == 0) ||
				(selector.GetAnyAuthenticated() && len(selector.GetCallers()) != 0) {
				return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "managed identity policy requires either any_authenticated or explicit callers")
			}
			for _, caller := range selector.GetCallers() {
				if strings.TrimSpace(caller.GetNamespace()) == "" || strings.TrimSpace(caller.GetService()) == "" {
					return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "managed identity caller requires namespace and service")
				}
			}
		}
	case apisecurity.TrafficSecurityAuthMode_CUSTOM_HEADER:
		custom := authn.GetCustomHeader()
		if custom == nil || authn.GetManagedIdentity() != nil || strings.TrimSpace(custom.GetHeaderName()) == "" || custom.GetValue() == "" {
			return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "custom header authentication requires header_name and a write-only value")
		}
		for _, policy := range rule.GetPolicies() {
			if policy.GetManagedCaller() != nil || policy.GetTrafficMatchRule() != nil {
				return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "custom header policy cannot contain managed callers or request match authentication")
			}
		}
		digest := sha256.Sum256([]byte(custom.GetValue()))
		custom.HeaderName = strings.TrimSpace(custom.GetHeaderName())
		custom.ValueSha256 = hex.EncodeToString(digest[:])
		custom.Value = ""
	default:
		return api.NewResponseWithMsg(apimodel.Code_InvalidParameter, "unsupported traffic security authentication mode")
	}
	return nil
}

func (s *Server) CreateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateTrafficMirrorRules(ctx, req)
}
func (s *Server) UpdateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateTrafficMirrorRules(ctx, req)
}
func (s *Server) DeleteTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficMirrorRules(ctx, req)
}
func (s *Server) GetTrafficMirrorRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMirrorRules(ctx, query)
}
func (s *Server) GetOneTrafficMirrorRule(ctx context.Context, req *apitraffic.TrafficMirror) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMirrorRule(ctx, req)
}

func (s *Server) CreateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.CreateTrafficMockRules(ctx, req)
}
func (s *Server) UpdateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.UpdateTrafficMockRules(ctx, req)
}
func (s *Server) DeleteTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return s.nextSvr.DeleteTrafficMockRules(ctx, req)
}
func (s *Server) GetTrafficMockRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	return s.nextSvr.GetTrafficMockRules(ctx, query)
}
func (s *Server) GetOneTrafficMockRule(ctx context.Context, req *apitraffic.TrafficMock) *apimodel.Response {
	return s.nextSvr.GetOneTrafficMockRule(ctx, req)
}
