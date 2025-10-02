package goverrule

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
	"go.uber.org/zap"
)

// PublishLaneGroups 发布多个治理规则
func (s *Server) PublishGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		return s.PublishLaneGroups(ctx, req)
	case apimodel.RuleRelease_CircuitBreakerRules:
		return s.PublishCircuitBreakerRules(ctx, req)
	case apimodel.RuleRelease_FaultDetectRules:
		return s.PublishFaultDetectRules(ctx, req)
	case apimodel.RuleRelease_RouteRules:
		return s.PublishRouterRules(ctx, req)
	case apimodel.RuleRelease_RateLimitRules:
		return s.PublishRateLimits(ctx, req)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}
}

func (s *Server) GetRuleReleases(ctx context.Context, filter map[string]string) *apiservice.BatchQueryResponse {
	resource := apimodel.RuleRelease_RuleType(apimodel.RuleRelease_RuleType_value[filter["resource"]])

	// 处理offset和limit
	offset, limit, _ := valid.ParseOffsetAndLimit(filter)
	var (
		total    uint64
		versions []*rules.RuleRelease
		err      error
	)

	switch resource {
	case apimodel.RuleRelease_LaneRules:
		total, versions, err = s.storage.GetLaneGroupVersions(ctx, filter, offset, limit)
	case apimodel.RuleRelease_CircuitBreakerRules:
		total, versions, err = s.storage.GetCircuitBreakerRuleVersions(ctx, filter, offset, limit)
	case apimodel.RuleRelease_FaultDetectRules:
		total, versions, err = s.storage.GetFaultDetectRuleVersions(ctx, filter, offset, limit)
	case apimodel.RuleRelease_RouteRules:
		total, versions, err = s.storage.GetRouterRuleVersions(ctx, filter, offset, limit)
	case apimodel.RuleRelease_RateLimitRules:
		total, versions, err = s.storage.GetRateLimitRuleVersions(ctx, filter, offset, limit)
	default:
		return api.NewBatchQueryResponse(apimodel.Code_InvalidParameter)
	}

	if err != nil {
		log.Error("[govertule][release][list] get rule releases error", utils.RequestID(ctx), zap.Error(err))
		return api.NewBatchQueryResponse(apimodel.Code_ExecuteException)
	}

	rsp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	rsp.Amount = protobuf.NewUInt32Value(uint32(total))
	rsp.Size = protobuf.NewUInt32Value(uint32(len(versions)))
	for i := range versions {
		if err := api.AddAnyDataIntoBatchQuery(rsp, versions[i].ToSpec()); err != nil {
			log.Error("[govertule][release][list] add data into batch query error", utils.RequestID(ctx), zap.Error(err))
			return api.NewBatchQueryResponse(apimodel.Code_ParseException)
		}
	}
	return rsp
}

// DeleteLaneGroups 删除多个治理规则已发布版本
func (s *Server) DeleteGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	return nil
}

// RollbackLaneGroups 回滚多个治理规则到目标版本
func (s *Server) RollbackGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		return s.RollbackLaneGroups(ctx, req)
	case apimodel.RuleRelease_CircuitBreakerRules:
		return s.RollbackCircuitBreakerRules(ctx, req)
	case apimodel.RuleRelease_FaultDetectRules:
		return s.RollbackFaultDetectRules(ctx, req)
	case apimodel.RuleRelease_RouteRules:
		return s.RollbackRouterRules(ctx, req)
	case apimodel.RuleRelease_RateLimitRules:
		return s.RollbackRateLimits(ctx, req)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}
}

// StopbetaLaneGroups 停止多个治理规则灰度发布版本
func (s *Server) StopbetaGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		return s.StopbetaLaneGroups(ctx, req)
	case apimodel.RuleRelease_CircuitBreakerRules:
		return s.StopbetaCircuitBreakerRules(ctx, req)
	case apimodel.RuleRelease_FaultDetectRules:
		return s.StopbetaFaultDetectRules(ctx, req)
	case apimodel.RuleRelease_RouteRules:
		return s.StopbetaRouterRules(ctx, req)
	case apimodel.RuleRelease_RateLimitRules:
		return s.StopbetaRateLimits(ctx, req)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}
}

