package anthropic

import (
	"encoding/json"
	"testing"

	a "github.com/anthropics/anthropic-sdk-go"
	"github.com/nikitakunevich/tui-agent/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAnthropicTool(t *testing.T) {
	td := llm.ToolDef{
		Name:        "bash",
		Description: "Execute a bash command",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`),
	}
	got := toAnthropicTool(td)
	require.NotNil(t, got.OfTool)
	assert.Equal(t, "bash", got.OfTool.Name)
}

func TestFromAnthropicMessage_Text(t *testing.T) {
	msg := &a.Message{
		Content: []a.ContentBlockUnion{
			{Type: "text", Text: "Hello!"},
		},
		StopReason: a.StopReasonEndTurn,
	}
	got := fromAnthropicMessage(msg)
	assert.Equal(t, "Hello!", got.Content)
	assert.Equal(t, "end_turn", got.StopReason)
	assert.Empty(t, got.ToolCalls)
}

func TestFromAnthropicMessage_ToolUse(t *testing.T) {
	msg := &a.Message{
		Content: []a.ContentBlockUnion{
			{
				Type:  "tool_use",
				ID:    "toolu_123",
				Name:  "bash",
				Input: json.RawMessage(`{"command":"ls"}`),
			},
		},
		StopReason: a.StopReasonToolUse,
	}
	got := fromAnthropicMessage(msg)
	assert.Equal(t, "tool_use", got.StopReason)
	assert.Len(t, got.ToolCalls, 1)
	assert.Equal(t, "toolu_123", got.ToolCalls[0].ID)
	assert.Equal(t, "bash", got.ToolCalls[0].Name)
	assert.Equal(t, `{"command":"ls"}`, got.ToolCalls[0].Arguments)
}

func TestFromAnthropicMessage_MixedContent(t *testing.T) {
	msg := &a.Message{
		Content: []a.ContentBlockUnion{
			{Type: "text", Text: "Let me run that for you."},
			{
				Type:  "tool_use",
				ID:    "toolu_456",
				Name:  "bash",
				Input: json.RawMessage(`{"command":"echo hi"}`),
			},
		},
		StopReason: a.StopReasonToolUse,
	}
	got := fromAnthropicMessage(msg)
	assert.Equal(t, "Let me run that for you.", got.Content)
	assert.Equal(t, "tool_use", got.StopReason)
	assert.Len(t, got.ToolCalls, 1)
}

func TestFromAnthropicMessage_InvalidToolInputFallsBackToEmptyObject(t *testing.T) {
	msg := &a.Message{
		Content: []a.ContentBlockUnion{
			{
				Type:  "tool_use",
				ID:    "toolu_789",
				Name:  "bash",
				Input: map[string]any{"bad": make(chan int)},
			},
		},
	}

	got := fromAnthropicMessage(msg)
	assert.Equal(t, "tool_use", got.StopReason)
	require.Len(t, got.ToolCalls, 1)
	assert.Equal(t, "toolu_789", got.ToolCalls[0].ID)
	assert.Equal(t, "bash", got.ToolCalls[0].Name)
	assert.Equal(t, "{}", got.ToolCalls[0].Arguments)
}
