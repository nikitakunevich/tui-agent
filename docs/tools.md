# tools/ — Tool System

## Tool Interface (tools/registry.go)
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() json.RawMessage  // JSON Schema
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```

## Registry
- `NewRegistry()` — create empty registry
- `Register(Tool)` — add tool
- `Get(name)` — lookup by name
- `List()` — all tool names
- `ToToolDefs()` — convert to `[]llm.ToolDef` for LLM requests
- `Execute(ctx, name, input)` — run tool, return result string (errors wrapped in string)

## Bash Tool (tools/bash.go)
- Runs commands via `os/exec` with `bash -c`
- Configurable timeout (default 30s)
- Configurable working directory
- Output truncated at 50KB
- Non-zero exit codes returned as results (not Go errors)

## Adding a new tool
1. Create `tools/newtool.go`
2. Implement `tools.Tool` interface
3. Register in `main.go`: `registry.Register(tools.NewMyTool())`
