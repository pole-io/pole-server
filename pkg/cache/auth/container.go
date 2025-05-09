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

package auth

import (
	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"

	authtypes "github.com/pole-io/pole-server/apis/pkg/types/auth"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

// PrincipalResourceContainer principal 资源容器
type PrincipalResourceContainer struct {
	denyResources  *container.SyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]]
	allowResources *container.SyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]]
}

// NewPrincipalResourceContainer 创建 PrincipalResourceContainer 对象
func NewPrincipalResourceContainer() *PrincipalResourceContainer {
	return &PrincipalResourceContainer{
		allowResources: container.NewSyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]](),
		denyResources:  container.NewSyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]](),
	}
}

// Hint 返回该资源命中的策略类型, 优先匹配 deny, 其次匹配 allow, 否则返回 deny
func (p *PrincipalResourceContainer) Hint(rt apisecurity.ResourceType, resId string) (apisecurity.AuthAction, bool) {
	ids, ok := p.denyResources.Load(rt)
	if ok {
		if ids.Contains(utils.MatchAll) {
			return apisecurity.AuthAction_DENY, true
		}
		if ids.Contains(resId) {
			return apisecurity.AuthAction_DENY, true
		}
	}
	ids, ok = p.allowResources.Load(rt)
	if ok {
		if ids.Contains(utils.MatchAll) {
			return apisecurity.AuthAction_ALLOW, true
		}
		if ids.Contains(resId) {
			return apisecurity.AuthAction_ALLOW, true
		}
	}
	return 0, false
}

// SaveResource 保存资源
func (p *PrincipalResourceContainer) SaveResource(a apisecurity.AuthAction, r authtypes.StrategyResource) {
	if a == apisecurity.AuthAction_ALLOW {
		p.saveResource(p.allowResources, r)
	} else {
		p.saveResource(p.denyResources, r)
	}
}

// DelResource 删除资源
func (p *PrincipalResourceContainer) DelResource(a apisecurity.AuthAction, r authtypes.StrategyResource) {
	if a == apisecurity.AuthAction_ALLOW {
		p.delResource(p.allowResources, r)
	} else {
		p.delResource(p.denyResources, r)
	}
}

func (p *PrincipalResourceContainer) saveResource(
	resContainer *container.SyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]], res authtypes.StrategyResource) {

	resType := apisecurity.ResourceType(res.ResType)
	resContainer.ComputeIfAbsent(resType, func(k apisecurity.ResourceType) *container.RefSyncSet[string, string] {
		return container.NewRefSyncSet[string, string]()
	})

	ids, _ := resContainer.Load(resType)
	ids.Add(container.Reference[string, string]{
		Key:        res.ResID,
		Referencer: res.StrategyID,
	})
}

func (p *PrincipalResourceContainer) delResource(
	resContainer *container.SyncMap[apisecurity.ResourceType, *container.RefSyncSet[string, string]], res authtypes.StrategyResource) {

	resType := apisecurity.ResourceType(res.ResType)
	resContainer.ComputeIfAbsent(resType, func(k apisecurity.ResourceType) *container.RefSyncSet[string, string] {
		return container.NewRefSyncSet[string, string]()
	})

	ids, _ := resContainer.Load(resType)
	ids.Remove(container.Reference[string, string]{
		Key:        res.ResID,
		Referencer: res.StrategyID,
	})
}
