package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestEnsureLogicalServiceSchemaCreatesBothTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS logical_service").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS service_environment_binding").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs("logical_service", "active_name").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs("logical_service", "uk_logical_service_active_name").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	require.NoError(t, ensureLogicalServiceSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureLogicalServiceSchemaUpgradesExistingTable(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS logical_service").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS service_environment_binding").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs("logical_service", "active_name").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE logical_service ADD COLUMN active_name").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs("logical_service", "uk_logical_service_active_name").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE logical_service").
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.NoError(t, ensureLogicalServiceSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}
