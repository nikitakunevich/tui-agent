package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/nikitakunevich/tui-agent/llm"
)

// Config holds settings for the Anthropic provider.
type Config struct {
	APIKey   string
	Model    string
	MaxTokens int64 // default 4096
}

// Provider implements llm.Provider using the Anthropic native API.
type Provider struct {
	client *anthropic.Client
	model  anthropic.Model
	maxTokens int64
}

// New creates a new Anthropic provider.
func New(cfg Config) *Provider {
	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)
	return &Provider{
		client:    &client,
		model:     anthropic.Model(cfg.Model),
		maxTokens: maxTokens,
	}
}

func (p *Provider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	// Separate system message from conversation messages.
	var system []anthropic.TextBlockParam
	var msgs []anthropic.MessageParam

	for _, m := range req.Messages {
		switch m.Role {
		case llm.RoleSystem:
			system = append(system, anthropic.TextBlockParam{Text: m.Content})
		case llm.RoleUser:
			if m.ToolCallID != "" {
				// Tool result
				msgs = append(msgs, anthropic.NewUserMessage(
					anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false),
				))
			} else {
				msgs = append(msgs, anthropic.NewUserMessage(
					anthropic.NewTextBlock(m.Content),
				))
			}
		case llm.RoleTool:
			// Tool results in Anthropic are sent as user messages
			msgs = append(msgs, anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false),
			))
		case llm.RoleAssistant:
			var blocks []anthropic.ContentBlockParamUnion
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tc := range m.ToolCalls {
				var input any
				if err := json.Unmarshal([]byte(tc.Arguments), &input); err != nil {
					slog.Warn("failed to unmarshal tool call arguments", "tool", tc.Name, "error", err)
					input = map[string]any{}
				}
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
			}
			if len(blocks) > 0 {
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			}
		}
	}

	// Build tools
	tools := make([]anthropic.ToolUnionParam, len(req.Tools))
	for i, t := range req.Tools {
		tools[i] = toAnthropicTool(t)
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: p.maxTokens,
		Messages:  msgs,
	}
	if len(system) > 0 {
		params.System = system
	}
	if len(tools) > 0 {
		params.Tools = tools
	}

	slog.Debug("anthropic request", "model", p.model, "messages", len(msgs), "tools", len(tools))
	start := time.Now()

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic chat: %w", err)
	}

	slog.Debug("anthropic response",
		"duration", time.Since(start),
		"stop_reason", resp.StopReason,
		"usage_input", resp.Usage.InputTokens,
		"usage_output", resp.Usage.OutputTokens,
	)

	return fromAnthropicMessage(resp), nil
}

func toAnthropicTool(t llm.ToolDef) anthropic.ToolUnionParam {
	var props any
	var required []string

	// Parse the JSON Schema to extract properties and required fields
	var schema map[string]any
	if err := json.Unmarshal(t.Parameters, &schema); err == nil {
		props = schema["properties"]
		if req, ok := schema["required"].([]any); ok {
			for _, r := range req {
				if s, ok := r.(string); ok {
					required = append(required, s)
				}
			}
		}
	}

	return anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        t.Name,
			Description: anthropic.String(t.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: props,
				Required:   required,
			},
		},
	}
}

func fromAnthropicMessage(msg *anthropic.Message) *llm.ChatResponse {
	resp := &llm.ChatResponse{}

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			resp.Content += block.Text
		case "tool_use":
			inputJSON, err := json.Marshal(block.Input)
			if err != nil {
				slog.Error("failed to marshal tool input", "tool", block.Name, "error", err)
				inputJSON = []byte("{}")
			}
			resp.ToolCalls = append(resp.ToolCalls, llm.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(inputJSON),
			})
		}
	}

	resp.StopReason = llm.StopReasonFromToolCalls(resp.ToolCalls)
	return resp
}
