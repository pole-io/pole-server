package goverrule

import (
	"context"

	"fmt"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.uber.org/zap"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	apitraffic "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
)

func (s *Server) CreateLossLessRules(ctx context.Context, request []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, rule := range request {
		response := s.CreateLossLessRule(ctx, rule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

func (s *Server) CreateLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response {
	// 检查规则是否存在

	data := &rules.LosslessRule{}
	data.FromSpec(req)
	data.ID = utils.NewUUID()
	data.Revision = revision.NewRevision()
	if err := s.storage.CreateLossLessRule(data); err != nil {
		log.Error("[lossless] create error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	s.RecordHistory(ctx, losslessRecordEntry(ctx, req, data, types.OCreate))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, data.ToSpec())
}

func (s *Server) DeleteLossLessRules(ctx context.Context, request []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, rule := range request {
		response := s.DeleteLossLessRule(ctx, rule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

func (s *Server) DeleteLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response {
	// 检查是否存在
	old, err := s.storage.GetOneLosslessRule(req.GetId())
	if err != nil {
		log.Error("[lossless] get for delete error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	if old == nil {
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}
	data := &rules.LosslessRule{}
	data.FromSpec(req)
	if err := s.storage.DeleteLossLessRule(data); err != nil {
		log.Error("[lossless] delete error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	s.RecordHistory(ctx, losslessRecordEntry(ctx, req, old, types.ODelete))
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, req)
}

func (s *Server) UpdateLossLessRules(ctx context.Context, request []*apitraffic.LosslessRule) *apiservice.BatchWriteResponse {
	responses := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, rule := range request {
		response := s.UpdateLossLessRule(ctx, rule)
		api.Collect(responses, response)
	}
	return api.FormatBatchWriteResponse(responses)
}

func (s *Server) UpdateLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response {
	old, err := s.storage.GetOneLosslessRule(req.GetId())
	if err != nil {
		log.Error("[lossless] get for update error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	if old == nil {
		return api.NewResponse(apimodel.Code_NotFoundResource)
	}
	data := &rules.LosslessRule{}
	data.FromSpec(req)
	if err != nil {
		log.Error("[lossless] api2LosslessRule error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	data.Revision = revision.NewRevision()
	if err := s.storage.UpdateLossLessRule(data); err != nil {
		log.Error("[lossless] update error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	s.RecordHistory(ctx, losslessRecordEntry(ctx, req, data, types.OUpdate))
	return api.NewResponse(apimodel.Code_ExecuteSuccess)
}

func (s *Server) GetLossLessRules(ctx context.Context, query map[string]string) *apiservice.BatchQueryResponse {
	// 参数解析
	offset, limit, _ := valid.ParseOffsetAndLimit(query)
	total, list, err := s.caches.Lossless().Query(ctx, &cachetypes.LosslessArgs{
		Filter: query,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Error("[lossless] query error", zap.Error(err), utils.RequestID(ctx))
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}
	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Size = protobuf.NewUInt32Value(uint32(len(list)))
	resp.Amount = protobuf.NewUInt32Value(total)
	for i := range list {
		api.AddAnyDataIntoBatchQuery(resp, list[i].ToSpec())
	}
	return resp
}

func (s *Server) GetOneLossLessRule(ctx context.Context, req *apitraffic.LosslessRule) *apiservice.Response {
	rule, err := s.storage.GetOneLosslessRule(req.GetId())
	if err != nil {
		log.Error("[lossless] get one error", zap.Error(err), utils.RequestID(ctx))
		return wrapperLosslessStoreResponse(req, err)
	}
	apiRule := rule.ToSpec()
	return api.NewAnyDataResponse(apimodel.Code_ExecuteSuccess, apiRule)
}

// wrapperLosslessStoreResponse 封装存储层错误
func wrapperLosslessStoreResponse(rule *apitraffic.LosslessRule, err error) *apiservice.Response {
	if err == nil {
		return nil
	}
	resp := api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
	return resp
}

// losslessRecordEntry 构建lossless的记录entry
func losslessRecordEntry(ctx context.Context, req *apitraffic.LosslessRule, md *rules.LosslessRule, opt types.OperationType) *types.RecordEntry {
	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)
	entry := &types.RecordEntry{
		ResourceType:  types.RLosslessRule,
		ResourceName:  fmt.Sprintf("%s(%s)", md.Service, md.ID),
		Namespace:     req.GetNamespace(),
		Operator:      utils.ParseOperator(ctx),
		OperationType: opt,
		Detail:        detail,
		HappenTime:    time.Now(),
	}
	return entry
}
