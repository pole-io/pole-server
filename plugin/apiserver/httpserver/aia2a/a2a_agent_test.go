package aia2a

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
)

func TestParseA2AAgentQuery(t *testing.T) {
	streaming := "true"
	query := parseA2AAgentQuery(map[string][]string{
		"name":                      {"plan"},
		"namespace":                 {"default"},
		"business":                  {"ai"},
		"department":                {"platform"},
		"protocol_binding":          {"JSONRPC"},
		"skill_tag":                 {"planning"},
		"backend_type":              {"service"},
		"backend_service_namespace": {"default"},
		"backend_service_name":      {"planner-agent"},
		"streaming":                 {streaming},
		"offset":                    {"2"},
		"limit":                     {"20"},
	})

	require.NotNil(t, query.Streaming)
	assert.True(t, *query.Streaming)
	assert.Equal(t, &aitypes.A2AAgentQuery{
		Name:                    "plan",
		Namespace:               "default",
		Business:                "ai",
		Department:              "platform",
		ProtocolBinding:         "JSONRPC",
		SkillTag:                "planning",
		BackendType:             "service",
		BackendServiceNamespace: "default",
		BackendServiceName:      "planner-agent",
		Streaming:               query.Streaming,
		Offset:                  2,
		Limit:                   20,
	}, query)
}

func TestPoleSystemA2AProjectionCannotUseGenericMutationPath(t *testing.T) {
	require.True(t, isSelfManagedA2AAgent(&aitypes.A2AAgent{
		Name: "pole-control-plane", Namespace: "pole-system",
	}))
	require.False(t, isSelfManagedA2AAgent(&aitypes.A2AAgent{
		Name: "planner", Namespace: "default",
	}))
}

func TestNewA2AAgentListResponse(t *testing.T) {
	agents := []*aitypes.A2AAgent{{Id: "agent-1", Name: "planner"}}

	resp := newA2AAgentListResponse(2, agents)

	assert.Equal(t, uint32(200000), resp.Code)
	assert.Equal(t, uint32(2), resp.Amount)
	assert.Equal(t, uint32(1), resp.Size)
	assert.Equal(t, agents, resp.Data)
}

func TestStartA2AAgentCache_UpdatesImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := &testA2AAgentCache{}
	require.NoError(t, startA2AAgentCache(ctx, cache, time.Hour))
	assert.Equal(t, int32(1), cache.updateCount.Load())
}

func TestStartA2AAgentCache_ReturnsUpdateError(t *testing.T) {
	expected := errors.New("update failed")
	cache := &testA2AAgentCache{updateErr: expected}

	err := startA2AAgentCache(context.Background(), cache, time.Hour)
	require.ErrorIs(t, err, expected)
	assert.Equal(t, int32(1), cache.updateCount.Load())
}

type testA2AAgentCache struct {
	updateCount atomic.Int32
	updateErr   error
}

func (t *testA2AAgentCache) Initialize(map[string]interface{}) error { return nil }

func (t *testA2AAgentCache) Update() error {
	t.updateCount.Add(1)
	return t.updateErr
}

func (t *testA2AAgentCache) Clear() error { return nil }

func (t *testA2AAgentCache) Name() string { return "a2aAgent" }

func (t *testA2AAgentCache) Close() error { return nil }

func (t *testA2AAgentCache) GetA2AAgentByID(string) *aitypes.A2AAgent { return nil }

func (t *testA2AAgentCache) GetA2AAgentByName(string, string) *aitypes.A2AAgent {
	return nil
}

func (t *testA2AAgentCache) GetA2AAgentsByNamespace(string) []*aitypes.A2AAgent { return nil }

func (t *testA2AAgentCache) GetA2AAgentSkills(string) []*aitypes.A2AAgentSkill { return nil }

func (t *testA2AAgentCache) Query(*aitypes.A2AAgentQuery) (uint32, []*aitypes.A2AAgent) {
	return 0, nil
}
