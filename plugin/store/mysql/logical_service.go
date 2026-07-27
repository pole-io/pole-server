package sqldb

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
)

type logicalServiceStore struct {
	master *BaseDB
	slave  *BaseDB
}

const logicalServiceSelect = `SELECT id, name, IFNULL(comment, ''), IFNULL(owner, ''),
	IFNULL(business, ''), IFNULL(department, ''), revision, flag,
	UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime) FROM logical_service`

func (s *logicalServiceStore) CreateLogicalService(service *svctypes.LogicalService) error {
	_, err := s.master.Exec(`INSERT INTO logical_service
		(id, name, comment, owner, business, department, revision, flag, ctime, mtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, sysdate(), sysdate())`,
		service.ID, service.Name, service.Comment, service.Owner, service.Business,
		service.Department, service.Revision)
	return store.Error(err)
}

func (s *logicalServiceStore) UpdateLogicalService(
	service *svctypes.LogicalService, previousRevision string) error {
	result, err := s.master.Exec(`UPDATE logical_service SET name = ?, comment = ?, owner = ?,
		business = ?, department = ?, revision = ?, mtime = sysdate()
		WHERE id = ? AND flag = 0 AND revision = ?`,
		service.Name, service.Comment, service.Owner, service.Business, service.Department,
		service.Revision, service.ID, previousRevision)
	if err != nil {
		return store.Error(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return store.Error(err)
	}
	if affected == 0 {
		return store.NewStatusError(store.DataConflictErr, "logical service revision conflict")
	}
	return nil
}

func (s *logicalServiceStore) DeleteLogicalService(id string) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	var logicalServiceID string
	if err := tx.QueryRow(`SELECT id FROM logical_service
		WHERE id = ? AND flag = 0 FOR UPDATE`, id).Scan(&logicalServiceID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "logical service not found")
	} else if err != nil {
		return store.Error(err)
	}
	var count uint32
	if err := tx.QueryRow(`SELECT COUNT(*) FROM service_environment_binding
		WHERE logical_service_id = ?`, id).Scan(&count); err != nil {
		return store.Error(err)
	}
	if count > 0 {
		return store.NewStatusError(store.DataConflictErr, "logical service still has environment bindings")
	}
	if _, err := tx.Exec(`UPDATE logical_service SET flag = 1, mtime = sysdate()
		WHERE id = ? AND flag = 0`, id); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

func (s *logicalServiceStore) GetLogicalService(id string) (*svctypes.LogicalService, error) {
	return scanLogicalService(s.slave.QueryRow(logicalServiceSelect+` WHERE id = ? AND flag = 0`, id))
}

func (s *logicalServiceStore) ListLogicalServices(
	name string, offset, limit uint32) (uint32, []*svctypes.LogicalService, error) {
	where := ` WHERE flag = 0`
	args := []any{}
	if name != "" {
		where += ` AND name LIKE ?`
		args = append(args, "%"+name+"%")
	}
	var total uint32
	if err := s.slave.QueryRow(`SELECT COUNT(*) FROM logical_service`+where, args...).Scan(&total); err != nil {
		return 0, nil, store.Error(err)
	}
	args = append(args, offset, limit)
	rows, err := s.slave.Query(logicalServiceSelect+where+` ORDER BY mtime DESC, id LIMIT ?, ?`, args...)
	if err != nil {
		return 0, nil, store.Error(err)
	}
	defer rows.Close()
	services := make([]*svctypes.LogicalService, 0)
	for rows.Next() {
		item, err := scanLogicalService(rows)
		if err != nil {
			return 0, nil, err
		}
		services = append(services, item)
	}
	return total, services, store.Error(rows.Err())
}

func scanLogicalService(row rowScanner) (*svctypes.LogicalService, error) {
	item := &svctypes.LogicalService{}
	var flag int
	var ctime, mtime int64
	if err := row.Scan(&item.ID, &item.Name, &item.Comment, &item.Owner, &item.Business,
		&item.Department, &item.Revision, &flag, &ctime, &mtime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, store.Error(err)
	}
	item.Valid = flag == 0
	item.CreateTime = time.Unix(ctime, 0)
	item.ModifyTime = time.Unix(mtime, 0)
	return item, nil
}

func (s *logicalServiceStore) BindServiceEnvironment(
	binding *svctypes.ServiceEnvironmentBinding, revision string) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	var logicalServiceID string
	if err := tx.QueryRow(`SELECT id FROM logical_service
		WHERE id = ? AND flag = 0 FOR UPDATE`, binding.LogicalServiceID).Scan(&logicalServiceID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "logical service not found")
	} else if err != nil {
		return store.Error(err)
	}
	var existingLogicalServiceID string
	err = tx.QueryRow(`SELECT logical_service_id FROM service_environment_binding
		WHERE service_id = ? FOR UPDATE`, binding.ServiceID).Scan(&existingLogicalServiceID)
	if err == nil {
		if existingLogicalServiceID == binding.LogicalServiceID {
			return store.Error(tx.Commit())
		}
		return store.NewStatusError(store.DataConflictErr, "environment service is already bound")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return store.Error(err)
	}
	_, err = tx.Exec(`INSERT INTO service_environment_binding
		(logical_service_id, service_id, namespace, service_name, ctime, mtime)
		VALUES (?, ?, ?, ?, sysdate(), sysdate())`,
		binding.LogicalServiceID, binding.ServiceID, binding.Namespace, binding.ServiceName)
	if err != nil {
		if store.Code(store.Error(err)) == store.DuplicateEntryErr {
			return store.NewStatusError(store.DataConflictErr, err.Error())
		}
		return store.Error(err)
	}
	if _, err := tx.Exec(`UPDATE logical_service SET revision = ?, mtime = sysdate()
		WHERE id = ? AND flag = 0`, revision, binding.LogicalServiceID); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

