package sqldb

import (
	"database/sql"
	"errors"
	"time"

	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/store"
)

type namespaceConfigTemplateDraftStore struct {
	master *BaseDB
	slave  *BaseDB
}

func (s *namespaceConfigTemplateDraftStore) GetNamespaceConfigTemplateDraft(
	namespace string, templateID uint64) (*conftypes.NamespaceConfigTemplateDraft, error) {
	row := s.slave.QueryRow(`SELECT namespace, template_id, content, format,
		IFNULL(parameter_schema, ''), engine, engine_version, revision, draft_version,
		initialized_from, IFNULL(create_by, ''), IFNULL(modify_by, ''),
		UNIX_TIMESTAMP(ctime), UNIX_TIMESTAMP(mtime)
		FROM namespace_config_template_draft WHERE namespace = ? AND template_id = ?`,
		namespace, templateID)
	draft := &conftypes.NamespaceConfigTemplateDraft{}
	var ctime, mtime int64
	if err := row.Scan(&draft.Namespace, &draft.TemplateID, &draft.Content, &draft.Format,
		&draft.ParameterSchema, &draft.Engine, &draft.EngineVersion, &draft.Revision,
		&draft.DraftVersion, &draft.InitializedFrom, &draft.CreateBy, &draft.ModifyBy,
		&ctime, &mtime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, store.Error(err)
	}
	draft.CreateTime = time.Unix(ctime, 0)
	draft.ModifyTime = time.Unix(mtime, 0)
	return draft, nil
}

func (s *namespaceConfigTemplateDraftStore) SaveNamespaceConfigTemplateDraft(
	draft *conftypes.NamespaceConfigTemplateDraft, expectedVersion uint64) error {
	tx, err := s.master.Begin()
	if err != nil {
		return store.Error(err)
	}
	defer func() { _ = tx.Rollback() }()
	var current uint64
	err = tx.QueryRow(`SELECT draft_version FROM namespace_config_template_draft
		WHERE namespace = ? AND template_id = ? FOR UPDATE`, draft.Namespace, draft.TemplateID).Scan(&current)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		if expectedVersion != 0 {
			return store.ErrNamespaceConfigTemplateDraftConflict
		}
		draft.DraftVersion = 1
		_, err = tx.Exec(`INSERT INTO namespace_config_template_draft (
			namespace, template_id, content, format, parameter_schema, engine, engine_version,
			revision, draft_version, initialized_from, create_by, modify_by, ctime, mtime
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, sysdate(), sysdate())`,
			draft.Namespace, draft.TemplateID, draft.Content, draft.Format, draft.ParameterSchema,
			draft.Engine, draft.EngineVersion, draft.Revision, draft.InitializedFrom,
			draft.CreateBy, draft.ModifyBy)
	case err != nil:
		return store.Error(err)
	case current != expectedVersion:
		return store.ErrNamespaceConfigTemplateDraftConflict
	default:
		draft.DraftVersion = current + 1
		_, err = tx.Exec(`UPDATE namespace_config_template_draft SET content = ?, format = ?,
			parameter_schema = ?, engine = ?, engine_version = ?, revision = ?, draft_version = ?,
			initialized_from = ?, modify_by = ?, mtime = sysdate()
			WHERE namespace = ? AND template_id = ?`, draft.Content, draft.Format,
			draft.ParameterSchema, draft.Engine, draft.EngineVersion, draft.Revision,
			draft.DraftVersion, draft.InitializedFrom, draft.ModifyBy, draft.Namespace, draft.TemplateID)
	}
	if err != nil {
		return store.Error(err)
	}
	return store.Error(tx.Commit())
}
