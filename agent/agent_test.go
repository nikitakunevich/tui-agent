package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nikitakunevich/tui-agent/llm"
	"github.com/nikitakunevich/tui-agent/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider returns scripted responses in order.
type mockProvider struct {
	responses []*llm.ChatResponse
	calls     int
}

func (m *mockProvider) Chat(_ context.Context, _ llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.calls >= len(m.responses) {
		return &llm.ChatResponse{Content: "no more responses", StopReason: "end_turn"}, nil
	}
	resp := m.responses[m.calls]
	m.calls++
	return resp, nil
}

// echoTool returns whatever it receives as input.
type echoTool struct{}

func (e *echoTool) Name() string               { return "echo" }
func (e *echoTool) Description() string         { return "echo tool" }
func (e *echoTool) Parameters() json.RawMessage { return json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`) }
func (e *echoTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct{ Text string `json:"text"` }
	json.Unmarshal(input, &params)
	return params.Text, nil
}

func TestAgentSingleTextResponse(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{Content: "Hello!", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	agent := New(provider, registry, "")

	result, err := agent.Run(context.Background(), "hi")
	require.NoError(t, err)
	assert.Equal(t, "Hello!", result)
	assert.Equal(t, 1, provider.calls)
}

func TestAgentToolCallThenText(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{
				StopReason: "tool_use",
				ToolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "echo", Arguments: `{"text":"world"}`},
				},
			},
			{Content: "The result is: world", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	agent := New(provider, registry, "")

	result, err := agent.Run(context.Background(), "echo world")
	require.NoError(t, err)
	assert.Equal(t, "The result is: world", result)
	assert.Equal(t, 2, provider.calls)
}

func TestAgentMultipleToolCalls(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{
				StopReason: "tool_use",
				ToolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "echo", Arguments: `{"text":"a"}`},
					{ID: "call_2", Name: "echo", Arguments: `{"text":"b"}`},
				},
			},
			{Content: "done", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	agent := New(provider, registry, "")

	result, err := agent.Run(context.Background(), "multi")
	require.NoError(t, err)
	assert.Equal(t, "done", result)
	assert.Equal(t, 2, provider.calls)
}

func TestAgentMaxIterations(t *testing.T) {
	// Provider always returns tool calls — should hit max iterations
	provider := &mockProvider{
		responses: make([]*llm.ChatResponse, 20),
	}
	for i := range provider.responses {
		provider.responses[i] = &llm.ChatResponse{
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{ID: "call", Name: "echo", Arguments: `{"text":"loop"}`},
			},
		}
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	agent := New(provider, registry, "")
	agent.SetMaxIterations(3)

	_, err := agent.Run(context.Background(), "infinite")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations")
	assert.Equal(t, 3, provider.calls)
}

func TestOnToolEventCallback(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{
				StopReason: "tool_use",
				ToolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "echo", Arguments: `{"text":"hello"}`},
				},
			},
			{Content: "done", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	a := New(provider, registry, "")

	var events []ToolEvent
	a.OnToolEvent = func(e ToolEvent) {
		events = append(events, e)
	}

	_, err := a.Run(context.Background(), "test")
	require.NoError(t, err)

	require.Len(t, events, 2)
	// start event
	assert.Equal(t, "start", events[0].Type)
	assert.Equal(t, "echo", events[0].Name)
	assert.Equal(t, `{"text":"hello"}`, events[0].Arguments)
	assert.Empty(t, events[0].Result)
	// end event
	assert.Equal(t, "end", events[1].Type)
	assert.Equal(t, "echo", events[1].Name)
	assert.Equal(t, `{"text":"hello"}`, events[1].Arguments)
	assert.Equal(t, "hello", events[1].Result)
}

func TestOnToolEventMultipleTools(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{
				StopReason: "tool_use",
				ToolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "echo", Arguments: `{"text":"a"}`},
					{ID: "call_2", Name: "echo", Arguments: `{"text":"b"}`},
				},
			},
			{Content: "done", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	a := New(provider, registry, "")

	var events []ToolEvent
	a.OnToolEvent = func(e ToolEvent) {
		events = append(events, e)
	}

	_, err := a.Run(context.Background(), "test")
	require.NoError(t, err)

	require.Len(t, events, 4) // start+end for each tool call
	assert.Equal(t, "start", events[0].Type)
	assert.Equal(t, "end", events[1].Type)
	assert.Equal(t, "a", events[1].Result)
	assert.Equal(t, "start", events[2].Type)
	assert.Equal(t, "end", events[3].Type)
	assert.Equal(t, "b", events[3].Result)
}

func TestOnToolEventNilCallbackSafe(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{
				StopReason: "tool_use",
				ToolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "echo", Arguments: `{"text":"x"}`},
				},
			},
			{Content: "done", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	registry.Register(&echoTool{})
	a := New(provider, registry, "")
	// OnToolEvent is nil — should not panic
	result, err := a.Run(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "done", result)
}

func TestAgentWithSystemPrompt(t *testing.T) {
	provider := &mockProvider{
		responses: []*llm.ChatResponse{
			{Content: "I am helpful", StopReason: "end_turn"},
		},
	}
	registry := tools.NewRegistry()
	agent := New(provider, registry, "You are a helpful assistant")

	result, err := agent.Run(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, "I am helpful", result)
}
