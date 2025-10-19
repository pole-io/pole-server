package sqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
	"github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ store.LosslessRuleStore = (*losslessStore)(nil)

type losslessStore struct {
	master *BaseDB
	slave  *BaseDB
}

// CreateLossLessRule 创建无损规则
func (s *losslessStore) CreateLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] create lossless rule missing id")
	}
	// 这里直接全部字段写入
	err := RetryTransaction("createLosslessRule", func() error {
		return s.master.processWithTransaction("createLosslessRule", func(tx *BaseTx) error {
			ruleJson, err := protojson.Marshal(rule.Proto)
			if err != nil {
				return err
			}
			insertSql := `INSERT INTO lossless_rule(
id, namespace, service, revision, description, config, ctime, mtime)
VALUES(?,?,?,?,?,?,sysdate(),sysdate())`
			// Revision / Description 目前从 Proto 中提取（如果没有对应字段，留空）
			// 尝试通过反射/断言获得字段（保持简单：使用 json 中的可能字段）
			// 简化：不做额外提取，revision/description 留空，由上层填充后再完善
			if _, err := tx.Exec(insertSql, rule.ID, rule.Namespace, rule.Service, rule.Revision, rule.Description, string(ruleJson)); err != nil {
				log.Errorf("[store][mysql][lossless] create lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] create lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// UpdateLossLessRule 更新无损规则
func (s *losslessStore) UpdateLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] update lossless rule missing id")
	}
	err := RetryTransaction("updateLosslessRule", func() error {
		return s.master.processWithTransaction("updateLosslessRule", func(tx *BaseTx) error {
			ruleJson, err := protojson.Marshal(rule.Proto)
			if err != nil {
				return err
			}
			updateSql := `UPDATE lossless_rule SET namespace = ?, service = ?, revision = ?, description = ?, config = ?, mtime = sysdate() WHERE id = ? AND flag = 0`
			if _, err := tx.Exec(updateSql, rule.Namespace, rule.Service, rule.Revision, rule.Description, string(ruleJson), rule.ID); err != nil {
				log.Errorf("[store][mysql][lossless] update lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] update lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// DeleteLossLessRule 删除无损规则
func (s *losslessStore) DeleteLossLessRule(rule *rules.LosslessRule) error {
	if rule == nil || rule.ID == "" {
		return errors.New("[store][mysql][lossless] delete lossless rule missing id")
	}
	err := RetryTransaction("deleteLosslessRule", func() error {
		return s.master.processWithTransaction("deleteLosslessRule", func(tx *BaseTx) error {
			if _, err := tx.Exec(`UPDATE lossless_rule SET flag = 1, mtime = sysdate() WHERE id = ?`, rule.ID); err != nil {
				log.Errorf("[store][mysql][lossless] delete lossless rule(%+v) err: %v", rule, err)
				return err
			}
			if err := tx.Commit(); err != nil {
				log.Errorf("[store][mysql][lossless] delete lossless rule(%+v) commit err: %v", rule, err)
				return err
			}
			return nil
		})
	})
	return store.Error(err)
}

// GetLossLessRuleWithID 根据无损ID拉取无损规则
func (s *losslessStore) GetOneLosslessRule(id string) (*rules.LosslessRule, error) {
	if id == "" {
		return nil, errors.New("id empty")
	}
	querySql := `SELECT id, namespace, service, revision, description, config, flag, unix_timestamp(ctime), unix_timestamp(mtime), IFNULL(metadata,'{}') FROM lossless_rule WHERE id = ?`
	row := s.master.QueryRow(querySql, id)
	var (
		ruleObj     = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
		cfgStr      string
		flag        int
		ctimeTs     int64
		mtimeTs     int64
		metadataStr string
	)
	if err := row.Scan(&ruleObj.ID, &ruleObj.Namespace, &ruleObj.Service, &ruleObj.Revision, &ruleObj.Description, &cfgStr, &flag, &ctimeTs, &mtimeTs, &metadataStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := protojson.Unmarshal([]byte(cfgStr), ruleObj.Proto); err != nil {
		// 如果解析失败，不阻塞，记录日志
		log.Errorf("[store][mysql][lossless] unmarshal config err: %v", err)
	}
	ruleObj.CTime = time.Unix(ctimeTs, 0)
	ruleObj.MTime = time.Unix(mtimeTs, 0)
	ruleObj.Valid = flag == 0
	if metadataStr != "" && metadataStr != "{}" {
		_ = json.Unmarshal([]byte(metadataStr), &ruleObj.Metadata)
	}
	return ruleObj, nil
}

// GetMoreLosslessRules 根据修改时间拉取增量无损规则及最新版本号, 此方法用于 cache 增量更新，需要注意 mtime 应为数据库时间戳
func (s *losslessStore) GetMoreLosslessRules(mtime time.Time, firstUpdate bool) ([]*rules.LosslessRule, error) {
	str := `SELECT id, namespace, service, revision, description, config, flag, unix_timestamp(ctime), unix_timestamp(mtime), IFNULL(metadata,'{}') FROM lossless_rule WHERE mtime > FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " AND flag != 1"
	}
	rows, err := s.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*rules.LosslessRule
	for rows.Next() {
		var lr = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
		var flag int
		var cfgStr string
		var ctimeTs, mtimeTs int64
		var metadataStr string
		if err := rows.Scan(&lr.ID, &lr.Namespace, &lr.Service, &lr.Revision, &lr.Description, &cfgStr, &flag, &ctimeTs, &mtimeTs, &metadataStr); err != nil {
			return nil, err
		}
		_ = protojson.Unmarshal([]byte(cfgStr), lr.Proto)
		lr.CTime = time.Unix(ctimeTs, 0)
		lr.MTime = time.Unix(mtimeTs, 0)
		lr.Valid = flag == 0
		if metadataStr != "" && metadataStr != "{}" {
			_ = json.Unmarshal([]byte(metadataStr), &lr.Metadata)
		}
		out = append(out, lr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// LockLosslessRule 锁住一个无损规则
func (s *losslessStore) LockLosslessRule(tx store.Tx, name string) (*rules.LosslessRule, error) {
	// Schema 中没有 name 字段，这里将 name 视为 id 处理
	if tx == nil {
		return nil, ErrTxIsNil
	}
	if name == "" {
		return nil, ErrorMissingParams
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, namespace, service, revision, description
	, config, flag, unix_timestamp(ctime)
	, unix_timestamp(mtime)
	, IFNULL(metadata, '{}')
FROM lossless_rule
WHERE id = ? OR name = ?
FOR UPDATE`
	row := dbTx.QueryRow(querySql, name, name)
	var (
		ruleObj          = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
		cfgStr           string
		flag             int
		ctimeTs, mtimeTs int64
		metadataStr      string
	)
	if err := row.Scan(&ruleObj.ID, &ruleObj.Namespace, &ruleObj.Service, &ruleObj.Revision, &ruleObj.Description, &cfgStr, &flag, &ctimeTs, &mtimeTs, &metadataStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = protojson.Unmarshal([]byte(cfgStr), ruleObj.Proto)
	ruleObj.CTime = time.Unix(ctimeTs, 0)
	ruleObj.MTime = time.Unix(mtimeTs, 0)
	ruleObj.Valid = flag == 0
	if metadataStr != "" && metadataStr != "{}" {
		_ = json.Unmarshal([]byte(metadataStr), &ruleObj.Metadata)
	}
	return ruleObj, nil
}

// 关于规则发布
func (s *losslessStore) GetLosslessRuleVersions(ctx context.Context, filter map[string]string, offset, limit uint32) (uint64, []*rules.RuleRelease, error) {
	// 版本按 rule_id 统计
	ruleID := filter["rule_id"]
	if ruleID == "" {
		return 0, nil, errors.New("rule_id empty")
	}
	countSql := `SELECT COUNT(*) FROM lossless_rule_release WHERE rule_id = ? AND flag = 0`
	row := s.slave.QueryRow(countSql, ruleID)
	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, nil, store.Error(err)
	}
	if count == 0 {
		return 0, nil, nil
	}
	querySql := `SELECT id, name, rule_id, service, flag, active, version, description, release_type, unix_timestamp(ctime), unix_timestamp(mtime)
FROM lossless_rule_release
WHERE rule_id = ? AND flag = 0 ORDER BY version DESC LIMIT ?, ?`
	rows, err := s.slave.Query(querySql, ruleID, offset, limit)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	defer rows.Close()
	var releases []*rules.RuleRelease
	for rows.Next() {
		var (
			item         = &rules.RuleRelease{}
			flag, active int
			ctime, mtime int64
			svc          string
		)
		if err := rows.Scan(&item.Id, &item.ReleaseName, &item.RuleId, &svc, &flag, &active, &item.Version, &item.Description, &item.ReleaseType, &ctime, &mtime); err != nil {
			return 0, nil, store.Error(err)
		}
		item.RuleName = svc // 暂时用 service 代替 ruleName
		item.Active = active == 1
		item.Valid = flag == 0
		item.Ctime = time.Unix(ctime, 0)
		item.Mtime = time.Unix(mtime, 0)
		item.Resource = apimodel.RuleRelease_LosslessRules
		releases = append(releases, item)
	}
	return count, releases, nil
}

// GetActiveLosslessRule 获取处于使用状态的无损规则
func (s *losslessStore) GetActiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) (*rules.LosslessRuleRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_id, namespace, service, rule, flag, active, version, description, release_type
FROM lossless_rule_release
WHERE %s AND active = 1 AND flag = 0 ORDER BY version DESC LIMIT 1`
	where := []string{"1=1"}
	args := make([]any, 0, 4)
	if release.RuleId != "" {
		where = append(where, "rule_id = ?")
		args = append(args, release.RuleId)
	}
	if release.ReleaseName != "" {
		where = append(where, "name = ?")
		args = append(args, release.ReleaseName)
	}
	if release.ReleaseType != "" {
		where = append(where, "release_type = ?")
		args = append(args, release.ReleaseType)
	}
	querySql = fmt.Sprintf(querySql, strings.Join(where, " AND "))
	row := dbTx.QueryRow(querySql, args...)
	var (
		id, name, ruleID, namespace, service, ruleStr, description, releaseType string
		flag, active                                                            int
		version                                                                 uint64
	)
	if err := row.Scan(&id, &name, &ruleID, &namespace, &service, &ruleStr, &flag, &active, &version, &description, &releaseType); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var ruleObj = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
	if err := protojson.Unmarshal([]byte(ruleStr), ruleObj.Proto); err != nil {
		log.Errorf("[store][mysql][lossless] unmarshal release rule err: %v", err)
	}
	ruleObj.ID = ruleID
	ruleObj.Namespace = namespace
	ruleObj.Service = service
	return &rules.LosslessRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			RuleId:      ruleID,
			RuleName:    service,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       flag == 0,
		}, Rule: ruleObj,
	}, nil
}

// GetReleaseLosslessRule 获取已发布的无损规则
func (s *losslessStore) GetReleaseLosslessRule(tx store.Tx, release *rules.RuleRelease) (*rules.LosslessRuleRelease, error) {
	if tx == nil {
		return nil, ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	querySql := `SELECT id, name, rule_id, namespace, service, rule, flag, active, version, description, release_type
FROM lossless_rule_release
WHERE name = ? AND rule_id = ? AND release_type = ? AND flag = 0 LIMIT 1`
	row := dbTx.QueryRow(querySql, release.ReleaseName, release.RuleId, release.ReleaseType)
	var (
		id, name, ruleID, namespace, service, ruleStr, description, releaseType string
		flag, active                                                            int
		version                                                                 uint64
	)
	if err := row.Scan(&id, &name, &ruleID, &namespace, &service, &ruleStr, &flag, &active, &version, &description, &releaseType); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var ruleObj = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
	if err := protojson.Unmarshal([]byte(ruleStr), ruleObj.Proto); err != nil {
		log.Errorf("[store][mysql][lossless] unmarshal release rule err: %v", err)
	}
	ruleObj.ID = ruleID
	ruleObj.Namespace = namespace
	ruleObj.Service = service
	return &rules.LosslessRuleRelease{
		RuleRelease: rules.RuleRelease{
			Id:          id,
			ReleaseName: name,
			RuleId:      ruleID,
			RuleName:    service,
			Description: description,
			ReleaseType: rules.ReleaseType(releaseType),
			Active:      active == 1,
			Version:     version,
			Valid:       flag == 0,
		},
		Rule: ruleObj,
	}, nil
}

// ActiveLosslessRule 设置某个无损规则发布为使用状态
func (s *losslessStore) ActiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	// 取消其它 active
	if _, err := dbTx.Exec("UPDATE lossless_rule_release SET active = 0, mtime = sysdate() WHERE rule_id = ? AND active = 1 AND release_type = ?", release.RuleId, release.ReleaseType); err != nil {
		return store.Error(err)
	}
	// 重新计算最大版本号
	var maxVersion uint64
	if err := dbTx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM lossless_rule_release WHERE rule_id = ?", release.RuleId).Scan(&maxVersion); err != nil {
		return store.Error(err)
	}
	if _, err := dbTx.Exec("UPDATE lossless_rule_release SET active = 1, version = ?, mtime = sysdate() WHERE name = ? AND rule_id = ?", maxVersion+1, release.ReleaseName, release.RuleId); err != nil {
		return store.Error(err)
	}
	return nil
}

// InactiveLosslessRule 设置某个无损规则的发布为不使用状态
func (s *losslessStore) InactiveLosslessRule(tx store.Tx, release *rules.LosslessRuleRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	if release.RuleId == "" && release.Rule != nil {
		release.RuleId = release.Rule.ID
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	if _, err := dbTx.Exec("UPDATE lossless_rule_release SET active = 0, mtime = sysdate() WHERE rule_id = ? AND name = ? AND active = 1 AND release_type = ?", release.RuleId, release.ReleaseName, release.ReleaseType); err != nil {
		return store.Error(err)
	}
	return nil
}

// PublishLosslessRule 发布无损规则
func (s *losslessStore) PublishLosslessRules(tx store.Tx, rule *rules.LosslessRuleRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	if rule.Rule == nil || rule.Rule.ID == "" || rule.ReleaseName == "" || rule.ReleaseType == "" {
		return errors.New("[store][mysql][lossless] publish lossless rule missing params")
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	// inactive 同类型
	if _, err := dbTx.Exec("UPDATE lossless_rule_release SET active = 0, mtime = sysdate() WHERE rule_id = ? AND active = 1 AND release_type = ?", rule.Rule.ID, rule.ReleaseType); err != nil {
		return store.Error(err)
	}
	var maxVersion uint64
	if err := dbTx.QueryRow("SELECT IFNULL(MAX(version), 0) FROM lossless_rule_release WHERE rule_id = ?", rule.Rule.ID).Scan(&maxVersion); err != nil {
		return store.Error(err)
	}
	ruleJson, err := protojson.Marshal(rule.Rule.Proto)
	if err != nil {
		return err
	}
	insertSql := `INSERT INTO lossless_rule_release(
id, name, rule_id, namespace, service, rule, flag, version, active, description, release_type, ctime, mtime)
VALUES(?,?,?,?,?,?,?,?,1,?,?,sysdate(),sysdate())`
	args := []any{
		rule.Id,
		rule.ReleaseName,
		rule.Rule.ID,
		rule.Rule.Namespace,
		rule.Rule.Service,
		string(ruleJson),
		0,
		maxVersion + 1,
		rule.Description,
		rule.ReleaseType,
	}
	if _, err := dbTx.Exec(insertSql, args...); err != nil {
		return store.Error(err)
	}
	return nil
}

// 获取已发布的无损规则
func (s *losslessStore) GetMoreLosslessReleases(mtime time.Time, firstUpdate bool) ([]*rules.LosslessRuleRelease, error) {
	str := `SELECT id, name, rule_id, namespace, service, rule, flag, active, version, description, release_type, unix_timestamp(mtime)
FROM lossless_rule_release WHERE mtime > FROM_UNIXTIME(?)`
	if firstUpdate {
		str += " AND flag != 1"
	}
	rows, err := s.slave.Query(str, timeToTimestamp(mtime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*rules.LosslessRuleRelease
	for rows.Next() {
		var (
			id, name, ruleID, namespace, service, ruleStr, description, releaseType string
			flag, active                                                            int
			version                                                                 uint64
			mtimeTs                                                                 int64
		)
		if err := rows.Scan(&id, &name, &ruleID, &namespace, &service, &ruleStr, &flag, &active, &version, &description, &releaseType, &mtimeTs); err != nil {
			return nil, err
		}
		var ruleObj = &rules.LosslessRule{Proto: &traffic_manage.LosslessRule{}}
		if err := json.Unmarshal([]byte(ruleStr), &ruleObj.Proto); err != nil {
			log.Errorf("[store][mysql][lossless] unmarshal release rule err: %v", err)
		}
		ruleObj.ID = ruleID
		ruleObj.Namespace = namespace
		ruleObj.Service = service
		release := &rules.LosslessRuleRelease{
			RuleRelease: rules.RuleRelease{
				Id:          id,
				ReleaseName: name,
				RuleId:      ruleID,
				RuleName:    service,
				Description: description,
				ReleaseType: rules.ReleaseType(releaseType),
				Active:      active == 1,
				Version:     version,
				Valid:       flag == 0,
				Mtime:       time.Unix(mtimeTs, 0),
			},
			Rule: ruleObj,
		}
		out = append(out, release)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (l *losslessStore) DeleteLosslessReleases(tx store.Tx, rule *rules.LosslessRuleRelease) error {
	if tx == nil {
		return ErrTxIsNil
	}
	dbTx := tx.GetDelegateTx().(*BaseTx)
	_, err := dbTx.Exec(`UPDATE lossless_rule_release SET flag = 1, mtime = sysdate() WHERE id = ?`, rule.Id)
	return store.Error(err)
}
