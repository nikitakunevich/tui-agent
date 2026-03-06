# TUI Agent — Project Conventions

## Build & Test
- `go build .` — build binary
- `go test ./...` — run all tests
- `go test ./agent/...` — run agent tests only
- `go vet ./...` — lint

## Code Style
- Go standard formatting (`gofmt`)
- Error handling: return errors, don't panic
- Use `log/slog` for all logging (never `fmt.Print` in library code)
- Tests: use `testify/assert`, table-driven where appropriate

## Directory Structure
```
tui-agent/
├── main.go                        # entrypoint, flags, wiring
├── go.mod / go.sum
├── .env                           # API keys (gitignored)
├── agent/
│   ├── agent.go                   # core tool-use loop
│   └── agent_test.go
├── llm/
│   ├── types.go                   # Message, ToolCall, ToolDef, ChatRequest/Response
│   ├── provider.go                # Provider interface
│   ├── openai/openai.go           # OpenAI-compatible provider
│   └── anthropic/anthropic.go     # Anthropic native provider
├── tools/
│   ├── registry.go                # tool registry
│   ├── bash.go                    # bash exec tool
│   └── *_test.go
├── ui/
│   ├── app.go                     # Bubble Tea model (chat UI)
│   ├── fireworks.go               # startup fireworks animation
│   └── app_test.go
├── logging/
│   └── logging.go                 # slog file + stderr setup
└── docs/                          # module documentation (*.md)
```

## Architecture
- `llm/` — provider abstraction (interface + implementations). See docs/llm.md
- `agent/` — core agent loop. See docs/agent.md
- `tools/` — tool registry + tool implementations. See docs/tools.md
- `ui/` — Bubble Tea TUI with fireworks startup. See docs/ui.md
- `logging/` — slog setup

## Key Interfaces
- `llm.Provider` — single `Chat(ctx, ChatRequest) (*ChatResponse, error)` method
- `tools.Tool` — `Name()`, `Description()`, `Parameters()`, `Execute(ctx, input)`
- Agent loop: send → check tool_calls → execute tools → repeat until text response

## Conventions
- Module: `github.com/nikitakunevich/tui-agent`
- No `pkg/` or `internal/` — flat top-level packages
- Provider interface returns concrete types, not interfaces
- Tools implement `tools.Tool` interface
- JSON Schema for tool parameters as `json.RawMessage`
- .env file loaded via godotenv for API keys
- Config via flags: `--provider`, `--model`, `--debug`, `--base-url`
- Config via env: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_BASE_URL`, `MODEL_NAME`
- Module docs in `docs/*.md` — update when adding/changing modules

## Workflow: Task Artifacts
When working on a big feature that requires fetching docs, API references, or other artifacts:
1. Save them in `.agent/tasks-artifacts/<task_name>/` as temp files
2. Use these to resume work if interrupted
3. Delete the artifacts directory after the feature is validated and complete
4. `.agent/` is gitignored
