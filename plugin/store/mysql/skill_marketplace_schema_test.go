package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestEnsureSkillMarketplaceSchemaCreatesAllTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	for _, table := range []string{"skill_publisher", "skill_marketplace_skill", "skill_publisher_key",
		"skill_publisher_member", "skill_bundle_blob", "skill_release", "skill_release_review",
		"skill_access_grant", "skill_registry_source"} {
		mock.ExpectExec("CREATE TABLE IF NOT EXISTS " + table).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	require.NoError(t, ensureSkillMarketplaceSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}
