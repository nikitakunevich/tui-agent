package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nikitakunevich/tui-agent/llm"
	openai "github.com/sashabaranov/go-openai"
)

// Config holds settings for the OpenAI-compatible provider.
type Config struct {
	APIKey  string
	BaseURL string // optional, for Ollama/other compatible APIs
	Model   string
}

// Provider implements llm.Provider for OpenAI-compatible APIs.
type Provider struct {
	client *openai.Client
	model  string
}

// New creates a new OpenAI-compatible provider.
func New(cfg Config) *Provider {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}
	return &Provider{
		client: openai.NewClientWithConfig(config),
		model:  cfg.Model,
	}
}

func (p *Provider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	msgs := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = toOpenAIMessage(m)
	}

	tools := make([]openai.Tool, len(req.Tools))
	for i, t := range req.Tools {
		tools[i] = toOpenAITool(t)
	}

	oaiReq := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
	}
	if len(tools) > 0 {
		oaiReq.Tools = tools
	}

	slog.Debug("openai request", "model", p.model, "messages", len(msgs), "tools", len(tools))
	start := time.Now()

	resp, err := p.client.CreateChatCompletion(ctx, oaiReq)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}

	slog.Debug("openai response",
		"duration", time.Since(start),
		"choices", len(resp.Choices),
		"usage_prompt", resp.Usage.PromptTokens,
		"usage_completion", resp.Usage.CompletionTokens,
	)

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}

	choice := resp.Choices[0]
	return fromOpenAIChoice(choice), nil
}

func toOpenAIMessage(m llm.Message) openai.ChatCompletionMessage {
	msg := openai.ChatCompletionMessage{
		Role:    m.Role,
		Content: m.Content,
	}
	if m.ToolCallID != "" {
		msg.ToolCallID = m.ToolCallID
	}
	if len(m.ToolCalls) > 0 {
		msg.ToolCalls = make([]openai.ToolCall, len(m.ToolCalls))
		for i, tc := range m.ToolCalls {
			msg.ToolCalls[i] = openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			}
		}
	}
	// Work around go-openai's omitempty on Content: when Content is ""
	// the field is omitted from JSON, causing OpenAI to receive null
	// and reject with 400 "expected a string, got null". Using
	// MultiContent forces the library to emit a non-null content field.
	if msg.Content == "" {
		msg.MultiContent = []openai.ChatMessagePart{
			{Type: openai.ChatMessagePartTypeText, Text: ""},
		}
	}
	return msg
}

func toOpenAITool(t llm.ToolDef) openai.Tool {
	var params json.RawMessage
	if t.Parameters != nil {
		params = t.Parameters
	} else {
		params = json.RawMessage(`{}`)
	}
	def := openai.FunctionDefinition{
		Name:        t.Name,
		Description: t.Description,
		Parameters:  params,
	}
	return openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &def,
	}
}

func fromOpenAIChoice(c openai.ChatCompletionChoice) *llm.ChatResponse {
	resp := &llm.ChatResponse{
		Content: c.Message.Content,
	}

	if len(c.Message.ToolCalls) > 0 {
		resp.StopReason = "tool_use"
		resp.ToolCalls = make([]llm.ToolCall, len(c.Message.ToolCalls))
		for i, tc := range c.Message.ToolCalls {
			resp.ToolCalls[i] = llm.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	} else {
		resp.StopReason = "end_turn"
	}

	return resp
}
