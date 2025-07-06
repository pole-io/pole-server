package goverrule

import (
	"context"

	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	apiservice "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

// RuleLocker 用于治理规则的并发控制
type RuleLocker struct {
	lock func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) error
}

// PublishCircuitBreakerRules implements GoverRuleServer.
func (s *Server) PublishCircuitBreakerRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	locker := &RuleLocker{
		lock: func(ctx context.Context, tx store.Tx, req *apimodel.RuleRelease) error {
			_, err := s.storage.LockRouterRule(tx, req.RuleName)
			return err
		},
	}
	return s.publishGovernanceRules(ctx, locker, request)
}

// PublishFaultDetectRules implements GoverRuleServer.
func (s *Server) PublishFaultDetectRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishLaneGroups implements GoverRuleServer.
func (s *Server) PublishLaneGroups(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishRateLimits implements GoverRuleServer.
func (s *Server) PublishRateLimits(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// PublishRouterRules implements GoverRuleServer.
func (s *Server) PublishRouterRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackCircuitBreakerRules implements GoverRuleServer.
func (s *Server) RollbackCircuitBreakerRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackFaultDetectRules implements GoverRuleServer.
func (s *Server) RollbackFaultDetectRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackLaneGroups implements GoverRuleServer.
func (s *Server) RollbackLaneGroups(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackRateLimits implements GoverRuleServer.
func (s *Server) RollbackRateLimits(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// RollbackRouterRules implements GoverRuleServer.
func (s *Server) RollbackRouterRules(ctx context.Context, req []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
}

// StopbetaCircuitBreakerRules implements GoverRuleServer.
func (s *Server) StopbetaCircuitBreakerRules(ctx context.Context, request []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	panic("unimplemented")
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

// publishGovernanceRules 通用逻辑，发布多个治理规则版本
func (s *Server) publishGovernanceRules(ctx context.Context, locker *RuleLocker, reqs []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	return nil
}

// removeGovernanceRules 通用逻辑，删除多个治理规则发布版本
func (s *Server) removeGovernanceRules(ctx context.Context, reqs []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	return nil
}

// rollbackGovernanceRules 通用逻辑，回滚多个治理规则发布版本
func (s *Server) rollbackGovernanceRules(ctx context.Context, reqs []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	return nil
}

// stopbetaGovernanceRules 通用逻辑，停止多个治理规则灰度发布版本
func (s *Server) stopbetaGovernanceRules(ctx context.Context, reqs []*apimodel.RuleRelease) *apiservice.BatchWriteResponse {
	return nil
}
