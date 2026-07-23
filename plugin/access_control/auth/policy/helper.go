/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"context"
	"time"

	apisecurity "github.com/pole-io/specification/source/go/api/v1/security"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	defaultMainUserPolicyComment  = "default main user auth policy rule"
	defaultReadWritePolicyComment = "global resources read and write"
	defaultReadOnlyPolicyComment  = "global resources read only policy rule"
)

type DefaultPolicyHelper struct {
	options  *AuthConfig
	storage  store.Store
	cacheMgr cachetypes.CacheManager
	checker  authapi.AuthChecker
}

type defaultPolicyResourceStore interface {
	GetMoreStrategies(mtime time.Time, firstUpdate bool) ([]*authtypes.StrategyDetail, error)
	LooseAddStrategyResources(resources []authtypes.StrategyResource) error
}

func (h *DefaultPolicyHelper) GetRole(id string) *authtypes.Role {
	return h.cacheMgr.Role().GetRole(id)
}

func (h *DefaultPolicyHelper) GetPolicyRule(id string) *authtypes.StrategyDetail {
	return h.cacheMgr.AuthStrategy().GetPolicyRule(id)
}

// CreatePrincipal 创建 principal 的默认 policy 资源
func (h *DefaultPolicyHelper) CreatePrincipalPolicy(ctx context.Context, tx store.Tx, p authtypes.Principal) error {
	if p.PrincipalType == authtypes.PrincipalUser && authapi.IsInitMainUser(ctx) {
		// 创建的是管理员帐户策略
		if err := h.storage.AddStrategy(tx, mainUserPrincipalPolicy(p)); err != nil {
			return err
		}
		// 创建默认策略
		policies := []*authtypes.StrategyDetail{defaultReadWritePolicy(p), defaultReadOnlyPolicy(p)}
		for i := range policies {
			if err := h.storage.AddStrategy(tx, policies[i]); err != nil {
				return err
			}
		}
		return nil
	}
	return h.storage.AddStrategy(tx, defaultPrincipalPolicy(p))
}

func mainUserPrincipalPolicy(p authtypes.Principal) *authtypes.StrategyDetail {
	// Create the user's default weight policy
	ruleId := utils.NewUUID()

	resources := []authtypes.StrategyResource{}

	for _, v := range apisecurity.ResourceType_value {
		resources = append(resources, authtypes.StrategyResource{
			StrategyID: ruleId,
			ResType:    apisecurity.ResourceType(v),
			ResID:      "*",
		})
	}

	calleeMethods := []string{"*"}
	return &authtypes.StrategyDetail{
		ID:            ruleId,
		Name:          authtypes.BuildDefaultStrategyName(p.PrincipalType, p.Name),
		Action:        apisecurity.AuthAction_ALLOW.String(),
		Default:       true,
		Revision:      utils.NewUUID(),
		Source:        "pole-io",
		Resources:     resources,
		Principals:    []authtypes.Principal{p},
		CalleeMethods: calleeMethods,
		Valid:         true,
		Comment:       defaultMainUserPolicyComment,
	}
}

func defaultReadWritePolicy(p authtypes.Principal) *authtypes.StrategyDetail {
	// Create the user's default weight policy
	ruleId := utils.NewUUID()

	resources := []authtypes.StrategyResource{}

	for _, v := range apisecurity.ResourceType_value {
		resources = append(resources, authtypes.StrategyResource{
			StrategyID: ruleId,
			ResType:    apisecurity.ResourceType(v),
			ResID:      "*",
		})
	}

	calleeMethods := []string{"*"}
	return &authtypes.StrategyDetail{
		ID:            ruleId,
		Name:          "全局读写策略",
		Action:        apisecurity.AuthAction_ALLOW.String(),
		Default:       true,
		Revision:      utils.NewUUID(),
		Source:        "pole-io",
		Resources:     resources,
		CalleeMethods: calleeMethods,
		Valid:         true,
		Comment:       defaultReadWritePolicyComment,
		Metadata: map[string]string{
			authtypes.MetadKeySystemDefaultPolicy: "true",
		},
	}
}

func defaultReadOnlyPolicy(p authtypes.Principal) *authtypes.StrategyDetail {
	// Create the user's default weight policy
	ruleId := utils.NewUUID()

	resources := []authtypes.StrategyResource{}

	for _, v := range apisecurity.ResourceType_value {
		resources = append(resources, authtypes.StrategyResource{
			StrategyID: ruleId,
			ResType:    apisecurity.ResourceType(v),
			ResID:      "*",
		})
	}

	calleeMethods := []string{
		"Describe*",
		"List*",
		"Get*",
	}
	return &authtypes.StrategyDetail{
		ID:            ruleId,
		Name:          "全局只读策略",
		Action:        apisecurity.AuthAction_ALLOW.String(),
		Default:       true,
		Revision:      utils.NewUUID(),
		Source:        "pole-io",
		Resources:     resources,
		CalleeMethods: calleeMethods,
		Valid:         true,
		Comment:       defaultReadOnlyPolicyComment,
		Metadata: map[string]string{
			authtypes.MetadKeySystemDefaultPolicy: "true",
		},
	}
}

