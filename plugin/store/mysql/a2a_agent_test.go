package sqldb

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

func TestNewA2AID_MatchesSchemaLength(t *testing.T) {
	id := newA2AID()

	assert.Len(t, id, 32)
	assert.False(t, strings.Contains(id, "-"))
}

func TestA2AAgentStore_CreateRejectsMissingNameOrNamespace(t *testing.T) {
	store := &a2aAgentStore{}

	require.Error(t, store.CreateA2AAgent(&aitypes.A2AAgent{Name: "", Namespace: "default"}))
	require.Error(t, store.CreateA2AAgent(&aitypes.A2AAgent{Name: "planner", Namespace: ""}))
}

func TestReplaceA2AAgentChildrenRevivesStableRows(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()
	mock.ExpectBegin()
	tx, err := rawDB.Begin()
	require.NoError(t, err)
	baseTx := &BaseTx{Tx: tx}
	agent := &aitypes.A2AAgent{
		Id: "pole-agent",
		Interfaces: []*aitypes.A2AAgentInterface{{
			Id: "interface-id", Url: "http://pole/ai/agent/a2a/v1",
			ProtocolBinding: "jsonrpc", ProtocolVersion: "0.3.0",
		}},
		Skills: []*aitypes.A2AAgentSkill{{
			Id: "skill-row-id", SkillId: "pole-management", Name: "Pole 管理",
		}},
	}
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE a2a_agent_interface SET flag = 1, mtime = sysdate() WHERE agent_id = ?`)).
		WithArgs(agent.Id).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO a2a_agent_interface.*ON DUPLICATE KEY UPDATE").
		WithArgs("interface-id", "pole-agent", "http://pole/ai/agent/a2a/v1", "jsonrpc", "0.3.0", "", uint32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE a2a_agent_skill SET flag = 1, mtime = sysdate() WHERE agent_id = ?`)).
		WithArgs(agent.Id).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO a2a_agent_skill.*ON DUPLICATE KEY UPDATE").
		WithArgs("skill-row-id", "pole-agent", "pole-management", "Pole 管理", "", "[]", "[]", "[]", "[]", "", uint32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	require.NoError(t, (&a2aAgentStore{}).replaceA2AAgentInterfaces(baseTx, agent))
	require.NoError(t, (&a2aAgentStore{}).replaceA2AAgentSkills(baseTx, agent))
	require.NoError(t, baseTx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