// Rulepipline 用于治理规则的并发控制
type RuleReleasePipeline struct {
	lock                  func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apiservice.Response)
	checkExistRelease     func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error)
	checkExistGrayRelease func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error)
	publish               func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) error
}

// NewRuleReleasePipeline 工厂函数，简化各类规则发布的 pipeline 构造
func NewRuleReleasePipeline[
	Rule any, // 规则类型
	Release any, // 发布类型
](
	lockFn func(store.Tx, string) (Rule, error),
	getReleaseFn func(store.Tx, *rules.RuleRelease) (Release, error),
	getActiveFn func(store.Tx, any) (Release, error),
	publishFn func(store.Tx, Release) error,
	releaseBuilder func(*apimodel.RuleRelease, any) Release,
) *RuleReleasePipeline {
	isNil := func(v any) bool {
		if v == nil {
			return true
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
			return rv.IsNil()
		}
		return false
	}
	return &RuleReleasePipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apiservice.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if isNil(rule) {
				return nil, api.NewResponse(apimodel.Code_NotFoundResource)
			}
			return rule, nil
		},
		checkExistRelease: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error) {
			existRes, err := getReleaseFn(tx, &rules.RuleRelease{
				RuleName:    req.RuleName,
				ReleaseName: req.ReleaseName,
			})
			if err != nil {
				return false, err
			}
			return !isNil(existRes), nil
		},
		checkExistGrayRelease: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error) {
			grayData := &rules.RuleRelease{}
			grayData.FromSpec(req)
			grayData.ReleaseType = rules.ReleaseTypeGray
			grayRelease := releaseBuilder(req, nil)
			// 用 grayData 替换 release 字段
			switch v := any(grayRelease).(type) {
			case *rules.LaneGroupRelease:
				v.RuleRelease = *grayData
			case *rules.CircuitBreakerRelease:
				v.RuleRelease = *grayData
			case *rules.FaultDetectRelease:
				v.RuleRelease = *grayData
			case *rules.RouterRuleRelease:
				v.RuleRelease = *grayData
			case *rules.RateLimitRelease:
				v.RuleRelease = *grayData
			}
			res, err := getActiveFn(tx, grayRelease)
			if err != nil {
				return false, err
			}
			return !isNil(res), nil
		},
		publish: func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) error {
			release := releaseBuilder(req, rule)
			return publishFn(tx, release)
		},
	}
}

