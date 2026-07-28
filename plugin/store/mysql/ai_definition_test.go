package sqldb

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	storeapi "github.com/pole-io/pole-server/apis/store"
)

func TestCreateAIResourceDefinitionRoutesByKind(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	for _, test := range []struct {
		kind  aitypes.ResourceKind
		table string
		id    string
	}{
		{kind: aitypes.ResourceKindMCPServer, table: "mcp_server_definition", id: "mcp-def"},
		{kind: aitypes.ResourceKindA2AAgent, table: "a2a_agent_definition", id: "a2a-def"},
	} {
		definition := &aitypes.ResourceDefinition{
			Kind: test.kind, ID: test.id, Name: "checkout", Description: "desc",
			Owner: "team-a", Business: "payments", Department: "platform", Revision: "rev-1",
		}
		mock.ExpectExec("INSERT INTO "+test.table).
			WithArgs(test.id, "checkout", "desc", "team-a", "payments", "platform", "rev-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		require.NoError(t, definitionStore.CreateAIResourceDefinition(definition))
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAIResourceDefinitionRejectsRevisionConflict(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}
	definition := &aitypes.ResourceDefinition{
		Kind: aitypes.ResourceKindMCPServer, ID: "def-1", Name: "checkout", Revision: "rev-2",
	}

	mock.ExpectExec("UPDATE mcp_server_definition SET name =").
		WithArgs("checkout", "", "", "", "", "rev-2", "def-1", "rev-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = definitionStore.UpdateAIResourceDefinition(definition, "rev-1")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBindAIResourceEnvironmentPersistsExplicitBinding(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}
	binding := &aitypes.EnvironmentBinding{
		Kind: aitypes.ResourceKindMCPServer, DefinitionID: "def-1", ResourceID: "server-1",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id FROM mcp_server_definition WHERE id = ? AND flag = 0 FOR UPDATE`)).
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT namespace, name, definition_id FROM mcp_server WHERE id = ? AND flag != 1 FOR UPDATE`)).
		WithArgs("server-1").
		WillReturnRows(sqlmock.NewRows([]string{"namespace", "name", "definition_id"}).
			AddRow("production", "checkout-mcp", nil))
	mock.ExpectExec("UPDATE mcp_server SET definition_id =").
		WithArgs("def-1", "server-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mcp_server_definition SET revision =").
		WithArgs("rev-2", "def-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, definitionStore.BindAIResourceEnvironment(binding, "rev-2"))
	require.Equal(t, "production", binding.Namespace)
	require.Equal(t, "checkout-mcp", binding.ResourceName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBindAIResourceEnvironmentRejectsResourceBoundToAnotherDefinition(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM a2a_agent_definition").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectQuery("SELECT namespace, name, definition_id FROM a2a_agent").
		WithArgs("agent-1").
		WillReturnRows(sqlmock.NewRows([]string{"namespace", "name", "definition_id"}).
			AddRow("test", "planner", "def-2"))
	mock.ExpectRollback()

	err = definitionStore.BindAIResourceEnvironment(&aitypes.EnvironmentBinding{
		Kind: aitypes.ResourceKindA2AAgent, DefinitionID: "def-1", ResourceID: "agent-1",
	}, "rev-2")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBindAIResourceEnvironmentMapsNamespaceUniquenessToConflict(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM mcp_server_definition").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectQuery("SELECT namespace, name, definition_id FROM mcp_server").
		WithArgs("server-2").
		WillReturnRows(sqlmock.NewRows([]string{"namespace", "name", "definition_id"}).
			AddRow("production", "checkout-mcp-v2", nil))
	mock.ExpectExec("UPDATE mcp_server SET definition_id =").
		WithArgs("def-1", "server-2").
		WillReturnError(&gomysql.MySQLError{Number: 1062, Message: "Duplicate entry for environment"})
	mock.ExpectRollback()

	err = definitionStore.BindAIResourceEnvironment(&aitypes.EnvironmentBinding{
		Kind: aitypes.ResourceKindMCPServer, DefinitionID: "def-1", ResourceID: "server-2",
	}, "rev-2")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnbindAIResourceEnvironmentRejectsMismatchedDefinition(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM mcp_server_definition").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectQuery("SELECT definition_id FROM mcp_server").
		WithArgs("server-1").
		WillReturnRows(sqlmock.NewRows([]string{"definition_id"}).AddRow("def-2"))
	mock.ExpectRollback()

	err = definitionStore.UnbindAIResourceEnvironment(
		aitypes.ResourceKindMCPServer, "def-1", "server-1", "rev-2")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnbindAIResourceEnvironmentClearsBindingAndBumpsRevision(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM a2a_agent_definition").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectQuery("SELECT definition_id FROM a2a_agent").
		WithArgs("agent-1").
		WillReturnRows(sqlmock.NewRows([]string{"definition_id"}).AddRow("def-1"))
	mock.ExpectExec("UPDATE a2a_agent SET definition_id = NULL").
		WithArgs("agent-1", "def-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE a2a_agent_definition SET revision =").
		WithArgs("rev-2", "def-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, definitionStore.UnbindAIResourceEnvironment(
		aitypes.ResourceKindA2AAgent, "def-1", "agent-1", "rev-2"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAIResourceDefinitionRejectsActiveBindings(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM a2a_agent_definition").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("def-1"))
	mock.ExpectExec("UPDATE a2a_agent SET definition_id = NULL").
		WithArgs("def-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM a2a_agent").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	err = definitionStore.DeleteAIResourceDefinition(aitypes.ResourceKindA2AAgent, "def-1")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAIResourceEnvironmentBindingReturnsNilForUnboundResource(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{slave: &BaseDB{DB: rawDB}}

	mock.ExpectQuery("SELECT definition_id, id, namespace, name FROM mcp_server").
		WithArgs("server-1").
		WillReturnError(sql.ErrNoRows)

	binding, err := definitionStore.GetAIResourceEnvironmentBinding(
		aitypes.ResourceKindMCPServer, "server-1")
	require.NoError(t, err)
	require.Nil(t, binding)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAIResourceDefinitionsAndBindingsMapsRows(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{slave: &BaseDB{DB: rawDB}}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM mcp_server_definition").
		WithArgs("%check%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT id, name, .* FROM mcp_server_definition").
		WithArgs("%check%", uint32(0), uint32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "owner", "business", "department", "revision", "ctime", "mtime",
		}).AddRow("def-1", "checkout", "desc", "team-a", "payments", "platform", "rev-1", int64(100), int64(200)))

	total, definitions, err := definitionStore.ListAIResourceDefinitions(
		aitypes.ResourceKindMCPServer, "check", 0, 10)
	require.NoError(t, err)
	require.Equal(t, uint32(1), total)
	require.Len(t, definitions, 1)
	require.Equal(t, "checkout", definitions[0].Name)
	require.Equal(t, aitypes.ResourceKindMCPServer, definitions[0].Kind)

	mock.ExpectQuery("SELECT definition_id, id, namespace, name FROM mcp_server").
		WithArgs("def-1").
		WillReturnRows(sqlmock.NewRows([]string{"definition_id", "id", "namespace", "name"}).
			AddRow("def-1", "server-1", "production", "checkout-mcp"))
	bindings, err := definitionStore.ListAIResourceEnvironmentBindings(
		aitypes.ResourceKindMCPServer, "def-1")
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	require.Equal(t, "production", bindings[0].Namespace)
	require.Equal(t, "server-1", bindings[0].ResourceID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountAIResourcesByNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectQuery("SELECT.*COUNT\\(\\*\\) FROM mcp_server.*COUNT\\(\\*\\) FROM a2a_agent").
		WithArgs("production", "production").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := definitionStore.CountAIResourcesByNamespace("production")
	require.NoError(t, err)
	require.Equal(t, uint32(5), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountAIBackendReferencesIncludesLegacyNameBindingFallback(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	definitionStore := &aiDefinitionStore{master: &BaseDB{DB: rawDB}}

	mock.ExpectQuery(
		"SELECT.*COUNT\\(\\*\\) FROM mcp_server.*backend_service_id IS NULL.*"+
			"backend_service_namespace.*backend_service_name.*COUNT\\(\\*\\) FROM a2a_agent").
		WithArgs("service-1", "service-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := definitionStore.CountAIBackendReferences("service-1")
	require.NoError(t, err)
	require.Equal(t, uint32(2), count)
	require.NoError(t, mock.ExpectationsWereMet())
}
