/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 */

package ai

import (
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

type a2aAgentCache struct {
	*cachebase.BaseCache
	storage store.Store

	ids            *container.SyncMap[string, *aitypes.A2AAgent]
	names          *container.SyncMap[string, *aitypes.A2AAgent]
	namespaceIndex *container.SyncMap[string, []*aitypes.A2AAgent]
	skills         *container.SyncMap[string, []*aitypes.A2AAgentSkill]

	singleFlight *singleflight.Group
}

func NewA2AAgentCache(storage store.Store, cacheMgr cacheapi.CacheManager) cacheapi.Cache {
	return &a2aAgentCache{
		BaseCache:      cachebase.NewBaseCache(storage, cacheMgr),
		storage:        storage,
		ids:            container.NewSyncMap[string, *aitypes.A2AAgent](),
		names:          container.NewSyncMap[string, *aitypes.A2AAgent](),
		namespaceIndex: container.NewSyncMap[string, []*aitypes.A2AAgent](),
		skills:         container.NewSyncMap[string, []*aitypes.A2AAgentSkill](),
		singleFlight:   &singleflight.Group{},
	}
}

func (c *a2aAgentCache) Name() string {
	return cacheapi.A2AAgentName
}

func (c *a2aAgentCache) Initialize(_ map[string]any) error {
	return nil
}

func (c *a2aAgentCache) Update() error {
	_, err, _ := c.singleFlight.Do(c.Name(), func() (any, error) {
		return nil, c.DoCacheUpdate(c.Name(), c.realUpdate)
	})
	return err
}

func (c *a2aAgentCache) realUpdate() (map[string]time.Time, int64, error) {
	results := make(map[string]time.Time)
	agents, err := c.storage.GetMoreA2AAgents(c.LastFetchTime(), c.IsFirstUpdate())
	if err != nil {
		return nil, 0, err
	}

	upsert := 0
	del := 0
	for _, agent := range agents {
		mtime := parseA2ATime(agent.Mtime)
		results[agent.Id] = mtime
		if agent.Flag == 1 {
			c.removeA2AAgent(agent)
			del++
			continue
		}
		c.storeA2AAgent(agent)
		upsert++
	}

	log.Infof("[Cache][A2AAgent] update agents, upsert: %d, delete: %d, total: %d", upsert, del, len(agents))
	return results, int64(len(agents)), nil
}

func (c *a2aAgentCache) Clear() error {
	c.BaseCache.Clear()
	c.ids = container.NewSyncMap[string, *aitypes.A2AAgent]()
	c.names = container.NewSyncMap[string, *aitypes.A2AAgent]()
	c.namespaceIndex = container.NewSyncMap[string, []*aitypes.A2AAgent]()
	c.skills = container.NewSyncMap[string, []*aitypes.A2AAgentSkill]()
	return nil
}

func (c *a2aAgentCache) Close() error {
	return nil
}

func (c *a2aAgentCache) GetA2AAgentByID(id string) *aitypes.A2AAgent {
	if id == "" {
		return nil
	}
	val, ok := c.ids.Load(id)
	if !ok {
		return nil
	}
	return val
}

func (c *a2aAgentCache) GetA2AAgentByName(name, namespace string) *aitypes.A2AAgent {
	if name == "" || namespace == "" {
		return nil
	}
	val, ok := c.names.Load(a2aAgentNameKey(namespace, name))
	if !ok {
		return nil
	}
	return val
}

func (c *a2aAgentCache) GetA2AAgentsByNamespace(namespace string) []*aitypes.A2AAgent {
	if namespace == "" {
		return nil
	}
	val, ok := c.namespaceIndex.Load(namespace)
	if !ok {
		return nil
	}
	return val
}

func (c *a2aAgentCache) GetA2AAgentSkills(agentID string) []*aitypes.A2AAgentSkill {
	if agentID == "" {
		return nil
	}
	val, ok := c.skills.Load(agentID)
	if !ok {
		return nil
	}
	return val
}

func (c *a2aAgentCache) Query(query *aitypes.A2AAgentQuery) (uint32, []*aitypes.A2AAgent) {
	if query == nil {
		query = &aitypes.A2AAgentQuery{}
	}
	if query.Limit == 0 {
		query.Limit = 100
	}

	matched := make([]*aitypes.A2AAgent, 0, 16)
	c.ids.Range(func(_ string, agent *aitypes.A2AAgent) {
		if agent == nil || !matchA2AAgent(query, agent) {
			return
		}
		matched = append(matched, agent)
	})

	sort.SliceStable(matched, func(i, j int) bool {
		return parseA2ATime(matched[i].Mtime).After(parseA2ATime(matched[j].Mtime))
	})

	total := uint32(len(matched))
	if query.Offset >= total {
		return total, []*aitypes.A2AAgent{}
	}
	end := query.Offset + query.Limit
	if end > total {
		end = total
	}
	return total, matched[query.Offset:end]
}

func (c *a2aAgentCache) storeA2AAgent(agent *aitypes.A2AAgent) {
	c.ids.Store(agent.Id, agent)
	c.names.Store(a2aAgentNameKey(agent.Namespace, agent.Name), agent)
	c.updateNamespaceIndex(agent)
	c.skills.Store(agent.Id, agent.Skills)
}

func (c *a2aAgentCache) removeA2AAgent(agent *aitypes.A2AAgent) {
	c.ids.Delete(agent.Id)
	c.names.Delete(a2aAgentNameKey(agent.Namespace, agent.Name))
	c.removeFromNamespaceIndex(agent)
	c.skills.Store(agent.Id, []*aitypes.A2AAgentSkill{})
}

func (c *a2aAgentCache) updateNamespaceIndex(agent *aitypes.A2AAgent) {
	agents, _ := c.namespaceIndex.ComputeIfAbsent(agent.Namespace, func(string) []*aitypes.A2AAgent {
		return make([]*aitypes.A2AAgent, 0)
	})
	for i, item := range agents {
		if item.Id == agent.Id {
			agents = append(agents[:i], agents[i+1:]...)
			break
		}
	}
	agents = append(agents, agent)
	c.namespaceIndex.Store(agent.Namespace, agents)
}

func (c *a2aAgentCache) removeFromNamespaceIndex(agent *aitypes.A2AAgent) {
	agents, ok := c.namespaceIndex.Load(agent.Namespace)
	if !ok {
		return
	}
	for i, item := range agents {
		if item.Id == agent.Id {
			agents = append(agents[:i], agents[i+1:]...)
			break
		}
	}
	c.namespaceIndex.Store(agent.Namespace, agents)
}

func matchA2AAgent(query *aitypes.A2AAgentQuery, agent *aitypes.A2AAgent) bool {
	if query.Name != "" && !strings.HasPrefix(agent.Name, query.Name) {
		return false
	}
	if query.Namespace != "" && agent.Namespace != query.Namespace {
		return false
	}
	if query.Business != "" && agent.Business != query.Business {
		return false
	}
	if query.Department != "" && agent.Department != query.Department {
		return false
	}
	if query.ProtocolBinding != "" && agent.PreferredProtocolBinding != query.ProtocolBinding {
		return false
	}
	if query.BackendType != "" && agent.BackendType != query.BackendType {
		return false
	}
	if query.BackendServiceNamespace != "" && agent.BackendServiceNamespace != query.BackendServiceNamespace {
		return false
	}
	if query.BackendServiceName != "" && agent.BackendServiceName != query.BackendServiceName {
		return false
	}
	if query.Streaming != nil && agent.Streaming != *query.Streaming {
		return false
	}
	if query.PushNotifications != nil && agent.PushNotifications != *query.PushNotifications {
		return false
	}
	if query.SkillTag != "" && !agentHasSkillTag(agent, query.SkillTag) {
		return false
	}
	return true
}

func agentHasSkillTag(agent *aitypes.A2AAgent, tag string) bool {
	for _, skill := range agent.Skills {
		for _, item := range skill.Tags {
			if item == tag {
				return true
			}
		}
	}
	return false
}

func a2aAgentNameKey(namespace, name string) string {
	return namespace + "/" + name
}

func formatA2ATime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(mcpTimeLayout)
}

func parseA2ATime(value string) time.Time {
	return parseMCPTime(value)
}
