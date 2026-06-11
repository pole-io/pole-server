package rules

import (
	"context"
	"sort"
	"time"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	ruletypes "github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/pkg/utils/revision"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"go.uber.org/zap"
)

var _ cachetypes.TrafficGovernanceCache = (*TrafficGovernanceCache)(nil)

type TrafficGovernanceCache struct {
	*cachebase.BaseCache

	name        string
	logName     string
	loadRules   func(time.Time, bool) ([]*ruletypes.TrafficGovernanceRule, error)
	loadRelease func(time.Time, bool) ([]*ruletypes.TrafficGovernanceRuleRelease, error)

	ids   *container.SyncMap[string, *ruletypes.TrafficGovernanceRule]
	rules *container.SyncMap[string, *ruletypes.TrafficGovernanceRuleRelease]
}

func NewTrafficSecurityCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.Cache {
	return newTrafficGovernanceCache(s, cacheMgr, cachetypes.TrafficSecurityRuleName, "traffic-security",
		s.GetMoreTrafficSecurityRules, s.GetMoreTrafficSecurityReleases)
}

func NewTrafficMirrorCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.Cache {
	return newTrafficGovernanceCache(s, cacheMgr, cachetypes.TrafficMirrorRuleName, "traffic-mirror",
		s.GetMoreTrafficMirrorRules, s.GetMoreTrafficMirrorReleases)
}

func NewTrafficMockCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.Cache {
	return newTrafficGovernanceCache(s, cacheMgr, cachetypes.TrafficMockRuleName, "traffic-mock",
		s.GetMoreTrafficMockRules, s.GetMoreTrafficMockReleases)
}

func newTrafficGovernanceCache(
	s store.Store,
	cacheMgr cachetypes.CacheManager,
	name string,
	logName string,
	loadRules func(time.Time, bool) ([]*ruletypes.TrafficGovernanceRule, error),
	loadRelease func(time.Time, bool) ([]*ruletypes.TrafficGovernanceRuleRelease, error),
) *TrafficGovernanceCache {
	return &TrafficGovernanceCache{
		BaseCache:   cachebase.NewBaseCache(s, cacheMgr),
		name:        name,
		logName:     logName,
		loadRules:   loadRules,
		loadRelease: loadRelease,
	}
}

func (c *TrafficGovernanceCache) Initialize(_ map[string]any) error {
	c.ids = container.NewSyncMap[string, *ruletypes.TrafficGovernanceRule]()
	c.rules = container.NewSyncMap[string, *ruletypes.TrafficGovernanceRuleRelease]()
	registerGovernanceRuleWatcher(c.Store(), c.CacheMgr, c)
	return nil
}

func (c *TrafficGovernanceCache) LastMtime() time.Time {
	return c.BaseCache.LastMtime(c.Name())
}

func (c *TrafficGovernanceCache) Update() error {
	if ok, err := updateGovernanceRuleCache(c.Store()); ok {
		return err
	}
	err, _ := c.singleUpdate()
	return err
}

func (c *TrafficGovernanceCache) singleUpdate() (error, bool) {
	_, err, shared := c.GetSingle().Do(c.Name(), func() (interface{}, error) {
		return nil, c.DoCacheUpdate(c.Name(), c.realUpdate)
	})
	return err, shared
}

