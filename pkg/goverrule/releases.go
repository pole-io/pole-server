package goverrule

import (
	"context"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pole-io/pole-server/apis/pkg/types/protobuf"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	apiutils "github.com/pole-io/pole-server/apis/pkg/utils"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"go.uber.org/zap"
)

// PublishLaneGroups 发布多个治理规则
func (s *Server) PublishGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
	case apimodel.RuleRelease_LosslessRules:
		return s.PublishLosslessRules(ctx, req)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}
}

func (s *Server) GetRuleReleases(ctx context.Context, filter map[string]string) *apimodel.BatchQueryResponse {
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
	case apimodel.RuleRelease_LosslessRules:
		total, versions, err = s.storage.GetLosslessRuleVersions(ctx, filter, offset, limit)
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
func (s *Server) DeleteGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	if len(req) == 0 {
		return api.NewBatchWriteResponse(apimodel.Code_EmptyRequest)
	}
	switch req[0].GetResource() {
	case apimodel.RuleRelease_LaneRules:
		return s.DeleteLaneGroupReleases(ctx, req)
	case apimodel.RuleRelease_CircuitBreakerRules:
		return s.DeleteCircuitBreakerReleases(ctx, req)
	case apimodel.RuleRelease_FaultDetectRules:
		return s.DeleteFaultDetectReleases(ctx, req)
	case apimodel.RuleRelease_RouteRules:
		return s.DeleteRouterReleases(ctx, req)
	case apimodel.RuleRelease_RateLimitRules:
		return s.DeleteRateLimitReleases(ctx, req)
	default:
		return api.NewBatchWriteResponse(apimodel.Code_InvalidParameter)
	}
}

// RollbackLaneGroups 回滚多个治理规则到目标版本
func (s *Server) RollbackGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) StopbetaGovernanceRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
	lock                  func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response)
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
	return &RuleReleasePipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if apiutils.IsNil(rule) {
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
			return !apiutils.IsNil(existRes), nil
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
			case *rules.LosslessRuleRelease:
				v.RuleRelease = *grayData
			}
			res, err := getActiveFn(tx, grayRelease)
			if err != nil {
				return false, err
			}
			return !apiutils.IsNil(res), nil
		},
		publish: func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) error {
			release := releaseBuilder(req, rule)
			return publishFn(tx, release)
		},
	}
}

// executeRuleReleasePipline 执行通用的治理规则灰度发布执行
func (s *Server) executeRuleReleasePipline(ctx context.Context, pipline *RuleReleasePipeline, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apimodel.Response {
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
		} else {
			// 如果是灰度发布版本，则需要检查下是否已经存在灰度发布版本，存在的话，不能重复发布
			grayRes, err := pipline.checkExistGrayRelease(ctx, tx, req)
			if err != nil {
				return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if grayRes {
				return api.NewResponse(apimodel.Code_ExistedResource)
			}
			// 保存灰度发布信息
			if errRsp := SaveGrayRule(ctx, tx, s.storage, curData); err != nil {
				log.Error("[goverrule][release] save gray rule when publish gray rule.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
					zap.String("rule-id", curData.RuleId), zap.String("rule-name", curData.RuleName), zap.String("error", errRsp.GetInfo().GetValue()))
				return errRsp
			}
		}

		// 发布规则
		if err := pipline.publish(ctx, tx, rule, req); err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] publish release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", curData.RuleId), zap.String("rule-name", curData.RuleName), zap.Error(err))
			return api.NewResponse(store.StoreCode2APICode(err))
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

// PublishCircuitBreakerRules implements GoverRuleServer.
func (s *Server) PublishCircuitBreakerRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) PublishFaultDetectRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) PublishLaneGroups(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
			var protoVal *rules.LaneGroupProto
			if rule != nil {
				laneRule := rule.(*rules.LaneGroup)
				protoVal, _ = laneRule.ToProto()
			}
			return &rules.LaneGroupRelease{
				RuleRelease: *curData,
				Rule:        protoVal,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

func (s *Server) PublishRateLimits(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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

func (s *Server) PublishLosslessRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleReleasePipeline(
		s.storage.LockLosslessRule,
		s.storage.GetReleaseLosslessRule,
		func(tx store.Tx, rel any) (*rules.LosslessRuleRelease, error) {
			return s.storage.GetActiveLosslessRule(tx, rel.(*rules.LosslessRuleRelease))
		},
		s.storage.PublishLosslessRules,
		func(req *apimodel.RuleRelease, rule any) *rules.LosslessRuleRelease {
			curData := &rules.RuleRelease{}
			curData.FromSpec(req)
			curData.Id = utils.NewUUID()
			var llRule *rules.LosslessRule
			if rule != nil {
				llRule = rule.(*rules.LosslessRule)
			}
			return &rules.LosslessRuleRelease{
				RuleRelease: *curData,
				Rule:        llRule,
			}
		},
	)
	return s.executeRuleReleasePipline(ctx, pipeline, requests)
}

func (s *Server) PublishRouterRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
	lock              func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response)
	checkExistRelease func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (*rules.RuleRelease, error)
	active            func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) *apimodel.Response
}