// PublishCircuitBreakerRules implements GoverRuleServer.
func (s *Server) PublishCircuitBreakerRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockCircuitBreakerRule,
		s.storage.GetReleaseCircuitBreakerRule,
		func(tx store.Tx, rel any) (*rules.CircuitBreakerRelease, error) {
			return s.storage.GetActiveCircuitBreakerRule(tx, rel.(*rules.CircuitBreakerRelease))
		},
		s.storage.PublishCircuitBreakerRule,
		func(req *apimodel.RuleRelease, rule any) *rules.CircuitBreakerRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			var cbRule *rules.CircuitBreakerRule
			if rule != nil {
				cbRule = rule.(*rules.CircuitBreakerRule)
			}
			return &rules.CircuitBreakerRelease{
				RuleRelease: *curData,
				Rule:        cbRule,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

// PublishFaultDetectRules implements GoverRuleServer.
func (s *Server) PublishFaultDetectRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockFaultDetectRule,
		s.storage.GetReleaseFaultDetectRule,
		func(tx store.Tx, rel any) (*rules.FaultDetectRelease, error) {
			return s.storage.GetActiveFaultDetectRule(tx, rel.(*rules.FaultDetectRelease))
		},
		s.storage.PublishFaultDetectRule,
		func(req *apimodel.RuleRelease, rule any) *rules.FaultDetectRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			var fdRule *rules.FaultDetectRule
			if rule != nil {
				fdRule = rule.(*rules.FaultDetectRule)
			}
			return &rules.FaultDetectRelease{
				RuleRelease: *curData,
				Rule:        fdRule,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

// LaneGroup 发布示例
func (s *Server) PublishLaneGroups(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockLaneGroup,
		s.storage.GetReleaseLaneGroupRule,
		func(tx store.Tx, rel any) (*rules.LaneGroupRelease, error) {
			return s.storage.GetActiveLaneGroup(tx, rel.(*rules.LaneGroupRelease))
		},
		s.storage.PublishLaneGroup,
		func(req *apimodel.RuleRelease, rule any) *rules.LaneGroupRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			curData.Id = utils.NewUUID()
			var laneRule *rules.LaneGroup
			if rule != nil {
				laneRule = rule.(*rules.LaneGroup)
			}
			protoVal, _ := laneRule.ToProto()
			return &rules.LaneGroupRelease{
				RuleRelease: *curData,
				Rule:        protoVal,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

func (s *Server) PublishRateLimits(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockRateLimitRule,
		s.storage.GetReleaseRateLimitRule,
		func(tx store.Tx, rel any) (*rules.RateLimitRelease, error) {
			return s.storage.GetActiveRateLimitRule(tx, rel.(*rules.RateLimitRelease))
		},
		s.storage.PublishRateLimitRule,
		func(req *apimodel.RuleRelease, rule any) *rules.RateLimitRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			var rlRule *rules.RateLimit
			if rule != nil {
				rlRule = rule.(*rules.RateLimit)
			}
			return &rules.RateLimitRelease{
				RuleRelease: *curData,
				Rule:        rlRule,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

func (s *Server) PublishRouterRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockRouterRule,
		s.storage.GetReleaseRouterRule,
		func(tx store.Tx, rel any) (*rules.RouterRuleRelease, error) {
			return s.storage.GetActiveRouterRule(tx, rel.(*rules.RouterRuleRelease))
		},
		s.storage.PublishRouterRule,
		func(req *apimodel.RuleRelease, rule any) *rules.RouterRuleRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			curData.Id = utils.NewUUID()
			var routerRule *rules.RouterConfig
			if rule != nil {
				routerRule = rule.(*rules.RouterConfig)
				curData.RuleId = routerRule.ID
				curData.RuleName = routerRule.Name
			}
			pdata, _ := routerRule.ToExpendRoutingConfig()
			return &rules.RouterRuleRelease{
				RuleRelease: *curData,
				Rule:        pdata,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

// RuleRollbackPipeline 用于治理规则的回滚控制操作
type RuleRollbackPipeline struct {
	lock              func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apiservice.Response)
	checkExistRelease func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (*rules.RuleRelease, error)
	active            func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) *apiservice.Response
}

// NewRuleRollbackPipeline 工厂函数，简化各类规则回滚的 pipeline 构造
func NewRuleRollbackPipeline(
	lockFn func(store.Tx, string) (any, error),
	getReleaseFn func(store.Tx, *rules.RuleRelease) (any, error),
	activeFn func(store.Tx, any, any, *apimodel.RuleRelease) error,
) *RuleRollbackPipeline {
	isNil := func(v any) bool {
		if v == nil {
			return true
		}
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
			return rv.IsNil()
		}
		return false
	}
	return &RuleRollbackPipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apiservice.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if isNil(rule) {
				return nil, api.NewResponse(apimodel.Code_NotFoundResource)
			}
			return rule, nil
		},
		checkExistRelease: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (*rules.RuleRelease, error) {
			existRes, err := getReleaseFn(tx, &rules.RuleRelease{
				RuleName:    req.RuleName,
				ReleaseName: req.ReleaseName,
			})
			if err != nil || isNil(existRes) {
				return nil, err
			}
			switch v := existRes.(type) {
			case *rules.CircuitBreakerRelease:
				return &v.RuleRelease, nil
			case *rules.FaultDetectRelease:
				return &v.RuleRelease, nil
			case *rules.LaneGroupRelease:
				return &v.RuleRelease, nil
			case *rules.RouterRuleRelease:
				return &v.RuleRelease, nil
			case *rules.RateLimitRelease:
				return &v.RuleRelease, nil
			}
			return nil, nil
		},
		active: func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) *apiservice.Response {
			existRes, err := getReleaseFn(tx, &rules.RuleRelease{
				RuleName:    req.RuleName,
				ReleaseName: req.ReleaseName,
			})
			if err != nil || isNil(existRes) {
				return api.NewResponseWithMsg(apimodel.Code_ExecuteException, "release not found or error")
			}
			if err := activeFn(tx, existRes, rule, req); err != nil {
				return api.NewResponse(storeapi.StoreCode2APICode(err))
			}
			return api.NewResponse(apimodel.Code_ExecuteSuccess)
		},
	}
}

// RollbackCircuitBreakerRules 泛型重构示例
func (s *Server) RollbackCircuitBreakerRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleRollbackPipeline(
		func(tx store.Tx, name string) (any, error) {
			return s.storage.LockCircuitBreakerRule(tx, name)
		},
		func(tx store.Tx, rel *rules.RuleRelease) (any, error) {
			return s.storage.GetReleaseCircuitBreakerRule(tx, rel)
		},
		func(tx store.Tx, rel any, rule any, req *apimodel.RuleRelease) error {
			return s.storage.ActiveCircuitBreakerRule(tx, &rules.CircuitBreakerRelease{
				RuleRelease: rel.(*rules.CircuitBreakerRelease).RuleRelease,
				Rule:        rule.(*rules.CircuitBreakerRule),
			})
		},
	)
	return s.exectueRuleRollbackPipline(ctx, pipeline, requests)
}

