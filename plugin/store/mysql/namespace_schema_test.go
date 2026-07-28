package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestEnsureNamespaceKindSchemaBackfillsSystemNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs("namespace", "kind").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("UPDATE namespace SET kind = 1 WHERE name = 'pole-system'").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, ensureNamespaceKindSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureNamespaceKindSchemaUpgradesLegacyTable(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs("namespace", "kind").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE namespace ADD COLUMN kind").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE namespace SET kind = 1 WHERE name = 'pole-system'").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, ensureNamespaceKindSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}
