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

package sqldb

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
)

// RuleOperation 定义规则操作的通用接口
type RuleOperation interface {
	// GetOperationLabel 获取操作标签（用于日志和重试）
	GetOperationLabel() string
	// Execute 执行具体的 SQL 操作
	Execute(tx *BaseTx) error
}

// CreateOperation 创建操作
type CreateOperation struct {
	Label string
	SQL   string
	Args  []interface{}
}

func (co *CreateOperation) GetOperationLabel() string {
	return co.Label
}

func (co *CreateOperation) Execute(tx *BaseTx) error {
	_, err := tx.Exec(co.SQL, co.Args...)
	return err
}

// UpdateOperation 更新操作
type UpdateOperation struct {
	Label string
	SQL   string
	Args  []interface{}
}

func (uo *UpdateOperation) GetOperationLabel() string {
	return uo.Label
}

func (uo *UpdateOperation) Execute(tx *BaseTx) error {
	_, err := tx.Exec(uo.SQL, uo.Args...)
	return err
}

// DeleteOperation 删除操作
type DeleteOperation struct {
	Label string
	SQL   string
	Args  []interface{}
}

func (do *DeleteOperation) GetOperationLabel() string {
	return do.Label
}

func (do *DeleteOperation) Execute(tx *BaseTx) error {
	_, err := tx.Exec(do.SQL, do.Args...)
	return err
}

// FnOperation 自定义函数操作（用于多语句或复杂事务逻辑）
type FnOperation struct {
	Label string
	Fn    func(tx *BaseTx) error
}

func (fo *FnOperation) GetOperationLabel() string {
	return fo.Label
}

func (fo *FnOperation) Execute(tx *BaseTx) error {
	return fo.Fn(tx)
}

// BaseRuleStore 治理规则存储的基础实现
type BaseRuleStore struct {
	master *BaseDB
	slave  *BaseDB
}

// NewBaseRuleStore 创建基础规则存储
func NewBaseRuleStore(master, slave *BaseDB) *BaseRuleStore {
	return &BaseRuleStore{
		master: master,
		slave:  slave,
	}
}

// ExecuteRuleOperation 执行规则操作（带重试，自动管理事务）
func (b *BaseRuleStore) ExecuteRuleOperation(op RuleOperation) error {
	return b.ExecuteRuleOperationTx(nil, op)
}

// ExecuteRuleOperationTx 执行规则操作（统一事务支持）
// - tx 为 nil 时：自动创建事务、执行、提交，带重试
// - tx 非 nil 时：在已有事务中执行，不提交（由调用方管理事务生命周期）
func (b *BaseRuleStore) ExecuteRuleOperationTx(tx store.Tx, op RuleOperation) error {
	if tx == nil {
		return RetryTransaction(op.GetOperationLabel(), func() error {
			return b.master.processWithTransaction(op.GetOperationLabel(), func(btx *BaseTx) error {
				if err := op.Execute(btx); err != nil {
					return err
				}
				return btx.Commit()
			})
		})
	}
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	return op.Execute(dbTx)
}

// MustGetBaseTx 从 store.Tx 获取 *BaseTx，供需要外部事务的规则方法使用
// 消除重复的 "if tx == nil; dbTx := tx.GetDelegateTx().(*BaseTx)" 样板代码
func MustGetBaseTx(tx store.Tx) (*BaseTx, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx, ok := tx.GetDelegateTx().(*BaseTx)
	if !ok || dbTx == nil {
		return nil, errors.New("invalid tx delegate")
	}
	return dbTx, nil
}

