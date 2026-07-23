package mysql

import (
	"strconv"

	"github.com/pole-io/pole-server/console/pkg/common/model"
)

// CREATE TABLE `discover_event` (
//   `id` bigint NOT NULL AUTO_INCREMENT,
//   `namespace` varchar(128) COLLATE utf8mb4_bin NOT NULL,
//   `service` varchar(128) COLLATE utf8mb4_bin NOT NULL,
//   `resource` varchar(64) COLLATE utf8mb4_bin NOT NULL,
//   `etype` varchar(64) COLLATE utf8mb4_bin NOT NULL,
//   `happen_time` datetime NOT NULL,
//   PRIMARY KEY (`id`,`happen_time`),
//   KEY `idx_happen_time` (`happen_time`),
//   KEY `idx_namespace_service` (`namespace`,`service`),
//   KEY `idx_resource` (`resource`)
// ) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
// /*!50500 PARTITION BY RANGE  COLUMNS(happen_time)
// (PARTITION p202506 VALUES LESS THAN ('2025-07-01') ENGINE = InnoDB,
//  PARTITION p202507 VALUES LESS THAN ('2025-08-01') ENGINE = InnoDB) */

type EventFetcher struct {
	master *BaseDB
	slave  *BaseDB
}

// GetEvents 获取事件记录
// filter 参数可以包含以下键：
// - "namespace": 命名空间
// - "service": 服务名称
// - "event_type": 事件类型
// - "resource": 资源名称或实例
// - "start_time": 开始时间
// - "end_time": 结束时间
// - "limit": 限制返回的记录数
// - "cursor": 游标，用于分页
// - "direction": 翻页方向，"prev" 表示向前翻页，"next" 表示向后翻页
// 返回值是一个 EventRecord 的切片，包含事件记录
func (e *EventFetcher) GetEvents(filter map[string]string) ([]*model.EventRecord, error) {
	query := "SELECT id, namespace, service, resource, etype, happen_time, server FROM discover_event WHERE 1=1"
	var args []interface{}

	// Add filter conditions
	if ns, ok := filter["namespace"]; ok && ns != "" {
		query += " AND namespace LIKE ?"
		args = append(args, "%"+ns+"%")
	}
	if svc, ok := filter["service"]; ok && svc != "" {
		query += " AND service LIKE ?"
		args = append(args, "%"+svc+"%")
	}
	if etype, ok := filter["event_type"]; ok && etype != "" {
		query += " AND etype = ?"
		args = append(args, etype)
	}
	if instance, ok := filter["instance"]; ok && instance != "" {
		query += " AND resource LIKE ?"
		args = append(args, "%"+instance+"%")
	}
	if startTime, ok := filter["start_time"]; ok && startTime != "" {
		query += " AND happen_time >= ?"
		args = append(args, startTime)
	}
	if endTime, ok := filter["end_time"]; ok && endTime != "" {
		query += " AND happen_time <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY id DESC"

	// Add limit
	if limitStr, ok := filter["limit"]; ok && limitStr != "" {
		query += " LIMIT ? OFFSET ?"
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err
		}
		args = append(args, limit)
		offset, err := strconv.Atoi(filter["offset"])
		if err != nil {
			return nil, err
		}
		args = append(args, offset)
	} else {
		// Default limit if not specified
		query += " LIMIT 10 OFFSET 0"
	}

	// Execute the query
	rows, err := e.slave.Query(query, args...)
	if err != nil {
		if isMissingTableError(err) {
			return []*model.EventRecord{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var events []*model.EventRecord

	for rows.Next() {
		event := &model.EventRecord{}
		var id int64
		// id, namespace, service, resource, etype, happen_time
		if err := rows.Scan(
			&id,
			&event.Namespace,
			&event.Service,
			&event.Resource,
			&event.EventType,
			&event.EventTime,
			&event.Server,
		); err != nil {
			return nil, err
		}
		event.Cursor = strconv.FormatInt(id, 10) // Use id as cursor
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
