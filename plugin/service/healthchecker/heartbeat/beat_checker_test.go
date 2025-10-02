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

package heartbeat

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/service/healthcheck"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

var (
	nodes = []healthcheck.CheckerPeer{
		{
			Host: "127.0.0.1",
			ID:   "127.0.0.1:7070",
			Port: 7070,
		},
		{
			Host: "127.0.0.1",
			ID:   "127.0.0.1:8080",
			Port: 8080,
		},
		{
			Host: "127.0.0.1",
			ID:   "127.0.0.1:9090",
			Port: 9090,
		},
	}
)

func prepareCheckerPeers(nodes []healthcheck.CheckerPeer, oldPeers map[string]*HeartBeatHealthChecker) map[string]*HeartBeatHealthChecker {
	log.SetOutputLevel(commonlog.DebugLevel)
	checkers := make(map[string]*HeartBeatHealthChecker, len(nodes))
	for i := range nodes {
		if val, ok := oldPeers[nodes[i].ID]; ok {
			val.refreshPeers(nodes)
			for j := range val.peers {
				peer := val.peers[j]
				_ = peer
			}
			checkers[nodes[i].ID] = val
			delete(oldPeers, nodes[i].ID)
			continue
		}
		checker := &HeartBeatHealthChecker{}
		checker.Initialize(&apis.ConfigEntry{
			Option: map[string]interface{}{
				"listenPort": int64(nodes[i].Port),
			},
		})

		checker.refreshPeers(nodes)
		for j := range checker.peers {
			peer := checker.peers[j]
			_ = peer
		}
		checkers[nodes[i].ID] = checker
	}

	for i := range oldPeers {
		oldPeers[i].Destroy()
	}

	wait := sync.WaitGroup{}
	wait.Add(len(checkers))
	for i := range checkers {
		checker := checkers[i]
		go func(checker *HeartBeatHealthChecker) {
			defer wait.Done()
			checker.calculateContinuum()
		}(checker)
	}

	wait.Wait()

	return checkers
}

