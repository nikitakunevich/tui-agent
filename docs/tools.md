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
- `Execute(ctx, name, input)` — run tool, return `(string, error)`; not-found and tool errors returned as `error`

## Helpers (tools/schema.go)
- `ParseInput[T](input, validate)` — generic JSON unmarshal + optional validation; eliminates repeated unmarshal+validate in every tool
- `ObjectSchema(props, required...)` — builds a JSON Schema `"object"` from a `map[string]Property`; replaces hard-coded JSON strings in tool `Parameters()` methods

## Bash Tool (tools/bash.go)
- Runs commands via `os/exec` with `bash -c`
- Configurable timeout (default 30s)
- Configurable working directory
- Output truncated at 50KB
- Non-zero exit codes returned as results (not Go errors)

## Read File Tool (tools/read_file.go)
- Reads file contents with optional offset (1-based line) and limit
- Output truncated at 50KB

## Adding a new tool
1. Create `tools/newtool.go`
2. Implement `tools.Tool` interface — use `ParseInput` for input parsing and `ObjectSchema` for parameters
3. Register in `main.go`: `registry.Register(tools.NewMyTool())`