// RollbackFaultDetectRules implements GoverRuleServer.
func (s *Server) RollbackFaultDetectRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleRollbackPipeline(
		func(tx store.Tx, name string) (any, error) {
			return s.storage.LockFaultDetectRule(tx, name)
		},
		func(tx store.Tx, rel *rules.RuleRelease) (any, error) {
			return s.storage.GetReleaseFaultDetectRule(tx, rel)
		},
		func(tx store.Tx, rel any, rule any, req *apimodel.RuleRelease) error {
			return s.storage.ActiveFaultDetectRule(tx, &rules.FaultDetectRelease{
				RuleRelease: rel.(*rules.FaultDetectRelease).RuleRelease,
				Rule:        rule.(*rules.FaultDetectRule),
			})
		},
	)
	return s.exectueRuleRollbackPipline(ctx, pipeline, requests)
}

// RollbackLaneGroups implements GoverRuleServer.
func (s *Server) RollbackLaneGroups(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleRollbackPipeline(
		func(tx store.Tx, name string) (any, error) {
			return s.storage.LockLaneGroup(tx, name)
		},
		func(tx store.Tx, rel *rules.RuleRelease) (any, error) {
			return s.storage.GetReleaseLaneGroupRule(tx, rel)
		},
		func(tx store.Tx, rel any, rule any, req *apimodel.RuleRelease) error {
			return s.storage.ActiveLaneGroup(tx, &rules.LaneGroupRelease{
				RuleRelease: rel.(*rules.LaneGroupRelease).RuleRelease,
			})
		},
	)
	return s.exectueRuleRollbackPipline(ctx, pipeline, requests)
}

// RollbackRateLimits implements GoverRuleServer.
func (s *Server) RollbackRateLimits(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleRollbackPipeline(
		func(tx store.Tx, name string) (any, error) {
			return s.storage.LockRateLimitRule(tx, name)
		},
		func(tx store.Tx, rel *rules.RuleRelease) (any, error) {
			return s.storage.GetReleaseRateLimitRule(tx, rel)
		},
		func(tx store.Tx, rel any, rule any, req *apimodel.RuleRelease) error {
			return s.storage.ActiveRateLimitRule(tx, &rules.RateLimitRelease{
				RuleRelease: rel.(*rules.RateLimitRelease).RuleRelease,
				Rule:        rule.(*rules.RateLimit),
			})
		},
	)
	return s.exectueRuleRollbackPipline(ctx, pipeline, requests)
}