// NewRuleRollbackPipeline 工厂函数，简化各类规则回滚的 pipeline 构造
func NewRuleRollbackPipeline(
	lockFn func(store.Tx, string) (any, error),
	getReleaseFn func(store.Tx, *rules.RuleRelease) (any, error),
	activeFn func(store.Tx, any, any, *apimodel.RuleRelease) error,
) *RuleRollbackPipeline {
	return &RuleRollbackPipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if apiutils.IsNil(rule) {
				return nil, api.NewResponse(apimodel.Code_NotFoundResource)
			}
			return rule, nil
		},
		checkExistRelease: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (*rules.RuleRelease, error) {
			existRes, err := getReleaseFn(tx, &rules.RuleRelease{
				Id:          req.Id,
				ReleaseName: req.ReleaseName,
				RuleId:      req.RuleId,
				RuleName:    req.RuleName,
			})
			if err != nil || apiutils.IsNil(existRes) {
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
		active: func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) *apimodel.Response {
			existRes, err := getReleaseFn(tx, &rules.RuleRelease{
				Id:          req.Id,
				ReleaseName: req.ReleaseName,
				RuleId:      req.RuleId,
				RuleName:    req.RuleName,
			})
			if err != nil {
				return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if apiutils.IsNil(existRes) {
				return api.NewResponse(apimodel.Code_NotFoundResource)
			}
			if err := activeFn(tx, existRes, rule, req); err != nil {
				return api.NewResponse(store.StoreCode2APICode(err))
			}
			return api.NewResponse(apimodel.Code_ExecuteSuccess)
		},
	}
}

func (s *Server) exectueRuleRollbackPipline(ctx context.Context, pipline *RuleRollbackPipeline, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apimodel.Response {
		reqData := &rules.RuleRelease{}
		reqData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
		}
		defer tx.Rollback() // 最终的异常 case 的兜底

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

		rule, errRsp := pipline.lock(ctx, tx, existRes.ToSpec())
		if errRsp != nil {
			return errRsp
		}

		if errRsp := pipline.active(ctx, tx, rule, req); !api.IsSuccess(errRsp) {
			return errRsp
		}
		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] rollback release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", reqData.RuleId), zap.String("rule-name", reqData.RuleName), zap.String("target-version", req.ReleaseName),
				zap.Error(err))
			return api.NewResponse(store.StoreCode2APICode(err))
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

// RollbackCircuitBreakerRules 泛型重构示例
func (s *Server) RollbackCircuitBreakerRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) RollbackFaultDetectRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) RollbackLaneGroups(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) RollbackRateLimits(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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
func (s *Server) RollbackRouterRules(ctx context.Context, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
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

// RuleStopbetaPipeline 用于治理规则的停止灰度发布控制操作（停止灰度版本）
type RuleStopbetaPipeline struct {
	// lock 锁定指定规则
	lock func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response)
	// checkExist 检查灰度版本是否存在
	checkExist func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error)
	// inactive 将灰度版本置为未激活
	inactive func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) error
}

// NewRuleStopbetaPipeline 工厂函数，简化各类规则停止灰度的 pipeline 构造
func NewRuleStopbetaPipeline[
	Rule any, // 规则实体类型（Lock 返回）
	Release any, // 发布实体类型（Inactive 入参）
](
	lockFn func(store.Tx, string) (Rule, error),
	getReleaseFn func(store.Tx, *rules.RuleRelease) (Release, error),
	inactiveFn func(store.Tx, Release) error,
	releaseBuilder func(*apimodel.RuleRelease, Rule) Release,
) *RuleStopbetaPipeline {
	return &RuleStopbetaPipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if apiutils.IsNil(rule) {
				return nil, api.NewResponse(apimodel.Code_NotFoundResource)
			}
			return rule, nil
		},
		checkExist: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (bool, error) {
			// 校验灰度版本是否存在
			check := &rules.RuleRelease{}
			check.FromSpec(req)
			check.ReleaseType = rules.ReleaseTypeGray
			existRes, err := getReleaseFn(tx, check)
			if err != nil {
				return false, err
			}
			return !apiutils.IsNil(existRes), nil
		},
		inactive: func(ctx context.Context, tx store.Tx, rule any, req *apimodel.RuleRelease) error {
			release := releaseBuilder(req, rule.(Rule))
			return inactiveFn(tx, release)
		},
	}
}

