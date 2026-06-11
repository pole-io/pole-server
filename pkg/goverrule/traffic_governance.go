package goverrule

import (
	"context"
	"fmt"
	"time"

	oldproto "github.com/golang/protobuf/proto"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/pkg/utils/revision"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	newproto "google.golang.org/protobuf/proto"
)

type trafficGovernanceSpec[T oldproto.Message] struct {
	logName     string
	resource    types.Resource
	newRule     func(T) *rules.TrafficGovernanceRule
	toSpec      func(*rules.TrafficGovernanceRule) T
	create      func(*rules.TrafficGovernanceRule) error
	update      func(*rules.TrafficGovernanceRule) error
	delete      func(*rules.TrafficGovernanceRule) error
	getOne      func(string) (*rules.TrafficGovernanceRule, error)
	query       func(context.Context, *cachetypes.TrafficGovernanceArgs) (uint32, []*rules.TrafficGovernanceRule, error)
	displayName func(*rules.TrafficGovernanceRule) string
}

func (s *Server) CreateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return createTrafficGovernanceRules(s, ctx, req, s.trafficSecuritySpec())
}

func (s *Server) UpdateTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return updateTrafficGovernanceRules(s, ctx, req, s.trafficSecuritySpec())
}

func (s *Server) DeleteTrafficSecurityRules(ctx context.Context, req []*apisecurity.TrafficSecurityRule) *apimodel.BatchWriteResponse {
	return deleteTrafficGovernanceRules(s, ctx, req, s.trafficSecuritySpec())
}

func (s *Server) GetTrafficSecurityRules(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	return getTrafficGovernanceRules(ctx, filter, s.trafficSecuritySpec())
}

func (s *Server) GetOneTrafficSecurityRule(ctx context.Context, req *apisecurity.TrafficSecurityRule) *apimodel.Response {
	return getOneTrafficGovernanceRule(ctx, req.GetId(), s.trafficSecuritySpec())
}

func (s *Server) CreateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return createTrafficGovernanceRules(s, ctx, req, s.trafficMirrorSpec())
}

func (s *Server) UpdateTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return updateTrafficGovernanceRules(s, ctx, req, s.trafficMirrorSpec())
}

func (s *Server) DeleteTrafficMirrorRules(ctx context.Context, req []*apitraffic.TrafficMirror) *apimodel.BatchWriteResponse {
	return deleteTrafficGovernanceRules(s, ctx, req, s.trafficMirrorSpec())
}

func (s *Server) GetTrafficMirrorRules(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	return getTrafficGovernanceRules(ctx, filter, s.trafficMirrorSpec())
}

func (s *Server) GetOneTrafficMirrorRule(ctx context.Context, req *apitraffic.TrafficMirror) *apimodel.Response {
	return getOneTrafficGovernanceRule(ctx, req.GetId(), s.trafficMirrorSpec())
}

func (s *Server) CreateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return createTrafficGovernanceRules(s, ctx, req, s.trafficMockSpec())
}

func (s *Server) UpdateTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return updateTrafficGovernanceRules(s, ctx, req, s.trafficMockSpec())
}

func (s *Server) DeleteTrafficMockRules(ctx context.Context, req []*apitraffic.TrafficMock) *apimodel.BatchWriteResponse {
	return deleteTrafficGovernanceRules(s, ctx, req, s.trafficMockSpec())
}

func (s *Server) GetTrafficMockRules(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
	return getTrafficGovernanceRules(ctx, filter, s.trafficMockSpec())
}

func (s *Server) GetOneTrafficMockRule(ctx context.Context, req *apitraffic.TrafficMock) *apimodel.Response {
	return getOneTrafficGovernanceRule(ctx, req.GetId(), s.trafficMockSpec())
}

func (s *Server) trafficSecuritySpec() trafficGovernanceSpec[*apisecurity.TrafficSecurityRule] {
	return trafficGovernanceSpec[*apisecurity.TrafficSecurityRule]{
		logName:  "traffic-security",
		resource: types.RTrafficSecurityRule,
		newRule:  rules.NewTrafficSecurityRule,
		toSpec: func(rule *rules.TrafficGovernanceRule) *apisecurity.TrafficSecurityRule {
			return rule.ToTrafficSecuritySpec()
		},
		create:      s.storage.CreateTrafficSecurityRule,
		update:      s.storage.UpdateTrafficSecurityRule,
		delete:      s.storage.DeleteTrafficSecurityRule,
		getOne:      s.storage.GetOneTrafficSecurityRule,
		query:       s.caches.TrafficSecurity().Query,
		displayName: func(rule *rules.TrafficGovernanceRule) string { return fmt.Sprintf("%s(%s)", rule.Name, rule.ID) },
	}
}

func (s *Server) trafficMirrorSpec() trafficGovernanceSpec[*apitraffic.TrafficMirror] {
	return trafficGovernanceSpec[*apitraffic.TrafficMirror]{
		logName:     "traffic-mirror",
		resource:    types.RTrafficMirrorRule,
		newRule:     rules.NewTrafficMirrorRule,
		toSpec:      func(rule *rules.TrafficGovernanceRule) *apitraffic.TrafficMirror { return rule.ToTrafficMirrorSpec() },
		create:      s.storage.CreateTrafficMirrorRule,
		update:      s.storage.UpdateTrafficMirrorRule,
		delete:      s.storage.DeleteTrafficMirrorRule,
		getOne:      s.storage.GetOneTrafficMirrorRule,
		query:       s.caches.TrafficMirror().Query,
		displayName: func(rule *rules.TrafficGovernanceRule) string { return fmt.Sprintf("%s(%s)", rule.Name, rule.ID) },
	}
}