// RollbackRouterRules implements GoverRuleServer.
func (s *Server) RollbackRouterRules(ctx context.Context, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	pipeline := NewRuleRollbackPipeline(
		func(tx store.Tx, name string) (any, error) {
			return s.storage.LockRouterRule(tx, name)
		},
		func(tx store.Tx, rel *rules.RuleRelease) (any, error) {
			return s.storage.GetReleaseRouterRule(tx, rel)
		},
		func(tx store.Tx, rel any, rule any, req *apimodel.RuleRelease) error {
			return s.storage.ActiveRouterRule(tx, &rules.RouterRuleRelease{
				RuleRelease: rel.(*rules.RouterRuleRelease).RuleRelease,
			})
		},
	)
	return s.exectueRuleRollbackPipline(ctx, pipeline, requests)
}

// RuleStopbetaPipeline 用于治理规则的停止灰度发布控制操作
type RuleStopbetaPipeline struct {
	lock              func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apiservice.Response)
	checkExistRelease func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (*rules.RuleRelease, error)
	active            func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) *apiservice.Response
}

// StopbetaCircuitBreakerRules implements GoverRuleServer.
func (s *Server) StopbetaCircuitBreakerRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apiservice.Response {
		reqData := &rules.RuleRelease{}
		reqData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}

		defer tx.Rollback() // 最终的异常 case 的兜底

		rule, err := s.storage.LockCircuitBreakerRule(tx, req.RuleName)
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		if rule == nil {
			return api.NewResponse(apimodel.Code_NotFoundResource)
		}

		// 查看目标版本是否存在
		existRes, err := s.storage.GetReleaseCircuitBreakerRule(tx, &rules.RuleRelease{
			RuleName:    req.RuleName,
			ReleaseName: req.ReleaseName,
			ReleaseType: rules.ReleaseTypeGray,
		})
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		if existRes == nil {
			return api.NewResponse(apimodel.Code_NotFoundResource)
		}

		if err := s.storage.InactiveCircuitBreakerRule(tx, &rules.CircuitBreakerRelease{
			RuleRelease: *reqData,
			Rule:        rule,
		}); err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}

		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] stopbeta release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", reqData.RuleId), zap.String("rule-name", reqData.RuleName), zap.Error(err))
			return api.NewResponse(storeapi.StoreCode2APICode(err))
		}
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}

	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range request {
		rsp := handle(ctx, request[i])
		api.Collect(batchRsp, rsp)
	}
	return batchRsp
}

