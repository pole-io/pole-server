# Skill 存储层查询方法实现设计

## 1. 概述

本文档描述 `QuerySkillVersions` 和 `QuerySkillSubscriptions` 方法的实现设计。

## 2. 参考实现

参考 `plugin/store/mysql/mcp_server.go` 中 `QueryMCPServers` 的实现模式：

```go
func (m *mcpServerStore) QueryMCPServers(filter map[string]string, offset, limit uint32) (uint32, []*ai.MCPServer, error) {
    // 构建查询条件
    whereClause := "WHERE flag != 1"
    args := make([]interface{}, 0)

    if name, ok := filter["name"]; ok && name != "" {
        whereClause += " AND name = ?"
        args = append(args, name)
    }
    // ... 其他过滤条件

    // 查询总数
    countSql := "SELECT COUNT(*) FROM mcp_server " + whereClause
    row := m.slave.QueryRow(countSql, args...)
    var count int
    if err := row.Scan(&count); err != nil {
        return 0, nil, err
    }

    // 查询列表
    querySql := fmt.Sprintf(`SELECT ... FROM mcp_server %s ORDER BY mtime DESC LIMIT ?, ?`, whereClause)
    queryArgs := append(args, offset, limit)
    rows, err := m.slave.Query(querySql, queryArgs...)
    // ...

    return uint32(count), servers, nil
}
```

## 3. QuerySkillVersions 实现

### 3.1 方法签名

```go
func (s *skillVersionStore) QuerySkillVersions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillVersion, error)
```

### 3.2 实现代码

```go
func (s *skillVersionStore) QuerySkillVersions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillVersion, error) {
    // 构建查询条件
    whereClause := "WHERE flag != 1"
    args := make([]interface{}, 0)

    // 按 skill_name 过滤
    if skillName, ok := filter["skill_name"]; ok && skillName != "" {
        whereClause += " AND skill_name = ?"
        args = append(args, skillName)
    }

    // 按 namespace 过滤
    if namespace, ok := filter["namespace"]; ok && namespace != "" {
        whereClause += " AND namespace = ?"
        args = append(args, namespace)
    }

    // 按 skill_id 过滤
    if skillID, ok := filter["skill_id"]; ok && skillID != "" {
        whereClause += " AND skill_id = ?"
        args = append(args, skillID)
    }

    // 按 active 状态过滤
    if active, ok := filter["active"]; ok && active != "" {
        whereClause += " AND active = ?"
        args = append(args, active == "true" || active == "1")
    }

    // 查询总数
    countSql := "SELECT COUNT(*) FROM skill_version " + whereClause
    row := s.slave.QueryRow(countSql, args...)
    var count int
    if err := row.Scan(&count); err != nil {
        log.Errorf("[Store][database] query skill versions count err: %s", err.Error())
        return 0, nil, err
    }

    // 查询列表
    querySql := fmt.Sprintf(`SELECT id, skill_id, skill_name, namespace, version, comment,
        input_schema, output_schema, skill_type, metadata, active, flag,
        unix_timestamp(ctime), unix_timestamp(mtime)
        FROM skill_version %s ORDER BY mtime DESC LIMIT ?, ?`, whereClause)

    queryArgs := append(args, offset, limit)
    rows, err := s.slave.Query(querySql, queryArgs...)
    if err != nil {
        log.Errorf("[Store][database] query skill versions err: %s", err.Error())
        return 0, nil, err
    }
    defer func() { _ = rows.Close() }()

    var versions []*ai.SkillVersion
    for rows.Next() {
        version, err := fetchSkillVersionRow(rows)
        if err != nil {
            return 0, nil, err
        }
        if version != nil {
            versions = append(versions, version)
        }
    }

    if err := rows.Err(); err != nil {
        log.Errorf("[Store][database] query skill versions rows err: %s", err.Error())
        return 0, nil, err
    }

    return uint32(count), versions, nil
}
```

### 3.3 支持的过滤参数

| 参数 | 说明 | 示例 |
|------|------|------|
| skill_name | Skill 名称 | "my-skill" |
| namespace | 命名空间 | "default" |
| skill_id | Skill ID | "uuid-xxx" |
| active | 是否激活 | "true", "false" |

