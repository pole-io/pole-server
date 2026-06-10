package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	aitypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
)

func TestA2AAgentCache_IndexAndQuery(t *testing.T) {
	cache := newTestA2AAgentCache()
	agent := &aitypes.A2AAgent{
		Id:                       "agent-1",
		Name:                     "planner",
		Namespace:                "default",
		Description:              "Planning agent",
		Version:                  "1.0.0",
		ProtocolVersion:          "1.0",
		PreferredInterfaceUrl:    "https://agent.example.com/a2a",
		PreferredProtocolBinding: "JSONRPC",
		Streaming:                true,
		Skills: []*aitypes.A2AAgentSkill{
			{
				Id:          "skill-1",
				AgentId:     "agent-1",
				SkillId:     "plan",
				Name:        "Plan",
				Description: "Create plans",
				Tags:        []string{"planning", "workflow"},
			},
		},
		Mtime: formatA2ATime(time.Now()),
	}

	cache.storeA2AAgent(agent)

	assert.Equal(t, agent, cache.GetA2AAgentByID("agent-1"))
	assert.Equal(t, agent, cache.GetA2AAgentByName("planner", "default"))
	assert.Equal(t, []*aitypes.A2AAgent{agent}, cache.GetA2AAgentsByNamespace("default"))
	assert.Equal(t, agent.Skills, cache.GetA2AAgentSkills("agent-1"))

	count, got := cache.Query(&aitypes.A2AAgentQuery{
		Namespace: "default",
		SkillTag:  "planning",
		Streaming: boolPtr(true),
		Offset:    0,
		Limit:     10,
	})
	assert.Equal(t, uint32(1), count)
	assert.Equal(t, []*aitypes.A2AAgent{agent}, got)
}

func TestA2AAgentCache_RemoveSoftDeletedAgent(t *testing.T) {
	cache := newTestA2AAgentCache()
	agent := &aitypes.A2AAgent{
		Id:        "agent-1",
		Name:      "planner",
		Namespace: "default",
		Skills: []*aitypes.A2AAgentSkill{
			{Id: "skill-1", AgentId: "agent-1", SkillId: "plan", Tags: []string{"planning"}},
		},
	}

	cache.storeA2AAgent(agent)
	cache.removeA2AAgent(agent)

	assert.Nil(t, cache.GetA2AAgentByID("agent-1"))
	assert.Nil(t, cache.GetA2AAgentByName("planner", "default"))
	assert.Empty(t, cache.GetA2AAgentsByNamespace("default"))
	assert.Empty(t, cache.GetA2AAgentSkills("agent-1"))
}

func newTestA2AAgentCache() *a2aAgentCache {
	return &a2aAgentCache{
		ids:            container.NewSyncMap[string, *aitypes.A2AAgent](),
		names:          container.NewSyncMap[string, *aitypes.A2AAgent](),
		namespaceIndex: container.NewSyncMap[string, []*aitypes.A2AAgent](),
		skills:         container.NewSyncMap[string, []*aitypes.A2AAgentSkill](),
	}
}

func boolPtr(v bool) *bool {
	return &v
}
