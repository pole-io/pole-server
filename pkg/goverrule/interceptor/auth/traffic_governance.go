package goverrule_auth

import (
	"context"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
)

type trafficGovernanceAuthSpec struct {
	resource     apisecurity.ResourceType
	readMethod   authtypes.ServerFunctionName
	updateMethod authtypes.ServerFunctionName
	deleteMethod authtypes.ServerFunctionName
}

func trafficSecurityAuthSpec() trafficGovernanceAuthSpec {
	return trafficGovernanceAuthSpec{
		resource:     apisecurity.ResourceType_SecurityRules,
		readMethod:   authtypes.DescribeTrafficSecurityRules,
		updateMethod: authtypes.UpdateTrafficSecurityRules,
		deleteMethod: authtypes.DeleteTrafficSecurityRules,
	}
}

func trafficMirrorAuthSpec() trafficGovernanceAuthSpec {
	return trafficGovernanceAuthSpec{
		resource:     apisecurity.ResourceType_MirrorRules,
		readMethod:   authtypes.DescribeTrafficMirrorRules,
		updateMethod: authtypes.UpdateTrafficMirrorRules,
		deleteMethod: authtypes.DeleteTrafficMirrorRules,
	}
}

func trafficMockAuthSpec() trafficGovernanceAuthSpec {
	return trafficGovernanceAuthSpec{
		resource:     apisecurity.ResourceType_MockRules,
		readMethod:   authtypes.DescribeTrafficMockRules,
		updateMethod: authtypes.UpdateTrafficMockRules,
		deleteMethod: authtypes.DeleteTrafficMockRules,
	}
}

func (s *Server) CreateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficSecurityRules(ctx, req)
}
func (s *Server) UpdateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficSecurityRules(ctx, req)
}
func (s *Server) DeleteTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficSecurityRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficSecurityRules(ctx, req)
}
func (s *Server) GetTrafficSecurityRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficSecurityAuthSpec())
	if rsp != nil {
		return rsp
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetTrafficSecurityRules(ctx, query)
	s.markTrafficGovernanceBatchPermissions(authCtx, resp, trafficSecurityAuthSpec())
	return resp
}
func (s *Server) GetOneTrafficSecurityRule(ctx context.Context, req *apisecurity.TrafficSecurityRule) *apimodel.Response {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficSecurityAuthSpec())
	if rsp != nil {
		return api.NewResponse(apimodel.Code(rsp.Code))
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetOneTrafficSecurityRule(ctx, req)
	s.markTrafficGovernanceOnePermission(authCtx, resp, trafficSecurityAuthSpec())
	return resp
}

func (s *Server) CreateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficMirrorRules(ctx, req)
}
func (s *Server) UpdateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficMirrorRules(ctx, req)
}
func (s *Server) DeleteTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficMirrorRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficMirrorRules(ctx, req)
}
func (s *Server) GetTrafficMirrorRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficMirrorAuthSpec())
	if rsp != nil {
		return rsp
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetTrafficMirrorRules(ctx, query)
	s.markTrafficGovernanceBatchPermissions(authCtx, resp, trafficMirrorAuthSpec())
	return resp
}
func (s *Server) GetOneTrafficMirrorRule(ctx context.Context, req *apitraffic.TrafficMirror) *apimodel.Response {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficMirrorAuthSpec())
	if rsp != nil {
		return api.NewResponse(apimodel.Code(rsp.Code))
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetOneTrafficMirrorRule(ctx, req)
	s.markTrafficGovernanceOnePermission(authCtx, resp, trafficMirrorAuthSpec())
	return resp
}

func (s *Server) CreateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Create, authtypes.CreateTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.CreateTrafficMockRules(ctx, req)
}
func (s *Server) UpdateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Modify, authtypes.UpdateTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.UpdateTrafficMockRules(ctx, req)
}
func (s *Server) DeleteTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	if rsp := s.checkSimpleTrafficPermission(ctx, authtypes.Delete, authtypes.DeleteTrafficMockRules); rsp != nil {
		return rsp
	}
	return s.nextSvr.DeleteTrafficMockRules(ctx, req)
}
func (s *Server) GetTrafficMockRules(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficMockAuthSpec())
	if rsp != nil {
		return rsp
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetTrafficMockRules(ctx, query)
	s.markTrafficGovernanceBatchPermissions(authCtx, resp, trafficMockAuthSpec())
	return resp
}
func (s *Server) GetOneTrafficMockRule(ctx context.Context, req *apitraffic.TrafficMock) *apimodel.Response {
	authCtx, rsp := s.checkReadTrafficPermission(ctx, trafficMockAuthSpec())
	if rsp != nil {
		return api.NewResponse(apimodel.Code(rsp.Code))
	}
	ctx = authCtx.GetRequestContext()
	resp := s.nextSvr.GetOneTrafficMockRule(ctx, req)
	s.markTrafficGovernanceOnePermission(authCtx, resp, trafficMockAuthSpec())
	return resp
}

