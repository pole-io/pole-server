/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, Tencent. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package rules

import (
	"sync"
	"time"

	cachetypes "github.com/pole-io/pole-server/apis/cache"
	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
)

type governanceRuleUpdateStore interface {
	GetUnixSecond(offset int64) (int64, error)
	GetMoreGovernanceRuleUpdates(mtime time.Time, firstUpdate bool) (*ruletypes.GovernanceRuleUpdates, error)
	GetMoreGovernanceRuleReleaseUpdates(mtime time.Time, firstUpdate bool) (*ruletypes.GovernanceRuleReleaseUpdates, error)
}

type governanceRuleUpdateWatcher interface {
	applyGovernanceRuleUpdate(*ruletypes.GovernanceRuleUpdates, *ruletypes.GovernanceRuleReleaseUpdates) (int64, error)
}

type governanceRuleUpdateCache struct {
	lock          sync.Mutex
	storage       governanceRuleUpdateStore
	cacheMgr      cachetypes.CacheManager
	firstUpdate   bool
	lastFetchTime int64
	watchers      []governanceRuleUpdateWatcher
}

var governanceRuleUpdater *governanceRuleUpdateCache
var governanceRuleUpdaterLock sync.Mutex

func registerGovernanceRuleWatcher(storage any, cacheMgr cachetypes.CacheManager, watcher governanceRuleUpdateWatcher) bool {
	updateStore, ok := storage.(governanceRuleUpdateStore)
	if !ok {
		return false
	}
	governanceRuleUpdaterLock.Lock()
	defer governanceRuleUpdaterLock.Unlock()
	if governanceRuleUpdater == nil || governanceRuleUpdater.storage != updateStore {
		governanceRuleUpdater = &governanceRuleUpdateCache{
			storage:       updateStore,
			cacheMgr:      cacheMgr,
			firstUpdate:   true,
			lastFetchTime: 1,
		}
	}
	governanceRuleUpdater.watchers = append(governanceRuleUpdater.watchers, watcher)
	return true
}

func updateGovernanceRuleCache(storage any) (bool, error) {
	updateStore, ok := storage.(governanceRuleUpdateStore)
	if !ok {
		return false, nil
	}
	governanceRuleUpdaterLock.Lock()
	updater := governanceRuleUpdater
	governanceRuleUpdaterLock.Unlock()
	if updater == nil || updater.storage != updateStore {
		return false, nil
	}
	return true, updater.Update()
}

func resetGovernanceRuleUpdateCache(storage any) {
	updateStore, ok := storage.(governanceRuleUpdateStore)
	if !ok {
		return
	}
	governanceRuleUpdaterLock.Lock()
	defer governanceRuleUpdaterLock.Unlock()
	if governanceRuleUpdater == nil || governanceRuleUpdater.storage != updateStore {
		return
	}
	governanceRuleUpdater.lock.Lock()
	defer governanceRuleUpdater.lock.Unlock()
	governanceRuleUpdater.firstUpdate = true
	governanceRuleUpdater.lastFetchTime = 1
}

func (g *governanceRuleUpdateCache) Update() error {
	g.lock.Lock()
	defer g.lock.Unlock()

	storeTime, err := g.storage.GetUnixSecond(0)
	if err != nil {
		storeTime = g.lastFetchTime
	}
	if !g.firstUpdate && storeTime <= g.lastFetchTime {
		return nil
	}

	lastFetch := time.Unix(g.lastFetchTime, 0)
	if g.cacheMgr != nil {
		lastFetch = lastFetch.Add(g.cacheMgr.GetTimeDiff())
		if lastFetch.Before(time.Unix(0, 0)) {
			lastFetch = time.Unix(g.lastFetchTime, 0)
		}
	}

	updates, err := g.storage.GetMoreGovernanceRuleUpdates(lastFetch, g.firstUpdate)
	if err != nil {
		return err
	}
	releases, err := g.storage.GetMoreGovernanceRuleReleaseUpdates(lastFetch, g.firstUpdate)
	if err != nil {
		return err
	}
	if updates == nil {
		updates = &ruletypes.GovernanceRuleUpdates{}
	}
	if releases == nil {
		releases = &ruletypes.GovernanceRuleReleaseUpdates{}
	}
	for i := range g.watchers {
		if _, err := g.watchers[i].applyGovernanceRuleUpdate(updates, releases); err != nil {
			return err
		}
	}
	g.lastFetchTime = storeTime
	g.firstUpdate = false
	return nil
}

func (rc *RouteRuleCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	rc.setRouterRuleConsole(updates.RouterRules)
	rc.setRouterRuleClient(releases.RouterRules)
	rc.container.reload()
	return int64(len(updates.RouterRules) + len(releases.RouterRules)), nil
}

func (rlc *rateLimitCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	rlc.setRateLimitConsole(updates.RateLimitRules)
	rlc.setRateLimitClient(releases.RateLimitRules)
	return int64(len(updates.RateLimitRules) + len(releases.RateLimitRules)), nil
}

func (c *circuitBreakerCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	c.setCircuitBreakerConsole(updates.CircuitBreakerRules)
	c.setCircuitBreakerClient(releases.CircuitBreakerRules)
	return int64(len(updates.CircuitBreakerRules) + len(releases.CircuitBreakerRules)), nil
}

func (f *faultDetectCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	f.setFaultDetectConsole(updates.FaultDetectRules)
	f.setFaultDetectClient(releases.FaultDetectRules)
	return int64(len(updates.FaultDetectRules) + len(releases.FaultDetectRules)), nil
}

func (lc *LaneCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	lc.setLaneRulesConsole(updates.LaneGroups)
	lc.setLaneRulesClient(releases.LaneGroupRules)
	return int64(len(updates.LaneGroups) + len(releases.LaneGroupRules)), nil
}

func (llc *LossLessCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	llc.setRulesConsole(updates.LosslessRules)
	llc.setRulesClient(releases.LosslessRules)
	return int64(len(updates.LosslessRules) + len(releases.LosslessRules)), nil
}

func (c *TrafficGovernanceCache) applyGovernanceRuleUpdate(
	updates *ruletypes.GovernanceRuleUpdates, releases *ruletypes.GovernanceRuleReleaseUpdates,
) (int64, error) {
	switch c.name {
	case cachetypes.TrafficSecurityRuleName:
		c.setRulesConsole(updates.TrafficSecurityRules)
		c.setRulesClient(releases.TrafficSecurityRules)
		return int64(len(updates.TrafficSecurityRules) + len(releases.TrafficSecurityRules)), nil
	case cachetypes.TrafficMirrorRuleName:
		c.setRulesConsole(updates.TrafficMirrorRules)
		c.setRulesClient(releases.TrafficMirrorRules)
		return int64(len(updates.TrafficMirrorRules) + len(releases.TrafficMirrorRules)), nil
	case cachetypes.TrafficMockRuleName:
		c.setRulesConsole(updates.TrafficMockRules)
		c.setRulesClient(releases.TrafficMockRules)
		return int64(len(updates.TrafficMockRules) + len(releases.TrafficMockRules)), nil
	default:
		return 0, nil
	}
}
