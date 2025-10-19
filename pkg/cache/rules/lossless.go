package rules

import (
	"context"
	"time"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	cachetypes "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	cachebase "github.com/pole-io/pole-server/pkg/cache/base"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"go.uber.org/zap"
)

var (
	_ cachetypes.LosslessCache = (*LossLessCache)(nil)
)

type LossLessCache struct {
	*cachebase.BaseCache

	// 用于控制台查询列表
	ids *container.SyncMap[string, *rules.LosslessRule]
	// --------- 以下缓存均用于客户端数据查询 --------- //
	// increment cache
	rules *container.SyncMap[string, *rules.LosslessRuleRelease]
}

func NewLossLessCache(s store.Store, cacheMgr cachetypes.CacheManager) cachetypes.Cache {
	return &LossLessCache{
		BaseCache: cachebase.NewBaseCache(s, cacheMgr),
	}
}

func (llc *LossLessCache) Initialize(_ map[string]any) error {
	llc.ids = container.NewSyncMap[string, *rules.LosslessRule]()
	llc.rules = container.NewSyncMap[string, *rules.LosslessRuleRelease]()
	return nil
}

func (llc *LossLessCache) LastMtime() time.Time {
	return llc.BaseCache.LastMtime(llc.Name())
}

func (llc *LossLessCache) Update() error {
	// 多个线程竞争，只有一个线程进行更新
	err, _ := llc.singleUpdate()
	return err
}

func (llc *LossLessCache) singleUpdate() (error, bool) {
	// 多个线程竞争，只有一个线程进行更新
	_, err, shared := llc.GetSingle().Do(llc.Name(), func() (interface{}, error) {
		return nil, llc.DoCacheUpdate(llc.Name(), llc.realUpdate)
	})
	return err, shared
}

func (llc *LossLessCache) realUpdate() (map[string]time.Time, int64, error) {
	start := time.Now()

	// 获取泳道规则信息
	rules, err := llc.Store().GetMoreLosslessRules(llc.LastFetchTime(), llc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][lossless] query cache update err: %s", err.Error())
		return nil, -1, err
	}

	// 获取泳道发布信息
	// 这里的泳道发布信息是客户端发布的泳道规则
	// 需要注意的是，泳道规则的发布信息可能会有多个版本
	releases, err := llc.Store().GetMoreLosslessReleases(llc.LastFetchTime(), llc.IsFirstUpdate())
	if err != nil {
		log.Errorf("[cache][lossless] publish rule cache update err: %s", err.Error())
		return nil, -1, err
	}

	cmtime, addCnt, updateCnt, delCnt := llc.setRulesConsole(rules)
	log.Info("[cache][lossless] console cache update",
		zap.Int("pull-from-store", len(rules)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", llc.LastMtime()), zap.Duration("used", time.Since(start)))

	pmtime, addCnt, updateCnt, delCnt := llc.setRulesClient(releases)
	log.Info("[cache][lossless] client cache update",
		zap.Int("pull-from-store", len(releases)), zap.Int("add", addCnt), zap.Int("update", updateCnt),
		zap.Int("delete", delCnt), zap.Time("last", llc.LastMtime()), zap.Duration("used", time.Since(start)))

	return map[string]time.Time{
		llc.Name() + "_console": cmtime,
		llc.Name() + "_client":  pmtime,
	}, int64(len(rules)), err
}

func (llc *LossLessCache) setRulesConsole(rules []*rules.LosslessRule) (time.Time, int, int, int) {
	lastMtime := llc.LastMtime()
	add := 0
	update := 0
	del := 0

	for _, rule := range rules {
		if rule.MTime.Unix() > lastMtime.Unix() {
			lastMtime = rule.MTime
		}

		if !rule.Valid {
			if _, ok := llc.ids.Delete(rule.ID); ok {
				del++
			}
			continue
		}

		if _, ok := llc.ids.Load(rule.ID); ok {
			update++
		} else {
			add++
		}
		llc.ids.Store(rule.ID, rule)
	}

	return lastMtime, add, update, del
}

func (llc *LossLessCache) setRulesClient(rules []*rules.LosslessRuleRelease) (time.Time, int, int, int) {
	lastMtime := llc.LastMtime()
	add := 0
	update := 0
	del := 0

	for _, rule := range rules {
		if rule.Mtime.Unix() > lastMtime.Unix() {
			lastMtime = rule.Mtime
		}

		id := rule.ActiveKey()
		old, exist := llc.rules.Load(id)

		// 版本低于当前版本，不处理
		if exist && old.Version > rule.Version {
			continue
		}

		// 如果规则是删除操作
		if !rule.Valid {
			// 如果规则和之前的不一样，不需要处理
			if rule.Id != old.Id {
				continue
			}
			if _, ok := llc.rules.Delete(id); ok {
				del++
			}
			continue
		}

		if !rule.Active {
			if rule.Id == old.Id {
				if _, ok := llc.rules.Delete(id); ok {
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
		llc.rules.Store(rule.Id, rule)
	}

	return lastMtime, add, update, del
}

func (llc *LossLessCache) Name() string {
	return cachetypes.LossLessRuleName
}

func (llc *LossLessCache) Clear() error {
	llc.ids = container.NewSyncMap[string, *rules.LosslessRule]()
	llc.rules = container.NewSyncMap[string, *rules.LosslessRuleRelease]()
	return nil
}

// Query .
func (llc *LossLessCache) Query(ctx context.Context, args *cachetypes.LosslessArgs) (uint32, []*rules.LosslessRule, error) {
	if err := llc.Update(); err != nil {
		return 0, nil, err
	}

	predicates := cacheapi.LoadLosslessRulePredicates(ctx)

	ruleId, hasRuleId := args.Filter["rule_id"]
	nsName, hasNamespace := args.Filter["namespace"]
	svcName, hasService := args.Filter["service"]

	res := make([]*rules.LosslessRule, 0, 8)
	process := func(rule *rules.LosslessRule) {
		if hasService && svcName != rule.Proto.GetService() {
			return
		}
		if hasNamespace && nsName != rule.Proto.GetNamespace() {
			return
		}
		if hasRuleId && ruleId != rule.ID {
			return
		}
		for i := range predicates {
			if !predicates[i](ctx, rule) {
				return
			}
		}

		res = append(res, rule)
	}
	llc.ids.Range(func(key string, value *rules.LosslessRule) {
		process(value)
	})
	amount, items := cachebase.SortBeforeTrim(res, args.Filter["order_type"], args.Offset, args.Limit)
	return amount, items, nil
}

// GetLosslessConfig 根据ServiceID获取无损配置
func (llc *LossLessCache) GetLosslessConfig(svcName string, namespace string) *rules.LosslessRule {
	return nil
}

// GetRule 获取规则 ID 获取无损规则
func (llc *LossLessCache) GetRule(id string) *rules.LosslessRule {
	val, _ := llc.ids.Load(id)
	return val
}