func TestHeartBeatHealthChecker(t *testing.T) {
	checkers := prepareCheckerPeers(nodes, map[string]*HeartBeatHealthChecker{})
	t.Cleanup(func() {
		for i := range checkers {
			_ = checkers[i].Destroy()
		}
	})

	checker1 := checkers[nodes[0].ID]
	checker2 := checkers[nodes[1].ID]
	checker3 := checkers[nodes[2].ID]

	mockKey := utils.NewUUID()
	mockHost := "172.0.0.1"
	mockTimeSec := time.Now().Unix()

	request := &healthcheck.ReportRequest{
		QueryRequest: healthcheck.QueryRequest{
			InstanceId: mockKey,
			Host:       mockHost,
		},
		LocalHost:  nodes[0].Host,
		CurTimeSec: mockTimeSec,
	}

	t.Run("checker_peers 写数据结果对比", func(t *testing.T) {
		// 随机选一个节点进行 report
		err := checkers[nodes[rand.Intn(len(nodes))].ID].Report(context.Background(), request)
		assert.NoError(t, err)
		// wait put op success
		time.Sleep(time.Second)

		// 判断每个 checker peer 对于 mockKey 计算的 responsible peer 是否一致
		repPeer1, ok := checker1.findResponsiblePeer(mockKey)
		assert.True(t, ok)
		repPeer2, ok := checker2.findResponsiblePeer(mockKey)
		assert.True(t, ok)
		repPeer3, ok := checker3.findResponsiblePeer(mockKey)
		assert.True(t, ok)
		assert.True(t, repPeer1.Host() == repPeer2.Host() && repPeer1.Host() == repPeer3.Host())
	})

	t.Run("checker_peers 读结果对比", func(t *testing.T) {
		// 从每个 checker_peer 中查询 mockKey 对应的 value，判断是否和 mockTimeSec 一致
		resp, err := checker1.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.True(t, resp.Exists)
		assert.Equal(t, mockTimeSec, resp.LastHeartbeatSec)

		resp, err = checker2.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.True(t, resp.Exists)
		assert.Equal(t, mockTimeSec, resp.LastHeartbeatSec)

		resp, err = checker3.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.True(t, resp.Exists)
		assert.Equal(t, mockTimeSec, resp.LastHeartbeatSec)
	})

	t.Run("checker_peers 删数据对比", func(t *testing.T) {
		// 随机选一个节点进行删除目标数据
		err := checkers[nodes[rand.Intn(len(nodes))].ID].Delete(context.Background(), mockKey)
		assert.NoError(t, err)
		time.Sleep(time.Second)

		resp, err := checker1.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.False(t, resp.Exists)

		resp, err = checker2.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.False(t, resp.Exists)

		resp, err = checker3.Query(context.Background(), &request.QueryRequest)
		assert.NoError(t, err)
		assert.False(t, resp.Exists)
	})

	t.Run("checker_peers 扩容", func(t *testing.T) {
		mockNewNodeID := "127.0.0.1:6060"

		newNodes := []healthcheck.CheckerPeer{}
		newNodes = append(newNodes, nodes...)
		newNodes = append(newNodes, healthcheck.CheckerPeer{
			ID:   mockNewNodeID,
			Host: "127.0.0.1",
			Port: 6060,
		})

		copyCheckers := map[string]*HeartBeatHealthChecker{}
		for i := range checkers {
			copyCheckers[i] = checkers[i]
		}
		newCheckers := prepareCheckerPeers(newNodes, copyCheckers)
		assert.Equal(t, len(newNodes), len(newCheckers))

		mockKeys := map[string]string{}
		newMockVal := time.Now().Unix()

		t.Run("扩容后-数据写入测试", func(t *testing.T) {
			// 每个 checker 都写一次数据
			for i := range newCheckers {
				checker := newCheckers[i]
				mockKeys[i] = fmt.Sprintf("add_peer_%s", i)
				request := &healthcheck.ReportRequest{
					QueryRequest: healthcheck.QueryRequest{
						InstanceId: mockKeys[i],
						Host:       i,
					},
					LocalHost:  i,
					CurTimeSec: newMockVal,
				}
				err := checker.Report(context.Background(), request)
				assert.NoError(t, err)
			}

			// 每个 checkers 都去查询 mockKeys 的全部信息
			respDirect := [][]*healthcheck.QueryResponse{}
			for i := range newCheckers {
				checker := newCheckers[i]
				tmpResps := []*healthcheck.QueryResponse{}
				for j := range mockKeys {
					key := mockKeys[j]
					resp, err := checker.Query(context.Background(), &healthcheck.QueryRequest{
						InstanceId: key,
						Host:       i,
					})
					assert.NoError(t, err)
					tmpResps = append(tmpResps, resp)
				}
				respDirect = append(respDirect, tmpResps)
			}

			assert.Equal(t, len(newCheckers), len(respDirect))
			expect := respDirect[0]
			for i := 1; i < len(respDirect); i++ {
				assert.ElementsMatch(t, expect, respDirect[i])
			}
		})

		t.Run("缩容后-数据查询测试", func(t *testing.T) {
			newNodes2 := newNodes[:len(nodes)-1]
			copyCheckers := map[string]*HeartBeatHealthChecker{}
			for i := range newCheckers {
				copyCheckers[i] = newCheckers[i]
			}
			newCheckers2 := prepareCheckerPeers(newNodes2, copyCheckers)
			assert.Equal(t, len(newNodes2), len(newCheckers2))

			respDirect := [][]*healthcheck.QueryResponse{}
			for i := range newCheckers2 {
				checker := newCheckers2[i]
				tmpResps := []*healthcheck.QueryResponse{}
				for j := range mockKeys {
					key := mockKeys[j]
					resp, err := checker.Query(context.Background(), &healthcheck.QueryRequest{
						InstanceId: key,
						Host:       i,
					})
					assert.NoError(t, err)
					tmpResps = append(tmpResps, resp)
				}
				respDirect = append(respDirect, tmpResps)
			}

			assert.Equal(t, len(newCheckers2), len(respDirect))
			expect := respDirect[0]
			for i := 1; i < len(respDirect); i++ {
				assert.ElementsMatch(t, expect, respDirect[i])
			}
		})
	})
}