func (s *Server) checkReadTrafficPermission(ctx context.Context, spec trafficGovernanceAuthSpec) (*authtypes.AcquireContext, *apimodel.BatchQueryResponse) {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(authtypes.Read),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(spec.readMethod),
	)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return authCtx, api.NewAuthBatchQueryResponse(authtypes.ConvertToErrCode(err))
	}
	return authCtx, nil
}

func (s *Server) checkSimpleTrafficPermission(ctx context.Context, op authtypes.ResourceOperation, method authtypes.ServerFunctionName) *apimodel.BatchWriteResponse {
	authCtx := authtypes.NewAcquireContext(
		authtypes.WithRequestContext(ctx),
		authtypes.WithOperation(op),
		authtypes.WithModule(authtypes.DiscoverModule),
		authtypes.WithMethod(method),
	)
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		return api.NewBatchWriteResponse(authtypes.ConvertToErrCode(err))
	}
	return nil
}

func (s *Server) markTrafficGovernanceBatchPermissions(authCtx *authtypes.AcquireContext, resp *apimodel.BatchQueryResponse, spec trafficGovernanceAuthSpec) {
	if resp == nil {
		return
	}
	for _, anyData := range resp.Data {
		msg := newTrafficGovernanceAuthMessage(spec)
		if msg == nil {
			continue
		}
		if err := anypb.UnmarshalTo(anyData, msg, proto.UnmarshalOptions{}); err != nil {
			continue
		}
		s.markTrafficGovernanceMessagePermissions(authCtx, msg, spec)
		if updatedData, err := anypb.New(msg); err == nil {
			anyData.Value = updatedData.Value
			anyData.TypeUrl = updatedData.TypeUrl
		}
	}
}

func (s *Server) markTrafficGovernanceOnePermission(authCtx *authtypes.AcquireContext, resp *apimodel.Response, spec trafficGovernanceAuthSpec) {
	if resp == nil || resp.Data == nil {
		return
	}
	msg := newTrafficGovernanceAuthMessage(spec)
	if msg == nil {
		return
	}
	if err := anypb.UnmarshalTo(resp.Data, msg, proto.UnmarshalOptions{}); err != nil {
		return
	}
	s.markTrafficGovernanceMessagePermissions(authCtx, msg, spec)
	if updatedData, err := anypb.New(msg); err == nil {
		resp.Data = updatedData
	}
}

func newTrafficGovernanceAuthMessage(spec trafficGovernanceAuthSpec) proto.Message {
	switch spec.resource {
	case apisecurity.ResourceType_SecurityRules:
		return &apisecurity.TrafficSecurityRule{}
	case apisecurity.ResourceType_MirrorRules:
		return &apitraffic.TrafficMirror{}
	case apisecurity.ResourceType_MockRules:
		return &apitraffic.TrafficMock{}
	default:
		return nil
	}
}

func (s *Server) markTrafficGovernanceMessagePermissions(authCtx *authtypes.AcquireContext, msg proto.Message, spec trafficGovernanceAuthSpec) {
	id, metadata := trafficGovernanceAuthMessageResource(msg)
	setTrafficGovernanceAuthFlags(msg, true, true)
	authCtx.SetAccessResources(map[apisecurity.ResourceType][]authtypes.ResourceEntry{
		spec.resource: {
			{
				Type:     spec.resource,
				ID:       id,
				Metadata: metadata,
			},
		},
	})
	authCtx.SetMethod([]authtypes.ServerFunctionName{spec.updateMethod})
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		setTrafficGovernanceEditable(msg, false)
	}
	authCtx.SetMethod([]authtypes.ServerFunctionName{spec.deleteMethod})
	if _, err := s.policySvr.GetAuthChecker().CheckConsolePermission(authCtx); err != nil {
		setTrafficGovernanceDeleteable(msg, false)
	}
}

func trafficGovernanceAuthMessageResource(msg proto.Message) (string, map[string]string) {
	switch rule := msg.(type) {
	case *apisecurity.TrafficSecurityRule:
		return rule.GetId(), rule.GetMetadata()
	case *apitraffic.TrafficMirror:
		return rule.GetId(), rule.GetMetadata()
	case *apitraffic.TrafficMock:
		return rule.GetId(), rule.GetMetadata()
	default:
		return "", nil
	}
}

func setTrafficGovernanceAuthFlags(msg proto.Message, editable, deleteable bool) {
	setTrafficGovernanceEditable(msg, editable)
	setTrafficGovernanceDeleteable(msg, deleteable)
}

func setTrafficGovernanceEditable(msg proto.Message, editable bool) {
	switch rule := msg.(type) {
	case *apisecurity.TrafficSecurityRule:
		rule.Editable = editable
	case *apitraffic.TrafficMirror:
		rule.Editable = editable
	case *apitraffic.TrafficMock:
		rule.Editable = editable
	}
}

func setTrafficGovernanceDeleteable(msg proto.Message, deleteable bool) {
	switch rule := msg.(type) {
	case *apisecurity.TrafficSecurityRule:
		rule.Deleteable = deleteable
	case *apitraffic.TrafficMirror:
		rule.Deleteable = deleteable
	case *apitraffic.TrafficMock:
		rule.Deleteable = deleteable
	}
}
