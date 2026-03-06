# llm/ — LLM Provider Abstraction

## Types (llm/types.go)
- `Message` — role, content, tool_calls, tool_call_id
- `ToolCall` — id, name, arguments (raw JSON string)
- `ToolDef` — name, description, parameters (JSON Schema as json.RawMessage)
- `ChatRequest` — messages + tools
- `ChatResponse` — content, tool_calls, stop_reason ("end_turn" | "tool_use")

## Provider Interface (llm/provider.go)
```go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
```

## Implementations
- `llm/openai/` — OpenAI-compatible (sashabaranov/go-openai). Supports custom base URL for Ollama etc.
- `llm/anthropic/` — Anthropic native (anthropics/anthropic-sdk-go). System messages extracted from conversation and passed separately.

## Adding a new provider
1. Create `llm/newprovider/newprovider.go`
2. Implement `llm.Provider` interface
3. Add case in `main.go` switch
