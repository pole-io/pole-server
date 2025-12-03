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

package rds

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pole-io/pole-server/apis"
	"github.com/pole-io/pole-server/apis/observability/event"
	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	commonlog "github.com/pole-io/pole-server/pkg/common/log"
	"github.com/pole-io/pole-server/pkg/common/utils"
)

const (
	PluginName        = "EventDB"
	defaultBufferSize = 1024
)

var log = commonlog.RegisterScope(PluginName, "", 0)

func init() {
	d := &discoverEventDB{}
	apis.RegisterPlugin(d.Name(), d)
}

type eventBufferHolder struct {
	writeCursor int
	readCursor  int
	size        int
	buffer      []event.DiscoverEvent
}

func newEventBufferHolder(cap int) *eventBufferHolder {
	return &eventBufferHolder{
		writeCursor: 0,
		readCursor:  0,
		size:        0,
		buffer:      make([]event.DiscoverEvent, cap),
	}
}

// Reset 重置 eventBufferHolder，使之可以复用
func (holder *eventBufferHolder) Reset() {
	holder.writeCursor = 0
	holder.readCursor = 0
	holder.size = 0
}

// Put 放入一个 model.DiscoverEvent
func (holder *eventBufferHolder) Put(event event.DiscoverEvent) {
	holder.buffer[holder.writeCursor] = event
	holder.size++
	holder.writeCursor++
}

// HasNext 判断是否还有下一个元素
func (holder *eventBufferHolder) HasNext() bool {
	return holder.readCursor < holder.size
}

// Next 返回下一个元素
//
//	@return model.DiscoverEvent 元素
//	@return bool 是否还有下一个元素可以继续读取
func (holder *eventBufferHolder) Next() event.DiscoverEvent {
	event := holder.buffer[holder.readCursor]
	holder.readCursor++

	return event
}

// Size 当前所存储的有效元素的个数
func (holder *eventBufferHolder) Size() int {
	return holder.size
}

type discoverEventDB struct {
	db             *sql.DB
	eventCh        chan event.DiscoverEvent
	bufferPool     sync.Pool
	curEventBuffer *eventBufferHolder
	syncLock       sync.Mutex
	eventHandler   func(eventHolder *eventBufferHolder)
	cancel         context.CancelFunc
}

// Name 插件名称
// @return string 返回插件名称
func (el *discoverEventDB) Name() string {
	return PluginName
}

// Initialize 根据配置文件进行初始化插件 discoverEventDB
// @param conf 配置文件内容
// @return error 初始化失败，返回 error 信息
func (el *discoverEventDB) Initialize(conf *apis.ConfigEntry) error {
	dns := conf.Option["dns"].(string)
	if dns == "" {
		return errors.New("dns is required for discoverEventDB plugin")
	}

	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Errorf("failed to connect to database: %v", err)
		return err
	}

	el.db = db
	el.eventCh = make(chan event.DiscoverEvent, 1024)
	el.eventHandler = el.writeToDB
	el.bufferPool = sync.Pool{
		New: func() interface{} {
			return newEventBufferHolder(defaultBufferSize)
		},
	}

	el.switchEventBuffer()
	const tableName = "discover_event"
	// 初始化时建表
	if err := el.ensureEventTable(tableName); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go el.Run(ctx)
	// 启动定时分区检查任务
	go el.partitionChecker(ctx, tableName)
	el.cancel = cancel
	return nil
}

func (el *discoverEventDB) Type() apis.PluginType {
	return apis.PluginTypeDiscoverEvent
}

// Destroy 执行插件销毁
func (el *discoverEventDB) Destroy() error {
	if el.cancel != nil {
		el.cancel()
	}
	return nil
}

// PublishEvent 发布一个服务事件
func (el *discoverEventDB) PublishEvent(event event.DiscoverEvent) {
	select {
	case el.eventCh <- event:
		return
	default:
		// do nothing
	}
}

var (
	subscribeEvents = map[string]struct{}{
		string(svctypes.EventInstanceCloseIsolate): {},
		string(svctypes.EventInstanceOpenIsolate):  {},
		string(svctypes.EventInstanceOffline):      {},
		string(svctypes.EventInstanceOnline):       {},
		string(svctypes.EventInstanceTurnHealth):   {},
		string(svctypes.EventInstanceTurnUnHealth): {},
	}
)

// Run 执行主逻辑
func (el *discoverEventDB) Run(ctx context.Context) {
	// 定时刷新事件到日志的定时器
	syncInterval := time.NewTicker(time.Duration(10) * time.Second)
	defer syncInterval.Stop()

	for {
		select {
		case event := <-el.eventCh:
			if _, ok := subscribeEvents[event.Event()]; !ok {
				break
			}

			el.curEventBuffer.Put(event)

			// 触发持久化到 log 阈值
			if el.curEventBuffer.Size() == defaultBufferSize {
				go el.eventHandler(el.curEventBuffer)
				el.switchEventBuffer()
			}
		case <-syncInterval.C:
			go el.eventHandler(el.curEventBuffer)
			el.switchEventBuffer()
		case <-ctx.Done():
			return
		}
	}
}