## 4. QuerySkillSubscriptions 实现

### 4.1 方法签名

```go
func (s *skillSubscriptionStore) QuerySkillSubscriptions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillSubscription, error)
```

### 4.2 实现代码

```go
func (s *skillSubscriptionStore) QuerySkillSubscriptions(filter map[string]string, offset, limit uint32) (uint32, []*ai.SkillSubscription, error) {
    // 构建查询条件
    whereClause := "WHERE flag != 1"
    args := make([]interface{}, 0)

    // 按 client_id 过滤
    if clientID, ok := filter["client_id"]; ok && clientID != "" {
        whereClause += " AND client_id = ?"
        args = append(args, clientID)
    }

    // 按 skill_name 过滤
    if skillName, ok := filter["skill_name"]; ok && skillName != "" {
        whereClause += " AND skill_name = ?"
        args = append(args, skillName)
    }

    // 按 namespace 过滤
    if namespace, ok := filter["namespace"]; ok && namespace != "" {
        whereClause += " AND namespace = ?"
        args = append(args, namespace)
    }

    // 按 active 状态过滤
    if active, ok := filter["active"]; ok && active != "" {
        whereClause += " AND active = ?"
        args = append(args, active == "true" || active == "1")
    }

    // 查询总数
    countSql := "SELECT COUNT(*) FROM skill_subscription " + whereClause
    row := s.slave.QueryRow(countSql, args...)
    var count int
    if err := row.Scan(&count); err != nil {
        log.Errorf("[Store][database] query skill subscriptions count err: %s", err.Error())
        return 0, nil, err
    }

    // 查询列表
    querySql := fmt.Sprintf(`SELECT id, client_id, skill_id, skill_name, namespace, version,
        active, flag, unix_timestamp(ctime), unix_timestamp(mtime)
        FROM skill_subscription %s ORDER BY mtime DESC LIMIT ?, ?`, whereClause)

    queryArgs := append(args, offset, limit)
    rows, err := s.slave.Query(querySql, queryArgs...)
    if err != nil {
        log.Errorf("[Store][database] query skill subscriptions err: %s", err.Error())
        return 0, nil, err
    }
    defer func() { _ = rows.Close() }()

    var subs []*ai.SkillSubscription
    for rows.Next() {
        sub, err := fetchSkillSubscriptionRow(rows)
        if err != nil {
            return 0, nil, err
        }
        if sub != nil {
            subs = append(subs, sub)
        }
    }

    if err := rows.Err(); err != nil {
        log.Errorf("[Store][database] query skill subscriptions rows err: %s", err.Error())
        return 0, nil, err
    }

    return uint32(count), subs, nil
}
```

### 4.3 辅助函数 fetchSkillSubscriptionRow

```go
func fetchSkillSubscriptionRow(rows *sql.Rows) (*ai.SkillSubscription, error) {
    var sub ai.SkillSubscription
    var ctime, mtime int64

    err := rows.Scan(
        &sub.ID,
        &sub.ClientID,
        &sub.SkillID,
        &sub.SkillName,
        &sub.Namespace,
        &sub.Version,
        &sub.Active,
        &sub.Flag,
        &ctime,
        &mtime,
    )
    if err != nil {
        return nil, err
    }

    sub.CTime = time.Unix(ctime, 0)
    sub.MTime = time.Unix(mtime, 0)
    return &sub, nil
}
```

### 4.4 支持的过滤参数

| 参数 | 说明 | 示例 |
|------|------|------|
| client_id | 客户端 ID | "client-xxx" |
| skill_name | Skill 名称 | "my-skill" |
| namespace | 命名空间 | "default" |
| active | 是否激活 | "true", "false" |

## 5. 实现位置

| 文件 | 行号 | 方法 |
|------|------|------|
| `plugin/store/mysql/skill.go` | 791-793 | QuerySkillVersions |
| `plugin/store/mysql/skill.go` | 1080-1082 | QuerySkillSubscriptions |

## 6. 验收标准

1. 能按过滤条件查询 Skill 版本列表
2. 能按过滤条件查询 Skill 订阅列表
3. 分页参数正确传递
4. 返回总数正确
5. 错误处理完善