// backfillDefaultPolicyAIResourceWildcards 为历史全量默认策略补齐后续新增的 AI 资源类型。
// 默认策略在创建时会枚举当时已知的 ResourceType；MCP/A2A 后加入时，已有策略不会自动拥有它们的通配授权。
func backfillDefaultPolicyAIResourceWildcards(storage defaultPolicyResourceStore) (int, error) {
	policies, err := storage.GetMoreStrategies(time.Time{}, true)
	if err != nil {
		return 0, err
	}

	missingResources := make([]authtypes.StrategyResource, 0)
	for _, policy := range policies {
		if !isAllResourceDefaultPolicy(policy) {
			continue
		}

		resourceTypes := make(map[apisecurity.ResourceType]struct{}, len(policy.Resources))
		for _, resource := range policy.Resources {
			if resource.ResID == "*" {
				resourceTypes[resource.ResType] = struct{}{}
			}
		}

		for _, resourceType := range []apisecurity.ResourceType{
			apisecurity.ResourceType_MCPServerResources,
			apisecurity.ResourceType_A2AAgentResources,
		} {
			if _, exists := resourceTypes[resourceType]; exists {
				continue
			}
			missingResources = append(missingResources, authtypes.StrategyResource{
				StrategyID: policy.ID,
				ResType:    resourceType,
				ResID:      "*",
			})
		}
	}

	if len(missingResources) == 0 {
		return 0, nil
	}
	if err := storage.LooseAddStrategyResources(missingResources); err != nil {
		return 0, err
	}
	return len(missingResources), nil
}

func isAllResourceDefaultPolicy(policy *authtypes.StrategyDetail) bool {
	if policy == nil || !policy.Default || !policy.Valid {
		return false
	}
	switch policy.Comment {
	case defaultMainUserPolicyComment:
		return true
	case defaultReadWritePolicyComment, defaultReadOnlyPolicyComment:
		return policy.Metadata[authtypes.MetadKeySystemDefaultPolicy] == "true"
	default:
		return false
	}
}

func defaultPrincipalPolicy(p authtypes.Principal) *authtypes.StrategyDetail {
	// Create the user's default weight policy
	ruleId := utils.NewUUID()

	resources := []authtypes.StrategyResource{}
	calleeMethods := []string{
		// 用户操作权限
		string(authtypes.DescribeUsers),
		// 鉴权策略
		string(authtypes.DescribeAuthPolicies),
		string(authtypes.DescribeAuthPolicyDetail),
		// 角色
		string(authtypes.DescribeAuthRoles),
	}
	if p.PrincipalType == authtypes.PrincipalUser {
		resources = append(resources, authtypes.StrategyResource{
			StrategyID: ruleId,
			ResType:    apisecurity.ResourceType_Users,
			ResID:      p.PrincipalID,
		})
		calleeMethods = []string{
			// 用户操作权限
			string(authtypes.DescribeUsers),
			string(authtypes.DescribeUserToken),
			string(authtypes.UpdateUser),
			string(authtypes.UpdateUserPassword),
			string(authtypes.EnableUserToken),
			string(authtypes.ResetUserToken),
			// 鉴权策略
			string(authtypes.DescribeAuthPolicies),
			string(authtypes.DescribeAuthPolicyDetail),
			// 角色
			string(authtypes.DescribeAuthRoles),
		}
	}

	return &authtypes.StrategyDetail{
		ID:            ruleId,
		Name:          authtypes.BuildDefaultStrategyName(p.PrincipalType, p.Name),
		Action:        apisecurity.AuthAction_ALLOW.String(),
		Default:       true,
		Revision:      utils.NewUUID(),
		Source:        "pole-io",
		Resources:     resources,
		Principals:    []authtypes.Principal{p},
		CalleeMethods: calleeMethods,
		Valid:         true,
		Comment:       "default principal auth policy rule",
	}
}

// CleanPrincipal 清理 principal 所关联的 policy、role 资源
func (h *DefaultPolicyHelper) CleanPrincipal(ctx context.Context, tx store.Tx, p authtypes.Principal) error {
	if err := h.storage.CleanPrincipalPolicies(tx, p); err != nil {
		return err
	}

	if err := h.storage.CleanPrincipalRoles(tx, &p); err != nil {
		return err
	}
	return nil
}
