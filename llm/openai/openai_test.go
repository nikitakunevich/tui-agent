package openai

import (
	"encoding/json"
	"testing"

	"github.com/nikitakunevich/tui-agent/llm"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestToOpenAIMessage_User(t *testing.T) {
	msg := llm.Message{Role: llm.RoleUser, Content: "hello"}
	got := toOpenAIMessage(msg)
	assert.Equal(t, "user", got.Role)
	assert.Equal(t, "hello", got.Content)
}

func TestToOpenAIMessage_ToolResult(t *testing.T) {
	msg := llm.Message{Role: llm.RoleTool, Content: "output", ToolCallID: "call_123"}
	got := toOpenAIMessage(msg)
	assert.Equal(t, "tool", got.Role)
	assert.Equal(t, "call_123", got.ToolCallID)
	assert.Equal(t, "output", got.Content)
}

func TestToOpenAIMessage_AssistantWithToolCalls(t *testing.T) {
	msg := llm.Message{
		Role: llm.RoleAssistant,
		ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "bash", Arguments: `{"command":"ls"}`},
		},
	}
	got := toOpenAIMessage(msg)
	assert.Len(t, got.ToolCalls, 1)
	assert.Equal(t, "call_1", got.ToolCalls[0].ID)
	assert.Equal(t, "bash", got.ToolCalls[0].Function.Name)
	assert.Equal(t, `{"command":"ls"}`, got.ToolCalls[0].Function.Arguments)
}

func TestToOpenAITool(t *testing.T) {
	td := llm.ToolDef{
		Name:        "bash",
		Description: "Execute a bash command",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"}}}`),
	}
	got := toOpenAITool(td)
	assert.Equal(t, openai.ToolTypeFunction, got.Type)
	assert.Equal(t, "bash", got.Function.Name)
	assert.Equal(t, "Execute a bash command", got.Function.Description)
}

func TestToOpenAIMessage_AssistantToolCallsSerialization(t *testing.T) {
	// When the assistant responds with tool calls and no text content,
	// Content is "". The library's omitempty tag causes it to be omitted
	// from JSON, resulting in "content": null which OpenAI rejects.
	msg := llm.Message{
		Role: llm.RoleAssistant,
		ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "bash", Arguments: `{"command":"ls"}`},
		},
	}
	got := toOpenAIMessage(msg)
	b, err := json.Marshal(got)
	assert.NoError(t, err)
	// The serialized JSON must contain a "content" key (not omitted)
	var raw map[string]json.RawMessage
	err = json.Unmarshal(b, &raw)
	assert.NoError(t, err)
	contentVal, hasContent := raw["content"]
	assert.True(t, hasContent, "assistant message with tool_calls must have 'content' field in JSON")
	// And it must not be null
	assert.NotEqual(t, "null", string(contentVal), "content must not be null")
	t.Logf("serialized JSON: %s", string(b))
}


func TestFromOpenAIChoice_TextResponse(t *testing.T) {
	choice := openai.ChatCompletionChoice{
		Message: openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: "Hello!",
		},
	}
	got := fromOpenAIChoice(choice)
	assert.Equal(t, "Hello!", got.Content)
	assert.Equal(t, "end_turn", got.StopReason)
	assert.Empty(t, got.ToolCalls)
}

func TestFromOpenAIChoice_ToolCallResponse(t *testing.T) {
	choice := openai.ChatCompletionChoice{
		Message: openai.ChatCompletionMessage{
			Role: "assistant",
			ToolCalls: []openai.ToolCall{
				{
					ID:   "call_abc",
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      "bash",
						Arguments: `{"command":"echo hi"}`,
					},
				},
			},
		},
	}
	got := fromOpenAIChoice(choice)
	assert.Equal(t, "tool_use", got.StopReason)
	assert.Len(t, got.ToolCalls, 1)
	assert.Equal(t, "call_abc", got.ToolCalls[0].ID)
	assert.Equal(t, "bash", got.ToolCalls[0].Name)
	assert.Equal(t, `{"command":"echo hi"}`, got.ToolCalls[0].Arguments)
}