// executeRuleStopbetaPipline 执行停止灰度流水线
func (s *Server) executeRuleStopbetaPipline(ctx context.Context, pipline *RuleStopbetaPipeline, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apimodel.Response {
		reqData := &rules.RuleRelease{}
		reqData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		defer tx.Rollback()

		// 1. 锁定规则
		rule, errRsp := pipline.lock(ctx, tx, req)
		if errRsp != nil {
			return errRsp
		}

		// 2. 校验灰度版本是否存在
		exist, err := pipline.checkExist(ctx, tx, req)
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		if !exist {
			return api.NewResponse(apimodel.Code_NotFoundResource)
		}

		// 3. 停止灰度（置为未激活）
		if err := pipline.inactive(ctx, tx, rule, req); err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}

		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] stopbeta release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", reqData.RuleId), zap.String("rule-name", reqData.RuleName), zap.Error(err))
			return api.NewResponse(store.StoreCode2APICode(err))
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

// StopbetaCircuitBreakerRules implements GoverRuleServer.
func (s *Server) StopbetaCircuitBreakerRules(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockCircuitBreakerRule,
		s.storage.GetReleaseCircuitBreakerRule,
		s.storage.InactiveCircuitBreakerRule,
		func(req *apimodel.RuleRelease, rule *rules.CircuitBreakerRule) *rules.CircuitBreakerRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(req)
			return &rules.CircuitBreakerRelease{RuleRelease: *cur, Rule: rule}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, request)
}

// StopbetaFaultDetectRules implements GoverRuleServer.
func (s *Server) StopbetaFaultDetectRules(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockFaultDetectRule,
		s.storage.GetReleaseFaultDetectRule,
		s.storage.InactiveFaultDetectRule,
		func(req *apimodel.RuleRelease, rule *rules.FaultDetectRule) *rules.FaultDetectRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(req)
			return &rules.FaultDetectRelease{RuleRelease: *cur, Rule: rule}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, request)
}

// StopbetaLaneGroups implements GoverRuleServer.
func (s *Server) StopbetaLaneGroups(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockLaneGroup,
		s.storage.GetReleaseLaneGroupRule,
		s.storage.InactiveLaneGroup,
		func(req *apimodel.RuleRelease, group *rules.LaneGroup) *rules.LaneGroupRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(req)
			protoVal, _ := group.ToProto()
			return &rules.LaneGroupRelease{RuleRelease: *cur, Rule: protoVal}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, req)
}

// StopbetaRateLimits implements GoverRuleServer.
func (s *Server) StopbetaRateLimits(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockRateLimitRule,
		s.storage.GetReleaseRateLimitRule,
		s.storage.InactiveRateLimitRule,
		func(req *apimodel.RuleRelease, rule *rules.RateLimit) *rules.RateLimitRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(req)
			return &rules.RateLimitRelease{RuleRelease: *cur, Rule: rule}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, request)
}

// StopbetaRouterRules implements GoverRuleServer.
func (s *Server) StopbetaRouterRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockRouterRule,
		s.storage.GetReleaseRouterRule,
		s.storage.InactiveRouterRule,
		func(req *apimodel.RuleRelease, rule *rules.RouterConfig) *rules.RouterRuleRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(req)
			pdata, _ := rule.ToExpendRoutingConfig()
			return &rules.RouterRuleRelease{RuleRelease: *cur, Rule: pdata}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, req)
}

func (s *Server) StopbetaLosslessRules(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleStopbetaPipeline(
		s.storage.LockLosslessRule,
		s.storage.GetReleaseLosslessRule,
		s.storage.InactiveLosslessRule,
		func(r *apimodel.RuleRelease, rule *rules.LosslessRule) *rules.LosslessRuleRelease {
			cur := &rules.RuleRelease{}
			cur.FromSpec(r)
			return &rules.LosslessRuleRelease{RuleRelease: *cur, Rule: rule}
		},
	)
	return s.executeRuleStopbetaPipline(ctx, pipeline, req)
}

// RuleDeletePipeline 用于治理规则的删除发布版本控制
type RuleDeletePipeline struct {
	// lock 锁定规则，避免并发修改
	lock func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response)
	// find 查找目标发布记录
	find func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, error)
	// isActive 判断发布是否处于激活态
	isActive func(rel any) bool
	// delete 删除目标发布记录
	delete func(ctx context.Context, tx store.Tx, rel any) error
}

