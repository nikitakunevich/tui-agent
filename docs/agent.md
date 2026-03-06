# agent/ — Core Agent Loop

## Agent (agent/agent.go)
The heart of the project. Orchestrates the LLM tool-use loop.

### Flow
1. Append user message to conversation history
2. Send history + tool defs to LLM provider
3. Append assistant response to history
4. If no tool_calls → return text content (done)
5. Execute each tool_call via registry, append results
6. Go to step 2

### Config
- `maxIterations` — default 10, prevents infinite loops
- `systemPrompt` — prepended as system message

### Constructor
```go
agent.New(provider llm.Provider, registry *tools.Registry, systemPrompt string) *Agent
```

### Main method
```go
func (a *Agent) Run(ctx context.Context, userMessage string) (string, error)
```

## Testing
Mock provider with scripted responses. No real API calls in tests.