func (c *TrafficGovernanceCache) realUpdate() (map[string]time.Time, int64, error) {
	start := time.Now()
	rules, err := c.loadRules(c.LastFetchTime(), c.IsFirstUpdate())
	if err != nil {
		return nil, -1, err
	}
	releases, err := c.loadRelease(c.LastFetchTime(), c.IsFirstUpdate())
	if err != nil {
		return nil, -1, err
	}
	cmtime, addCnt, updateCnt, delCnt := c.setRulesConsole(rules)
	log.Info("[cache]["+c.logName+"] console cache update",
		zap.Int("pull-from-store", len(rules)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", c.LastMtime()), zap.Duration("used", time.Since(start)))

	pmtime, addCnt, updateCnt, delCnt := c.setRulesClient(releases)
	log.Info("[cache]["+c.logName+"] client cache update",
		zap.Int("pull-from-store", len(releases)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", c.LastMtime()), zap.Duration("used", time.Since(start)))

	return map[string]time.Time{
		c.Name() + "_console": cmtime,
		c.Name() + "_client":  pmtime,
	}, int64(len(rules) + len(releases)), nil
}

func (c *TrafficGovernanceCache) setRulesConsole(rules []*ruletypes.TrafficGovernanceRule) (time.Time, int, int, int) {
	lastMtime := c.LastMtime()
	add, update, del := 0, 0, 0
	for _, rule := range rules {
		if rule.MTime.Unix() > lastMtime.Unix() {
			lastMtime = rule.MTime
		}
		if !rule.Valid {
			if _, ok := c.ids.Delete(rule.ID); ok {
				del++
			}
			continue
		}
		if _, ok := c.ids.Load(rule.ID); ok {
			update++
		} else {
			add++
		}
		c.ids.Store(rule.ID, rule)
	}
	return lastMtime, add, update, del
}

func (c *TrafficGovernanceCache) setRulesClient(releases []*ruletypes.TrafficGovernanceRuleRelease) (time.Time, int, int, int) {
	lastMtime := c.LastMtime()
	add, update, del := 0, 0, 0
	for _, release := range releases {
		if release.Mtime.Unix() > lastMtime.Unix() {
			lastMtime = release.Mtime
		}
		if release.Rule == nil {
			continue
		}
		id := release.ActiveKey()
		old, exist := c.rules.Load(id)
		if exist && old.Version > release.Version {
			continue
		}
		if !release.Valid {
			if exist && release.Id == old.Id {
				if _, ok := c.rules.Delete(id); ok {
					del++
				}
			}
			continue
		}
		if !release.Active {
			if exist && release.Id == old.Id {
				if _, ok := c.rules.Delete(id); ok {
					del++
				}
			}
			continue
		}
		if exist {
			update++
		} else {
			add++
		}
		c.rules.Store(id, release)
	}
	return lastMtime, add, update, del
}

func (c *TrafficGovernanceCache) Name() string {
	return c.name
}

func (c *TrafficGovernanceCache) Clear() error {
	resetGovernanceRuleUpdateCache(c.Store())
	c.ids = container.NewSyncMap[string, *ruletypes.TrafficGovernanceRule]()
	c.rules = container.NewSyncMap[string, *ruletypes.TrafficGovernanceRuleRelease]()
	return nil
}

func (c *TrafficGovernanceCache) Query(ctx context.Context, args *cachetypes.TrafficGovernanceArgs) (uint32, []*ruletypes.TrafficGovernanceRule, error) {
	if err := c.Update(); err != nil {
		return 0, nil, err
	}
	predicates := cacheapi.LoadTrafficGovernancePredicates(ctx)
	ruleID, hasRuleID := args.Filter["rule_id"]
	nsName, hasNamespace := args.Filter["namespace"]
	svcName, hasService := args.Filter["service"]
	name, hasName := args.Filter["name"]

	res := make([]*ruletypes.TrafficGovernanceRule, 0, 8)
	c.ids.Range(func(key string, value *ruletypes.TrafficGovernanceRule) {
		if hasRuleID && ruleID != value.ID {
			return
		}
		if hasNamespace && nsName != value.Namespace {
			return
		}
		if hasService && svcName != value.Service {
			return
		}
		if hasName && name != value.Name {
			return
		}
		for i := range predicates {
			if !predicates[i](ctx, value) {
				return
			}
		}
		res = append(res, value)
	})
	amount, items := cachebase.SortBeforeTrim(res, args.Filter["order_type"], args.Offset, args.Limit)
	return amount, items, nil
}

func (c *TrafficGovernanceCache) GetRule(id string) *ruletypes.TrafficGovernanceRule {
	val, _ := c.ids.Load(id)
	return val
}

func (c *TrafficGovernanceCache) GetRulesForService(namespace string, service string) ([]*ruletypes.TrafficGovernanceRule, string) {
	if err := c.Update(); err != nil {
		return nil, ""
	}
	rules := make([]*ruletypes.TrafficGovernanceRule, 0, 4)
	c.rules.Range(func(key string, value *ruletypes.TrafficGovernanceRuleRelease) {
		if value == nil || value.Rule == nil {
			return
		}
		if value.Rule.Namespace == namespace && value.Rule.Service == service {
			rules = append(rules, value.Rule)
		}
	})
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority == rules[j].Priority {
			return rules[i].ID < rules[j].ID
		}
		return rules[i].Priority < rules[j].Priority
	})
	revisions := make([]string, 0, len(rules))
	for i := range rules {
		if rules[i].Revision != "" {
			revisions = append(revisions, rules[i].Revision)
		}
	}
	if len(revisions) == 0 {
		return rules, ""
	}
	rev, _ := revision.CompositeComputeRevision(revisions)
	return rules, rev
}
