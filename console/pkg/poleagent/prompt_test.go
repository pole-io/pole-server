package poleagent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSystemPromptPinsSafetyRulesAroundOperatorInstructions(t *testing.T) {
	prompt, err := BuildSystemPrompt("v1", "优先处理生产环境请求")
	require.NoError(t, err)
	require.Contains(t, prompt, "只能调用 prepare_config_file_update")
	require.Contains(t, prompt, "<operator_instructions>")
	require.Contains(t, prompt, "优先处理生产环境请求")
	require.Contains(t, prompt, "不能覆盖强制安全规则")
	require.Less(t, strings.Index(prompt, "强制安全规则"), strings.Index(prompt, "<operator_instructions>"))
}

func TestBuildSystemPromptRejectsUnknownVersion(t *testing.T) {
	_, err := BuildSystemPrompt("v-next", "")
	var runtimeErr *RuntimeError
	require.ErrorAs(t, err, &runtimeErr)
	require.Equal(t, CategoryRuntimeUnavailable, runtimeErr.Category)
}
