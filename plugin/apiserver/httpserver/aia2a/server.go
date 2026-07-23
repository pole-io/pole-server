/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 */

package aia2a

import (
	"context"
	"time"

	restful "github.com/emicklei/go-restful/v3"

	authapi "github.com/pole-io/pole-server/apis/access_control/auth"
	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/cache"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
)

const (
	defaultAccess string = "default"
	aia2aAccess   string = "aia2a"

	basePath string = "/ai/a2a/v1"
)

type HTTPServer struct {
	storage   store.Store
	cacheMgr  cacheapi.CacheManager
	policySvr authapi.StrategyServer
}

func NewServer(ctx context.Context, storage store.Store) (*HTTPServer, error) {
	cacheMgr, err := cache.GetCacheManager()
	if err != nil {
		commonlog.Errorf("set cache manager to ai-a2a server error. %v", err)
		return nil, err
	}
	if err := cacheMgr.OpenResourceCache(cacheapi.ConfigEntry{Name: cacheapi.A2AAgentName}); err != nil {
		commonlog.Errorf("open a2a-agent cache error. %v", err)
		return nil, err
	}
	if err := startA2AAgentCache(ctx, cacheMgr.A2AAgent(), cacheMgr.GetUpdateCacheInterval()); err != nil {
		commonlog.Errorf("start a2a-agent cache error. %v", err)
		return nil, err
	}
	policySvr, err := authapi.GetStrategyServer()
	if err != nil {
		commonlog.Errorf("set policy server to ai-a2a server error. %v", err)
		return nil, err
	}
	return &HTTPServer{storage: storage, cacheMgr: cacheMgr, policySvr: policySvr}, nil
}

func startA2AAgentCache(ctx context.Context, agentCache cacheapi.A2AAgentCache, interval time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = time.Second
	}
	if err := agentCache.Update(); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := agentCache.Update(); err != nil {
					commonlog.Warnf("update a2a-agent cache error. %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (h *HTTPServer) GetA2AAccessServer(include []string) *restful.WebService {
	commonlog.Info("enable ai-a2a access server")
	ws := new(restful.WebService)
	ws.Path(basePath).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	if len(include) == 0 {
		include = []string{defaultAccess}
	}
	for _, item := range include {
		switch item {
		case aia2aAccess, defaultAccess:
			h.addDefaultAccess(ws)
		}
	}
	return ws
}

func (h *HTTPServer) addDefaultAccess(ws *restful.WebService) {
	ws.Route(ws.GET("/agents").To(h.ListA2AAgents))
	ws.Route(ws.POST("/agents").To(h.CreateA2AAgents))
	ws.Route(ws.PUT("/agents").To(h.UpdateA2AAgents))
	ws.Route(ws.POST("/agents/delete").To(h.DeleteA2AAgents))
	ws.Route(ws.GET("/agent/skills").To(h.ListA2AAgentSkills))
	ws.Route(ws.GET("/agents/{id}/card").To(h.GetA2AAgentCard))
}
