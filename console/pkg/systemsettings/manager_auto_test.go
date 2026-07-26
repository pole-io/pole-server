package systemsettings

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pole-io/pole-server/console/bootstrap"
	"github.com/pole-io/pole-server/console/pkg/agentworkbench"
	"github.com/pole-io/pole-server/console/pkg/poleagent"
)

func TestSaveAndReconcilePublishesHealthyPromptAsSelfManager(t *testing.T) {
	manager := newAutoReconcileTestManager(t, nil)
	view, err := manager.SaveAndReconcile(context.Background(), agentworkbench.Actor{
		UserID: "admin", Token: "admin-token",
	}, validAgentDraft("新的管理员 Prompt 指令"))
	require.NoError(t, err)
	require.Nil(t, view.Draft)
	require.NotNil(t, view.Active)
	require.Equal(t, "admin", view.Active.CreatedBy)
	require.Equal(t, SelfManagerActorID, view.Active.PublishedBy)
	require.Equal(t, "新的管理员 Prompt 指令", view.Active.Values.OperatorInstructions)
	require.Equal(t, "applied", view.ApplyStatus)
	require.NotEqual(t, "static", view.EffectiveRevision)
}

func TestSaveAndReconcileKeepsLastHealthyRuntimeWhenProbeFails(t *testing.T) {
	manager := newAutoReconcileTestManager(t, errors.New("prompt contract probe failed"))
	before := manager.Current()
	view, err := manager.SaveAndReconcile(context.Background(), agentworkbench.Actor{
		UserID: "admin", Token: "admin-token",
	}, validAgentDraft("不健康的 Prompt 指令"))
	require.ErrorContains(t, err, "prompt contract probe failed")
	require.NotNil(t, view)
	require.NotNil(t, view.Draft)
	require.Nil(t, view.Active)
	require.Equal(t, "rejected", view.ApplyStatus)
	require.Same(t, before.Agent, manager.Current().Agent)
	require.Equal(t, before.Revision, manager.Current().Revision)

	manager.systemCandidate = func(context.Context, agentworkbench.Actor, AgentProfile, string) (*poleagent.Agent, error) {
		return poleagent.New(poleagent.Options{
			ID: "pole-control-plane", PromptVersion: "v1",
		}, nil, nil, manager.workbench), nil
	}
	require.NoError(t, manager.reconcilePendingDraft(context.Background()))
	reconciled, err := manager.Domain(context.Background())
	require.NoError(t, err)
	require.Nil(t, reconciled.Draft)
	require.NotNil(t, reconciled.Active)
	require.Equal(t, SelfManagerActorID, reconciled.Active.PublishedBy)
	require.Equal(t, "applied", reconciled.ApplyStatus)
}

func TestPoleServerRegistryEndpointPreservesExplicitScheme(t *testing.T) {
	require.Equal(t, "https://pole.example/ai/mcp/v1/servers",
		poleServerRegistryEndpoint("https://pole.example/"))
	require.Equal(t, "http://127.0.0.1:8090/ai/mcp/v1/servers",
		poleServerRegistryEndpoint("127.0.0.1:8090"))
	require.Empty(t, poleServerRegistryEndpoint(" "))
}

func TestManualRetryStillRecordsSelfManagerAsExecutor(t *testing.T) {
	manager := newAutoReconcileTestManager(t, nil)
	draft, err := manager.SaveDraft(context.Background(), "admin", validAgentDraft("待重试 Prompt"))
	require.NoError(t, err)
	require.NotNil(t, draft.Draft)

	view, err := manager.Publish(context.Background(), agentworkbench.Actor{
		UserID: "admin", Token: "admin-token",
	}, PublishRequest{DraftRevision: draft.Draft.Revision})
	require.NoError(t, err)
	require.Equal(t, "admin", view.Active.CreatedBy)
	require.Equal(t, SelfManagerActorID, view.Active.PublishedBy)
}

func newAutoReconcileTestManager(t *testing.T, probeErr error) *Manager {
	t.Helper()
	repository := NewMemoryRepository()
	envelope, err := newSecretEnvelope(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	require.NoError(t, err)
	workbench := agentworkbench.New(nil, agentworkbench.Options{TTL: time.Minute})
	manager := &Manager{
		config:    &bootstrap.Config{},
		repo:      repository,
		workbench: workbench,
		envelope:  envelope,
		status:    "static",
	}
	staticAgent := poleagent.New(poleagent.Options{ID: "pole-control-plane", PromptVersion: "v1"}, nil, nil, workbench)
	manager.current.Store(&RuntimeSnapshot{
		Revision:  "static",
		Profile:   profileFromBootstrap(bootstrap.AgentConfig{}.Normalize()),
		Agent:     staticAgent,
		AppliedAt: time.Now().UTC(),
	})
	manager.candidate = func(context.Context, agentworkbench.Actor, AgentProfile, string) (*poleagent.Agent, error) {
		if probeErr != nil {
			return nil, probeErr
		}
		return poleagent.New(poleagent.Options{
			ID: "pole-control-plane", PromptVersion: "v1",
		}, nil, nil, workbench), nil
	}
	return manager
}

func validAgentDraft(operatorInstructions string) SaveDraftRequest {
	return SaveDraftRequest{
		Values: AgentProfile{
			RuntimeMode:          "llm",
			AgentID:              "pole-control-plane",
			PromptVersion:        "v1",
			OperatorInstructions: operatorInstructions,
			Provider:             "openai-compatible",
			BaseURL:              "https://llm.example.test",
			Model:                "test-model",
			ModelTimeout:         "10s",
			MCPEndpoint:          "http://pole.example.test/ai/mcp/v1/sse",
			MCPToolAllowlist:     []string{"list_namespaces"},
			ProposalTTL:          "30m",
			UpstreamTimeout:      "10s",
		},
		APIKey: &SecretUpdate{Operation: "replace", Value: "test-api-key"},
	}
}