func (s *logicalServiceStore) UnbindServiceEnvironment(
	logicalServiceID, serviceID, revision string) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	var lockedLogicalServiceID string
	if err := tx.QueryRow(`SELECT id FROM logical_service
		WHERE id = ? AND flag = 0 FOR UPDATE`, logicalServiceID).Scan(&lockedLogicalServiceID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "logical service not found")
	} else if err != nil {
		return store.Error(err)
	}
	var actualLogicalServiceID string
	if err := tx.QueryRow(`SELECT logical_service_id FROM service_environment_binding
		WHERE service_id = ? FOR UPDATE`, serviceID).Scan(&actualLogicalServiceID); errors.Is(err, sql.ErrNoRows) {
		return store.NewStatusError(store.NotFoundService, "environment binding not found")
	} else if err != nil {
		return store.Error(err)
	}
	if actualLogicalServiceID != logicalServiceID {
		return store.NewStatusError(store.DataConflictErr, "environment binding belongs to another logical service")
	}
	if _, err := tx.Exec(`DELETE FROM service_environment_binding
		WHERE logical_service_id = ? AND service_id = ?`, logicalServiceID, serviceID); err != nil {
		return store.Error(err)
	}
	if _, err := tx.Exec(`UPDATE logical_service SET revision = ?, mtime = sysdate()
		WHERE id = ? AND flag = 0`, revision, logicalServiceID); err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}

const environmentBindingSelect = `SELECT logical_service_id, service_id, namespace,
	service_name, UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
	FROM service_environment_binding`

func (s *logicalServiceStore) ListServiceEnvironmentBindings(
	logicalServiceID string) ([]*svctypes.ServiceEnvironmentBinding, error) {
	query := environmentBindingSelect
	args := []any{}
	if logicalServiceID != "" {
		query += ` WHERE logical_service_id = ?`
		args = append(args, logicalServiceID)
	}
	rows, err := s.slave.Query(query+` ORDER BY namespace, service_name`, args...)
	if err != nil {
		return nil, store.Error(err)
	}
	return scanEnvironmentBindings(rows)
}

func (s *logicalServiceStore) ListServiceEnvironmentBindingsByLogicalIDs(
	logicalServiceIDs []string) (map[string][]*svctypes.ServiceEnvironmentBinding, error) {
	result := make(map[string][]*svctypes.ServiceEnvironmentBinding, len(logicalServiceIDs))
	if len(logicalServiceIDs) == 0 {
		return result, nil
	}
	args := make([]any, 0, len(logicalServiceIDs))
	for _, id := range logicalServiceIDs {
		args = append(args, id)
	}
	rows, err := s.slave.Query(environmentBindingSelect+` WHERE logical_service_id IN (`+
		strings.TrimSuffix(strings.Repeat("?,", len(args)), ",")+`) ORDER BY namespace, service_name`, args...)
	if err != nil {
		return nil, store.Error(err)
	}
	bindings, err := scanEnvironmentBindings(rows)
	if err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		result[binding.LogicalServiceID] = append(result[binding.LogicalServiceID], binding)
	}
	return result, nil
}

func (s *logicalServiceStore) GetServiceEnvironmentBinding(
	serviceID string) (*svctypes.ServiceEnvironmentBinding, error) {
	item := &svctypes.ServiceEnvironmentBinding{}
	var ctime, mtime int64
	err := s.slave.QueryRow(environmentBindingSelect+` WHERE service_id = ?`, serviceID).
		Scan(&item.LogicalServiceID, &item.ServiceID, &item.Namespace, &item.ServiceName, &ctime, &mtime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, store.Error(err)
	}
	item.CreateTime = time.Unix(ctime, 0)
	item.ModifyTime = time.Unix(mtime, 0)
	return item, nil
}

func scanEnvironmentBindings(rows *sql.Rows) ([]*svctypes.ServiceEnvironmentBinding, error) {
	defer rows.Close()
	items := make([]*svctypes.ServiceEnvironmentBinding, 0)
	for rows.Next() {
		item := &svctypes.ServiceEnvironmentBinding{}
		var ctime, mtime int64
		if err := rows.Scan(&item.LogicalServiceID, &item.ServiceID, &item.Namespace,
			&item.ServiceName, &ctime, &mtime); err != nil {
			return nil, store.Error(err)
		}
		item.CreateTime = time.Unix(ctime, 0)
		item.ModifyTime = time.Unix(mtime, 0)
		items = append(items, item)
	}
	return items, store.Error(rows.Err())
}
