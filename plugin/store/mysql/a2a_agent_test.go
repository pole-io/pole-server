package sqldb

import (
	"strings"
	"testing"

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