// switchEventBuffer 换一个新的 buffer 实例继续使用
func (el *discoverEventDB) switchEventBuffer() {
	el.curEventBuffer = el.bufferPool.Get().(*eventBufferHolder)
}

// partitionChecker 定时检查并自动添加分区
func (el *discoverEventDB) partitionChecker(ctx context.Context, tableName string) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 检查未来 12 个月分区
			now := time.Now()
			for i := 0; i < 12; i++ {
				t := now.AddDate(0, i, 0)
				year, month := t.Year(), int(t.Month())
				_ = el.ensurePartition(tableName, year, month)
			}
		case <-ctx.Done():
			return
		}
	}
}

// writeToDB 事件落盘
func (el *discoverEventDB) writeToDB(eventHolder *eventBufferHolder) {
	el.syncLock.Lock()
	defer func() {
		el.syncLock.Unlock()
		eventHolder.Reset()
		el.bufferPool.Put(eventHolder)
	}()

	if el.db == nil {
		log.Error("db is nil, skip writeToDB")
		return
	}

	const tableName = "discover_event"
	// 1. 检查表是否存在，不再需要，已在初始化阶段完成

	tmp := make([]event.DiscoverEvent, 0, eventHolder.Size())
	for eventHolder.HasNext() {
		tmp = append(tmp, eventHolder.Next())
	}
	// 4. 批量插入事件
	if err := el.batchInsertEvents(tableName, tmp); err != nil {
		log.Errorf("batchInsertEvents failed: %v", err)
	}
}

// ensureEventTable 检查并自动建表
func (el *discoverEventDB) ensureEventTable(tableName string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
		id BIGINT NOT NULL AUTO_INCREMENT,
		namespace VARCHAR(128) NOT NULL,
		service VARCHAR(128) NOT NULL,
		resource VARCHAR(64) NOT NULL,
		etype VARCHAR(64) NOT NULL,
		happen_time DATETIME NOT NULL,
		server VARCHAR(64) NOT NULL,
		PRIMARY KEY (id, happen_time),
		KEY idx_happen_time (happen_time),
		KEY idx_namespace_service (namespace, service),
		KEY idx_resource (resource)
	) PARTITION BY RANGE COLUMNS(happen_time) (
		PARTITION p202506 VALUES LESS THAN ('2025-07-01'),
		PARTITION p202507 VALUES LESS THAN ('2025-08-01')
	);`
	_, err := el.db.Exec(createTableSQL)
	return err
}

// ensurePartition 检查并自动添加分区
func (el *discoverEventDB) ensurePartition(tableName string, year, month int) error {
	partitionName := fmt.Sprintf("p%04d%02d", year, month)
	partitionValue := fmt.Sprintf("'%04d-%02d-01'", year, month+1)
	if month == 12 {
		partitionValue = fmt.Sprintf("'%04d-01-01'", year+1)
	}
	checkSQL := `SELECT COUNT(1) FROM information_schema.PARTITIONS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND PARTITION_NAME=?`
	var cnt int
	err := el.db.QueryRow(checkSQL, tableName, partitionName).Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt == 0 {
		alterSQL := fmt.Sprintf(`ALTER TABLE %s ADD PARTITION (PARTITION %s VALUES LESS THAN (%s))`, tableName, partitionName, partitionValue)
		_, err := el.db.Exec(alterSQL)
		if err != nil {
			return err
		}
	}
	return nil
}

// batchInsertEvents 批量插入事件
func (el *discoverEventDB) batchInsertEvents(tableName string, events []event.DiscoverEvent) error {
	if len(events) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]interface{}, 0, len(events)*7)
	for _, e := range events {
		if v, ok := e.(*svctypes.InstanceEvent); ok {
			res := v.Id
			if v.Instance != nil {
				res = fmt.Sprintf("%s:%d", v.Instance.GetHost(), v.Instance.GetPort())
			}
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, v.Namespace, v.Service, res, v.EType, v.CreateTime, utils.LocalHost)
		} else if v, ok := e.(*svctypes.ServiceEvent); ok {
			valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
			valueArgs = append(valueArgs, v.Namespace, v.Service, v.Id, v.EType, v.CreateTime, utils.LocalHost)
		}
	}

	insertSQL := fmt.Sprintf(`INSERT IGNORE INTO %s (namespace, service, resource, etype, happen_time, server) VALUES %s`,
		tableName,
		joinStrings(valueStrings, ", "))

	tx, err := el.db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(insertSQL, valueArgs...); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func joinStrings(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	res := arr[0]
	for i := 1; i < len(arr); i++ {
		res += sep + arr[i]
	}
	return res
}
