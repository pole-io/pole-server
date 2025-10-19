package paramcheck

import (
	"context"
	"strconv"

	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

var (
	_allowRuleReleasesFilters = map[string]struct{}{
		"id":        {},
		"resource":  {},
		"rule_id":   {},
		"rule_name": {},
		"offset":    {},
		"limit":     {},
	}
)

// PublishLaneGroups 发布多个治理规则
func (svr *Server) PublishGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	return svr.nextSvr.PublishGovernanceRules(ctx, req)
}

func (svr *Server) GetRuleReleases(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse {
	newFilters := make(map[string]string)
	for key := range filter {
		if _, ok := _allowRuleReleasesFilters[key]; !ok {
			log.Errorf("[goverrule][release] params %s is not allowed in querying", key)
			continue
		}
		newFilters[key] = filter[key]
	}

	filter = newFilters
	if _, ok := apimodel.RuleRelease_RuleType_value[filter["resource"]]; !ok {
		log.Errorf("[goverrule][release] params resource %s is not allowed in querying", filter["resource"])
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	if filter["rule_id"] == "" && filter["rule_name"] == "" {
		log.Error("[goverrule][release] params rule_id or rule_name is required in querying")
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	// 处理offset和limit
	offset, limit, err := valid.ParseOffsetAndLimit(filter)
	if err != nil {
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}
	filter["offset"] = strconv.Itoa(int(offset))
	filter["limit"] = strconv.Itoa(int(limit))
	return svr.nextSvr.GetRuleReleases(ctx, filter)
}

// DeleteLaneGroups 删除多个治理规则已发布版本
func (svr *Server) DeleteGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	return svr.nextSvr.DeleteGovernanceRules(ctx, req)
}

// RollbackLaneGroups 回滚多个治理规则到目标版本
func (svr *Server) RollbackGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	return svr.nextSvr.RollbackGovernanceRules(ctx, req)
}

// StopbetaLaneGroups 停止多个治理规则灰度发布版本
func (svr *Server) StopbetaGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}

	if len(req) > valid.MaxBatchSize {
		return api.NewBatchWriteResponse(apimodel.Code_BatchSizeOverLimit)
	}
	return svr.nextSvr.StopbetaGovernanceRules(ctx, req)
}
