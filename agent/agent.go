package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nikitakunevich/tui-agent/llm"
	"github.com/nikitakunevich/tui-agent/tools"
)

const defaultMaxIterations = 10

// ToolEvent represents a tool execution event (start or end).
type ToolEvent struct {
	Type      string // "start" or "end"
	Name      string
	Arguments string
	Result    string
}

// Agent orchestrates the LLM tool-use loop.
type Agent struct {
	provider      llm.Provider
	registry      *tools.Registry
	messages      []llm.Message
	maxIterations int
	OnToolEvent   func(ToolEvent)
}

// New creates a new Agent.
func New(provider llm.Provider, registry *tools.Registry, systemPrompt string) *Agent {
	var msgs []llm.Message
	if systemPrompt != "" {
		msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: systemPrompt})
	}
	return &Agent{
		provider:      provider,
		registry:      registry,
		messages:      msgs,
		maxIterations: defaultMaxIterations,
	}
}

// SetMaxIterations overrides the default max iterations.
func (a *Agent) SetMaxIterations(n int) {
	a.maxIterations = n
}

// Run sends a user message and runs the tool-use loop until the LLM returns text.
func (a *Agent) Run(ctx context.Context, userMessage string) (string, error) {
	a.messages = append(a.messages, llm.Message{Role: llm.RoleUser, Content: userMessage})

	for i := 0; i < a.maxIterations; i++ {
		slog.Debug("agent loop iteration", "iteration", i+1)

		resp, err := a.provider.Chat(ctx, llm.ChatRequest{
			Messages: a.messages,
			Tools:    a.registry.ToToolDefs(),
		})
		if err != nil {
			return "", fmt.Errorf("agent chat: %w", err)
		}

		// Append assistant message to history
		assistantMsg := llm.Message{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		a.messages = append(a.messages, assistantMsg)

		// No tool calls — we're done
		if len(resp.ToolCalls) == 0 {
			slog.Debug("agent done", "iterations", i+1)
			return resp.Content, nil
		}

		// Execute each tool call and append results
		for _, tc := range resp.ToolCalls {
			slog.Info("executing tool", "name", tc.Name, "id", tc.ID)
			if a.OnToolEvent != nil {
				a.OnToolEvent(ToolEvent{Type: "start", Name: tc.Name, Arguments: tc.Arguments})
			}
			result := a.registry.Execute(ctx, tc.Name, json.RawMessage(tc.Arguments))
			if a.OnToolEvent != nil {
				a.OnToolEvent(ToolEvent{Type: "end", Name: tc.Name, Arguments: tc.Arguments, Result: result})
			}
			a.messages = append(a.messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", fmt.Errorf("agent exceeded max iterations (%d)", a.maxIterations)
}
