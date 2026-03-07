# Tasks

## Done
- [x] Project scaffold (git, go.mod, CLAUDE.md, logging, main.go)
- [x] LLM provider abstraction (types, interface, OpenAI, Anthropic)
- [x] Tool system (registry, bash tool)
- [x] Agent loop (core while-true loop)
- [x] TUI (Bubble Tea + fireworks + main wiring)
- [x] .env file support for API keys
- [x] Module documentation in docs/

## TODO

Theme: "works" → "trustworthy and inspectable"

### P1: Tool confirmation UI
- [ ] Before running bash, show command and ask approve/deny
- Highest leverage for trust and safety

### P2: Show tool activity in chat
- [ ] "Running bash: ls -la" messages in chat
- [ ] Collapsible or dimmed tool output

### P3: Refactor agent state/session model
- [ ] Avoid hidden mutable conversation state in `Agent`

### P4: Streaming
- [ ] Streaming responses

### P5: Persistence
- [ ] Save/load chat sessions

### Minor Polish
- [ ] Sort tool names in `Registry.List()` / `ToToolDefs()` for deterministic behavior
- [ ] Show current provider/model in TUI title/status
- [ ] Add `/clear` or `/reset` command
- [ ] Add file read tool before file write tool
- [ ] Consider limiting bash environment / cwd explicitly
- [ ] More tools (HTTP, etc.)
