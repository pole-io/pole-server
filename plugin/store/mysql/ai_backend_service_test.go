package sqldb

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	storeapi "github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/specification/source/go/api/v1/ai"
)

func TestResolveAIBackendServiceIDRequiresSameNamespace(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := rawDB.Begin()
	require.NoError(t, err)

	_, err = resolveAIBackendServiceID(
		&BaseTx{Tx: tx}, "test", "service", "production", "checkout")
	require.Equal(t, storeapi.DataConflictErr, storeapi.Code(err))
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResolveAIBackendServiceIDRejectsAliasOrMissingService(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM service").
		WithArgs("production", "checkout").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	tx, err := rawDB.Begin()
	require.NoError(t, err)

	_, err = resolveAIBackendServiceID(
		&BaseTx{Tx: tx}, "production", "service", "production", "checkout")
	require.Equal(t, storeapi.NotFoundService, storeapi.Code(err))
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateMCPServerPersistsStableBackendServiceID(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	serverStore := &mcpServerStore{master: &BaseDB{DB: rawDB}}
	server := &ai.MCPServer{
		Id: "mcp-1", Name: "checkout-mcp", Namespace: "production", Revision: "rev-1",
		BackendType: "service", BackendServiceNamespace: "production",
		BackendServiceName: "checkout",
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM service").
		WithArgs("production", "checkout").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("service-1"))
	mock.ExpectExec("INSERT INTO mcp_server").
		WithArgs(
			"mcp-1", "checkout-mcp", "production", "", "", "", "", "rev-1", uint32(0),
			"checkout", "", "", "service", "production", "checkout", "service-1", "",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, serverStore.CreateMCPServer(server))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateMCPServerClearsStableBackendServiceIDForAddress(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	serverStore := &mcpServerStore{master: &BaseDB{DB: rawDB}}
	server := &ai.MCPServer{
		Id: "mcp-1", Name: "external-mcp", Namespace: "production", Revision: "rev-2",
		BackendType: "address", BackendServiceNamespace: "production",
		BackendServiceName: "checkout", BackendAddress: "https://mcp.example.com",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE mcp_server SET").
		WithArgs(
			"external-mcp", "production", "", "", "", "", "rev-2", "", "", "",
			"address", "", "", nil, "https://mcp.example.com", uint32(0), "mcp-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, serverStore.UpdateMCPServer(server))
	require.Empty(t, server.BackendServiceNamespace)
	require.Empty(t, server.BackendServiceName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateA2AAgentPersistsStableBackendServiceID(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	agentStore := &a2aAgentStore{master: &BaseDB{DB: rawDB}}
	agent := &aitypes.A2AAgent{
		Id: "agent-1", Name: "planner", Namespace: "production",
		BackendType: "service", BackendServiceNamespace: "production",
		BackendServiceName: "checkout",
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM service").
		WithArgs("production", "checkout").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("service-1"))
	mock.ExpectExec("INSERT INTO a2a_agent").
		WithArgs(
			"agent-1", "planner", "production", "", "", "", "", "", "", "", "", "", "",
			"service", "production", "checkout", "service-1", "", "", "", "", false, false,
			false, "", "", "", "", "", "{}", uint32(0),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE a2a_agent_interface SET flag = 1, mtime = sysdate() WHERE agent_id = ?`)).
		WithArgs("agent-1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE a2a_agent_skill SET flag = 1, mtime = sysdate() WHERE agent_id = ?`)).
		WithArgs("agent-1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, agentStore.CreateA2AAgent(agent))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateA2AAgentClearsStableBackendServiceIDForAddress(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	agentStore := &a2aAgentStore{master: &BaseDB{DB: rawDB}}
	agent := &aitypes.A2AAgent{
		Id: "agent-1", Name: "planner", Namespace: "production",
		BackendType: "address", BackendServiceNamespace: "production",
		BackendServiceName: "checkout",
		BackendAddress:     "https://agent.example.com",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE a2a_agent SET").
		WithArgs(
			"planner", "production", "", "", "", "", "", "", "", "", "", "", "address",
			"", "", nil, "https://agent.example.com", "", "", "", false, false, false, "",
			"", "", "", "", "{}", uint32(0), "agent-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE a2a_agent_interface SET flag = 1").
		WithArgs("agent-1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE a2a_agent_skill SET flag = 1").
		WithArgs("agent-1").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, agentStore.UpdateA2AAgent(agent))
	require.Empty(t, agent.BackendServiceNamespace)
	require.Empty(t, agent.BackendServiceName)
	require.Equal(t, "https://agent.example.com", agent.BackendAddress)
	require.NoError(t, mock.ExpectationsWereMet())
}