// NewRuleDeletePipeline 通用删除流水线工厂
func NewRuleDeletePipeline[
	Rule any, // 锁定返回的规则实体
	Release any, // 查找与删除的发布实体
](
	lockFn func(store.Tx, string) (Rule, error),
	getReleaseFn func(store.Tx, *rules.RuleRelease) (Release, error),
	deleteFn func(store.Tx, Release) error,
) *RuleDeletePipeline {
	// 从具体的 Release 类型中提取 Active 状态
	isActive := func(rel any) bool {
		switch v := rel.(type) {
		case *rules.CircuitBreakerRelease:
			return v.Active
		case *rules.FaultDetectRelease:
			return v.Active
		case *rules.LaneGroupRelease:
			return v.Active
		case *rules.RouterRuleRelease:
			return v.Active
		case *rules.RateLimitRelease:
			return v.Active
		case *rules.LosslessRuleRelease:
			return v.Active
		default:
			return false
		}
	}
	return &RuleDeletePipeline{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, *apimodel.Response) {
			rule, err := lockFn(tx, req.RuleName)
			if err != nil {
				return nil, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error())
			}
			if apiutils.IsNil(rule) {
				return nil, api.NewResponse(apimodel.Code_NotFoundResource)
			}
			return rule, nil
		},
		find: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) (any, error) {
			// 依据三元组(name, rule_id, release_type)精确查找
			key := &rules.RuleRelease{}
			key.FromSpec(req)
			return getReleaseFn(tx, key)
		},
		isActive: isActive,
		delete: func(ctx context.Context, tx store.Tx, rel any) error {
			return deleteFn(tx, rel.(Release))
		},
	}
}

// executeRuleDeletePipeline 执行删除流水线
func (s *Server) executeRuleDeletePipeline(ctx context.Context, pipeline *RuleDeletePipeline, requests []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	handle := func(ctx context.Context, req *apimodel.RuleRelease) *apimodel.Response {
		reqData := &rules.RuleRelease{}
		reqData.FromSpec(req)

		tx, err := s.storage.StartTx()
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		defer tx.Rollback()

		// 1. 加锁目标规则
		if _, errRsp := pipeline.lock(ctx, tx, req); errRsp != nil {
			return errRsp
		}

		// 2. 查找目标发布记录
		rel, err := pipeline.find(ctx, tx, req)
		if err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}
		if rel == nil {
			// 遵循幂等：不存在视为成功
			return api.NewResponse(apimodel.Code_ExecuteSuccess)
		}

		// 3. 不允许删除当前激活版本
		if pipeline.isActive(rel) {
			return api.NewResponseWithMsg(apimodel.Code_BadRequest, "cannot delete active release")
		}

		// 4. 删除发布记录
		if err := pipeline.delete(ctx, tx, rel); err != nil {
			return api.NewResponseWithMsg(store.StoreCode2APICode(err), err.Error())
		}

		if err := tx.Commit(); err != nil {
			log.Error("[goverrule][release] delete release when commit tx.", utils.RequestID(ctx), zap.Any("resource", req.Resource),
				zap.String("rule-id", reqData.RuleId), zap.String("rule-name", reqData.RuleName), zap.Error(err))
			return api.NewResponse(store.StoreCode2APICode(err))
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

// DeleteCircuitBreakerReleases implements GoverRuleServer.
func (s *Server) DeleteCircuitBreakerReleases(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockCircuitBreakerRule,
		s.storage.GetReleaseCircuitBreakerRule,
		s.storage.DeleteCircuitBreakerReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, request)
}

// DeleteFaultDetectReleases implements GoverRuleServer.
func (s *Server) DeleteFaultDetectReleases(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockFaultDetectRule,
		s.storage.GetReleaseFaultDetectRule,
		s.storage.DeleteFaultDetectReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, request)
}

// DeleteLaneGroupReleases implements GoverRuleServer.
func (s *Server) DeleteLaneGroupReleases(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockLaneGroup,
		s.storage.GetReleaseLaneGroupRule,
		s.storage.DeleteLaneGroupReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, req)
}

// DeleteRateLimitReleases implements GoverRuleServer.
func (s *Server) DeleteRateLimitReleases(ctx context.Context, request []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockRateLimitRule,
		s.storage.GetReleaseRateLimitRule,
		s.storage.DeleteRateLimitReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, request)
}

// DeleteRouterReleases implements GoverRuleServer.
func (s *Server) DeleteRouterReleases(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockRouterRule,
		s.storage.GetReleaseRouterRule,
		s.storage.DeleteRouterRuleReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, req)
}

// DeleteLosslessReleases implements GoverRuleServer.
func (s *Server) DeleteLosslessReleases(ctx context.Context, req []*apimodel.RuleRelease) *apimodel.BatchWriteResponse {
	pipeline := NewRuleDeletePipeline(
		s.storage.LockLosslessRule,
		s.storage.GetReleaseLosslessRule,
		s.storage.DeleteLosslessReleases,
	)
	return s.executeRuleDeletePipeline(ctx, pipeline, req)
}

// SaveGrayRule
func SaveGrayRule(ctx context.Context, tx store.Tx, s store.GrayStore, rule rules.GrayRule) *apimodel.Response {
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
		return api.NewResponse(store.StoreCode2APICode(err))
	}
	return nil
}
