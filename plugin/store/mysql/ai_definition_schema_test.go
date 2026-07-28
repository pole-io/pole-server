package sqldb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestEnsureAIResourceDefinitionSchemaCreatesDefinitionsAndKeepsExistingColumns(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS mcp_server_definition").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS a2a_agent_definition").
		WillReturnResult(sqlmock.NewResult(0, 0))
	expectAIEnvironmentSchemaState(mock, "mcp_server",
		"idx_mcp_server_definition", "uk_mcp_server_definition_namespace",
		"idx_mcp_server_backend_service_id", 1)
	expectAIEnvironmentSchemaState(mock, "a2a_agent",
		"idx_a2a_agent_definition", "uk_a2a_agent_definition_namespace",
		"idx_a2a_agent_backend_service_id", 1)

	require.NoError(t, ensureAIResourceDefinitionSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureAIResourceDefinitionSchemaUpgradesEnvironmentTables(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS mcp_server_definition").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS a2a_agent_definition").
		WillReturnResult(sqlmock.NewResult(0, 0))
	expectAIEnvironmentSchemaUpgrade(mock, "mcp_server",
		"idx_mcp_server_definition", "uk_mcp_server_definition_namespace",
		"idx_mcp_server_backend_service_id")
	expectAIEnvironmentSchemaUpgrade(mock, "a2a_agent",
		"idx_a2a_agent_definition", "uk_a2a_agent_definition_namespace",
		"idx_a2a_agent_backend_service_id")

	require.NoError(t, ensureAIResourceDefinitionSchema(&BaseDB{DB: rawDB}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func expectAIEnvironmentSchemaState(
	mock sqlmock.Sqlmock, table, definitionIndex, environmentKey, backendIndex string, count int) {
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs(table, "definition_id").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, definitionIndex).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, environmentKey).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs(table, "backend_service_id").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, backendIndex).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectExec("UPDATE " + table + " AS ai.*INNER JOIN service AS backend").
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func expectAIEnvironmentSchemaUpgrade(
	mock sqlmock.Sqlmock, table, definitionIndex, environmentKey, backendIndex string) {
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs(table, "definition_id").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE " + table + " ADD COLUMN definition_id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, definitionIndex).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE " + table + " ADD KEY " + definitionIndex).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, environmentKey).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE " + table + " ADD UNIQUE KEY " + environmentKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.columns").
		WithArgs(table, "backend_service_id").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE " + table + " ADD COLUMN backend_service_id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("(?s)select count\\(\\*\\).*information_schema.statistics").
		WithArgs(table, backendIndex).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE " + table + " ADD KEY " + backendIndex).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE " + table + " AS ai.*INNER JOIN service AS backend").
		WillReturnResult(sqlmock.NewResult(0, 1))
}