// StopbetaFaultDetectRules implements GoverRuleServer.
func (s *Server) StopbetaFaultDetectRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaLaneGroups implements GoverRuleServer.
func (s *Server) StopbetaLaneGroups(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRateLimits implements GoverRuleServer.
func (s *Server) StopbetaRateLimits(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRouterRules implements GoverRuleServer.
func (s *Server) StopbetaRouterRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaCircuitBreakerRules implements GoverRuleServer.
func (s *Server) DeleteCircuitBreakerReleases(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaFaultDetectRules implements GoverRuleServer.
func (s *Server) DeleteFaultDetectReleases(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaLaneGroups implements GoverRuleServer.
func (s *Server) DeleteLaneGroupReleases(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRateLimits implements GoverRuleServer.
func (s *Server) DeleteRateLimitReleases(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaRouterRules implements GoverRuleServer.
func (s *Server) DeleteRouterReleases(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// executeRuleReleasePipline 执行通用的治理规则灰度发布执行
func (s *Server) executeRuleReleasePipline(ctx context.Context, pipline *RuleReleasePipeline, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apiservice.Response {
		curData := &rules.RuleRelease{}
		curData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		defer tx.Rollback() // 最终的异常 case 的兜底

		// 锁住目标规则，避免同时还有别的操作，导致发布不符合预期
		rule, errRsp := pipline.lock(ctx, tx, req)
		if errRsp != nil {
			return errRsp
		}

		// 查看目标版本是否存在
		existRes, err := pipline.checkExistRelease(ctx, tx, req)
		if err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		if existRes {
			return api.NewResponse(apimodel.Code_ExistedResource)
		}

		// 如果是发布正常的版本，需要检查下是否存在灰度发布的版本，存在的话，需要先结束
		if curData.ReleaseType == rules.ReleaseTypeNormal {
			grayRes, err := pipline.checkExistGrayRelease(ctx, tx, req)
			if err != nil {
				return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if grayRes {
				return api.NewResponse(apimodel.Code_DataConflict)
			}
		}

		// 发布规则
		if err := pipline.publish(ctx, tx, rule, req); err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] publish release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", curData.RuleId), zap.String("rule-name", curData.RuleName), zap.Error(err))
			return api.NewResponse(storeapi.StoreCode2APICode(err))
		}
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}

	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range requests {
		rsp := handle(ctx, requests[i])
		api.Collect(batchRsp, rsp)
	}
	return batchRsp
}

func (s *Server) exectueRuleRollbackPipline(ctx context.Context, pipline *RuleRollbackPipeline, requests []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apiservice.Response {
		reqData := &rules.RuleRelease{}
		reqData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		defer tx.Rollback() // 最终的异常 case 的兜底

		rule, errRsp := pipline.lock(ctx, tx, req)
		if errRsp != nil {
			return errRsp
		}

		// 检查目标版本是否存在
		existRes, err := pipline.checkExistRelease(ctx, tx, req)
		if err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		if existRes == nil {
			return api.NewResponse(apimodel.Code_NotFoundResource)
		}

		// 灰度版本不允许回滚
		if existRes.ReleaseType == rules.ReleaseTypeGray {
			return api.NewResponseWithMsg(apimodel.Code_BadRequest, "gray release not allow rollback")
		}

		if errRsp := pipline.active(ctx, tx, rule, req); errRsp != nil {
			return errRsp
		}
		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] rollback release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", reqData.RuleId), zap.String("rule-name", reqData.RuleName), zap.String("target-version", req.ReleaseName),
				zap.Error(err))
			return api.NewResponse(storeapi.StoreCode2APICode(err))
		}
		return api.NewResponse(apimodel.Code_ExecuteSuccess)
	}

	batchRsp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for i := range requests {
		rsp := handle(ctx, requests[i])
		api.Collect(batchRsp, rsp)
	}
	return batchRsp
}

// SaveGrayRule
func SaveGrayRule(ctx context.Context, tx store.Tx, s store.GrayStore, rule rules.GrayRule) *apiservice.Response {
	clientLabels := rule.GetClientLabels()
	raw := make([]json.RawMessage, 0, len(clientLabels))
	marshaler := jsonpb.Marshaler{}
	for i := range clientLabels {
		data, err := marshaler.MarshalToString(clientLabels[i])
		if err != nil {
			log.Error("[Config][Release] marshal gary rule error.",
				utils.RequestID(ctx), zap.String("resource", rule.GetGrayResource()), zap.Error(err))
			return api.NewResponseWithMsg(apimodel.Code_InvalidMatchRule, err.Error())
		}
		raw = append(raw, json.RawMessage(data))
	}
	grayResource := &rules.GrayResource{
		Name:      rule.GetGrayResource(),
		MatchRule: string(utils.MustJson(raw)),
		CreateBy:  utils.ParseUserName(ctx),
		ModifyBy:  utils.ParseUserName(ctx),
	}
	if err := s.CreateGrayResourceTx(tx, grayResource); err != nil {
		log.Error("[Config][Release] create gray resource error.",
			utils.RequestID(ctx), zap.String("resource", rule.GetGrayResource()), zap.Error(err))
		return api.NewResponse(storeapi.StoreCode2APICode(err))
	}
	return nil
}