func (s *Server) trafficMockSpec() trafficGovernanceSpec[*apitraffic.TrafficMock] {
	return trafficGovernanceSpec[*apitraffic.TrafficMock]{
		logName:     "traffic-mock",
		resource:    types.RTrafficMockRule,
		newRule:     rules.NewTrafficMockRule,
		toSpec:      func(rule *rules.TrafficGovernanceRule) *apitraffic.TrafficMock { return rule.ToTrafficMockSpec() },
		create:      s.storage.CreateTrafficMockRule,
		update:      s.storage.UpdateTrafficMockRule,
		delete:      s.storage.DeleteTrafficMockRule,
		getOne:      s.storage.GetOneTrafficMockRule,
		query:       s.caches.TrafficMock().Query,
		displayName: func(rule *rules.TrafficGovernanceRule) string { return fmt.Sprintf("%s(%s)", rule.Name, rule.ID) },
	}
}

func createTrafficGovernanceRules[T oldproto.Message](s *Server, ctx context.Context, request []T, spec trafficGovernanceSpec[T]) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range request {
		data := spec.newRule(req)
		data.ID = utils.NewUUID()
		data.Revision = revision.NewRevision()
		if err := spec.create(data); err != nil {
			log.Error("["+spec.logName+"] create error", zap.Error(err), utils.RequestID(ctx))
			api.Collect(responses, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			continue
		}
		s.RecordHistory(ctx, trafficGovernanceRecordEntry(ctx, req, data, spec, types.OCreate))
		api.Collect(responses, api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, spec.toSpec(data)))
	}
	return api.FormatBatchWriteResponse(responses)
}

func updateTrafficGovernanceRules[T oldproto.Message](s *Server, ctx context.Context, request []T, spec trafficGovernanceSpec[T]) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range request {
		data := spec.newRule(req)
		old, err := spec.getOne(data.ID)
		if err != nil {
			log.Error("["+spec.logName+"] get for update error", zap.Error(err), utils.RequestID(ctx))
			api.Collect(responses, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			continue
		}
		if old == nil {
			api.Collect(responses, api.NewResponse(apimodel.Code_NotFoundResource))
			continue
		}
		data.Revision = revision.NewRevision()
		if err := spec.update(data); err != nil {
			log.Error("["+spec.logName+"] update error", zap.Error(err), utils.RequestID(ctx))
			api.Collect(responses, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			continue
		}
		s.RecordHistory(ctx, trafficGovernanceRecordEntry(ctx, req, data, spec, types.OUpdate))
		api.Collect(responses, api.NewResponse(apimodel.Code_ExecuteSuccess))
	}
	return api.FormatBatchWriteResponse(responses)
}

func deleteTrafficGovernanceRules[T oldproto.Message](s *Server, ctx context.Context, request []T, spec trafficGovernanceSpec[T]) *apimodel.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, req := range request {
		data := spec.newRule(req)
		old, err := spec.getOne(data.ID)
		if err != nil {
			log.Error("["+spec.logName+"] get for delete error", zap.Error(err), utils.RequestID(ctx))
			api.Collect(responses, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			continue
		}
		if old == nil {
			api.Collect(responses, api.NewResponse(apimodel.Code_ExecuteSuccess))
			continue
		}
		if err := spec.delete(data); err != nil {
			log.Error("["+spec.logName+"] delete error", zap.Error(err), utils.RequestID(ctx))
			api.Collect(responses, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
			continue
		}
		s.RecordHistory(ctx, trafficGovernanceRecordEntry(ctx, req, old, spec, types.ODelete))
		api.Collect(responses, api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req))
	}
	return api.FormatBatchWriteResponse(responses)
}

func getTrafficGovernanceRules[T oldproto.Message](ctx context.Context, query map[string]string, spec trafficGovernanceSpec[T]) *apimodel.BatchQueryResponse {
	offset, limit, _ := valid.ParseOffsetAndLimit(query)
	total, list, err := spec.query(ctx, &cachetypes.TrafficGovernanceArgs{
		Filter: query,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error("["+spec.logName+"] query error", zap.Error(err), utils.RequestID(ctx))
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}
	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Size = uint32(len(list))
	resp.Amount = total
	for i := range list {
		api.AddAnyDataIntoBatchQuery(resp, spec.toSpec(list[i]))
	}
	return resp
}

func getOneTrafficGovernanceRule[T oldproto.Message](ctx context.Context, id string, spec trafficGovernanceSpec[T]) *apimodel.Response {
	rule, err := spec.getOne(id)
	if err != nil {
		log.Error("["+spec.logName+"] get one error", zap.Error(err), utils.RequestID(ctx))
		return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	}
	if rule == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, spec.toSpec(rule))
}

func trafficGovernanceRecordEntry[T oldproto.Message](
	ctx context.Context,
	req T,
	rule *rules.TrafficGovernanceRule,
	spec trafficGovernanceSpec[T],
	opt types.OperationType,
) *types.RecordEntry {
	var detail []byte
	if msg, ok := any(req).(newproto.Message); ok {
		detail, _ = protojson.Marshal(msg)
	} else {
		detail = []byte(oldproto.CompactTextString(req))
	}
	return &types.RecordEntry{
		ResourceType:  spec.resource,
		ResourceName:  spec.displayName(rule),
		Namespace:     rule.Namespace,
		Operator:      utils.ParseOperator(ctx),
		OperationType: opt,
		Detail:        string(detail),
		HappenTime:    time.Now(),
	}
}