// CheckRuleExists 检查规则是否存在
func (b *BaseRuleStore) CheckRuleExists(query string, args ...interface{}) (bool, error) {
	row := b.master.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

// GetRuleCount 获取规则数量
func (b *BaseRuleStore) GetRuleCount(query string, args ...interface{}) (int, error) {
	row := b.master.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// EtimeBuilder etime 字符串构建器
type EtimeBuilder struct {
	isEnabled bool
}

// NewEtimeBuilder 创建 EtimeBuilder
func NewEtimeBuilder(isEnabled bool) *EtimeBuilder {
	return &EtimeBuilder{isEnabled: isEnabled}
}

// Build 构建 etime 字符串
func (eb *EtimeBuilder) Build() string {
	if eb.isEnabled {
		return "sysdate()"
	}
	return emptyEnableTime
}

// RuleExistsChecker 规则存在检查器
type RuleExistsChecker struct {
	store *BaseRuleStore
}

// NewRuleExistsChecker 创建规则存在检查器
func NewRuleExistsChecker(store *BaseRuleStore) *RuleExistsChecker {
	return &RuleExistsChecker{store: store}
}

// ByID 根据 ID 检查规则是否存在
func (rec *RuleExistsChecker) ByID(query string, id string) (bool, error) {
	return rec.store.CheckRuleExists(query, id)
}

// ByName 根据名称检查规则是否存在
func (rec *RuleExistsChecker) ByName(query string, name, namespace string) (bool, error) {
	return rec.store.CheckRuleExists(query, name, namespace)
}

// ByNameExcludeID 根据名称检查规则是否存在（排除指定 ID）
func (rec *RuleExistsChecker) ByNameExcludeID(query string, name, namespace, id string) (bool, error) {
	return rec.store.CheckRuleExists(query, name, namespace, id)
}

// OperationBuilder CRUD 操作构建器
type OperationBuilder struct {
	label  string
	sql    string
	args   []interface{}
	opType string
}

// NewCreateBuilder 创建 Create 操作构建器
func NewCreateBuilder(label, sql string) *OperationBuilder {
	return &OperationBuilder{
		label:  label,
		sql:    sql,
		opType: "create",
	}
}

// NewUpdateBuilder 创建 Update 操作构建器
func NewUpdateBuilder(label, sql string) *OperationBuilder {
	return &OperationBuilder{
		label:  label,
		sql:    sql,
		opType: "update",
	}
}

// NewDeleteBuilder 创建 Delete 操作构建器
func NewDeleteBuilder(label, sql string) *OperationBuilder {
	return &OperationBuilder{
		label:  label,
		sql:    sql,
		opType: "delete",
	}
}

// Args 设置参数
func (ob *OperationBuilder) Args(args ...interface{}) *OperationBuilder {
	ob.args = args
	return ob
}

// Build 构建操作
func (ob *OperationBuilder) Build() RuleOperation {
	switch ob.opType {
	case "create":
		return &CreateOperation{Label: ob.label, SQL: ob.sql, Args: ob.args}
	case "update":
		return &UpdateOperation{Label: ob.label, SQL: ob.sql, Args: ob.args}
	case "delete":
		return &DeleteOperation{Label: ob.label, SQL: ob.sql, Args: ob.args}
	default:
		return nil
	}
}

// BriefSearchKey 列表查询时表示只查简要信息的 filter key
const BriefSearchKey = "brief"

// ParseBriefFilter 从 filter 中解析 brief 参数并移除该 key，避免透传到 SQL。
// 返回 isBrief（是否为简要查询）和过滤后的 filter（会复制一份再删除 key，不修改原 map）。
func ParseBriefFilter(filter map[string]string, briefKey string) (isBrief bool, out map[string]string) {
	out = make(map[string]string, len(filter))
	for k, v := range filter {
		out[k] = v
	}
	v, ok := out[briefKey]
	if ok {
		delete(out, briefKey)
		isBrief = strings.ToLower(v) == "true"
	}
	return isBrief, out
}

// GetMoreRulesByMtime 按 mtime 增量拉取规则的通用模板：拼 cacheSql + firstUpdate 条件，从 slave 查询后由 fetchRows 解析。
// fetchRows 内部需自行 defer rows.Close()。
func GetMoreRulesByMtime(
	slave *BaseDB,
	cacheSql string,
	firstUpdate bool,
	mtime time.Time,
	fetchRows func(*sql.Rows) (interface{}, error),
) (interface{}, error) {
	if firstUpdate {
		cacheSql += " and flag != 1"
	}
	rows, err := slave.Query(cacheSql, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	return fetchRows(rows)
}

// GetMoreReleasesByMtime 按 mtime 增量拉取规则发布（release）的通用模板。
// baseSql 需包含 WHERE mtime > FROM_UNIXTIME(?)；firstUpdate 为 true 时在 baseSql 后追加 firstUpdateSuffix（如 " AND active = 1"）。
// fetchRows 内部需自行 defer rows.Close()。
func GetMoreReleasesByMtime(
	slave *BaseDB,
	baseSql string,
	firstUpdate bool,
	firstUpdateSuffix string,
	mtime time.Time,
	fetchRows func(*sql.Rows) (interface{}, error),
) (interface{}, error) {
	if firstUpdate && firstUpdateSuffix != "" {
		baseSql += firstUpdateSuffix
	}
	rows, err := slave.Query(baseSql, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	return fetchRows(rows)
}

// ExecuteActiveRule 执行“激活某条发布规则”的通用流程：取 tx、inactive 得到 maxVersion、再 Exec updateSql。
// updateSql 占位符顺序应为：version, name, rule_name（或各 store 自定义），args 由 getUpdateArgs(release) 提供，不含 version。
func ExecuteActiveRule(
	tx store.Tx,
	release interface{},
	getInactiveMaxVersion func(*BaseTx, interface{}) (uint64, error),
	updateSql string,
	getUpdateArgs func(interface{}) []interface{},
) error {
	dbTx, err := MustGetBaseTx(tx)
	if err != nil {
		return err
	}
	maxVersion, err := getInactiveMaxVersion(dbTx, release)
	if err != nil {
		return err
	}
	args := append([]interface{}{maxVersion + 1}, getUpdateArgs(release)...)
	_, err = dbTx.Exec(updateSql, args...)
	return err
}

// BuildActiveRuleWhere 根据 rule_name / name / release_type 构建 GetActive* 的 WHERE 条件与参数（不含前缀 "WHERE"）。
func BuildActiveRuleWhere(ruleName, releaseName, releaseType string) (where string, args []interface{}) {
	parts := []string{"1=1"}
	if ruleName != "" {
		parts = append(parts, "rule_name = ?")
		args = append(args, ruleName)
	}
	if releaseName != "" {
		parts = append(parts, "name = ?")
		args = append(args, releaseName)
	}
	if releaseType != "" {
		parts = append(parts, "release_type = ?")
		args = append(args, releaseType)
	}
	return strings.Join(parts, " AND "), args
}

// ReleaseVersionColumns 发布表版本列表查询的公共列（与统一表结构一致）
const ReleaseVersionColumns = `id, name, rule_id, rule_name, flag, active, version, description, release_type, unix_timestamp(ctime), unix_timestamp(mtime)`

// ExecInactiveRelease 将发布表中符合 where 条件的记录置为 active=0（与统一表结构一致）。
// whereClause 不含 "WHERE" 前缀，args 为对应参数。
func ExecInactiveRelease(tx *BaseTx, releaseTable, whereClause string, args ...interface{}) error {
	if tx == nil {
		return ErrTxIsNil
	}
	sql := "UPDATE " + releaseTable + " SET active = 0, mtime = sysdate() WHERE " + whereClause
	_, err := tx.Exec(sql, args...)
	return err
}

// QueryRuleVersions 治理规则发布版本列表的通用查询：count + 分页查询 + 统一列 Scan 成 RuleRelease。
// 表结构统一为 (id, name, rule_id, rule_name, flag, active, version, description, release_type, ctime, mtime)。
// countWhere/queryWhere 为 WHERE 子句内容（不含 "WHERE"），countArgs 与 queryArgs 为对应参数；queryArgs 会追加 offset, limit。
// setResource 由调用方设置每条 RuleRelease 的 Resource 类型，避免 base 依赖 apimodel。
func QueryRuleVersions(
	slave *BaseDB,
	releaseTable, countWhere, queryWhere string,
	countArgs []interface{},
	queryArgs []interface{},
	offset, limit uint32,
	setResource func(*rules.RuleRelease),
) (uint64, []*rules.RuleRelease, error) {
	countSql := "SELECT COUNT(*) FROM " + releaseTable + " WHERE " + countWhere
	row := slave.QueryRow(countSql, countArgs...)
	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}
	querySql := "SELECT " + ReleaseVersionColumns + " FROM " + releaseTable + " WHERE " + queryWhere + " ORDER BY version DESC LIMIT ?, ?"
	args := append(queryArgs, offset, limit)
	rows, err := slave.Query(querySql, args...)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	defer rows.Close()
	var releases []*rules.RuleRelease
	for rows.Next() {
		item := &rules.RuleRelease{}
		var flag, active int
		var ctime, mtime int64
		if err := rows.Scan(&item.Id, &item.ReleaseName, &item.RuleId, &item.RuleName, &flag, &active, &item.Version, &item.Description, &item.ReleaseType, &ctime, &mtime); err != nil {
			return 0, nil, store.Error(err)
		}
		item.Active = active == 1
		item.Valid = flag == 0
		item.Ctime = time.Unix(ctime, 0)
		item.Mtime = time.Unix(mtime, 0)
		if setResource != nil {
			setResource(item)
		}
		releases = append(releases, item)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, store.Error(err)
	}
	return count, releases, nil
}
