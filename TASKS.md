# Tasks

## Done
- [x] Project scaffold (git, go.mod, CLAUDE.md, logging, main.go)
- [x] LLM provider abstraction (types, interface, OpenAI, Anthropic)
- [x] Tool system (registry, bash tool)
- [x] Agent loop (core while-true loop)
- [x] TUI (Bubble Tea + fireworks + main wiring)
- [x] .env file support for API keys
- [x] Module documentation in docs/
- [x] Show tool activity in chat (dimmed ▶/✓ messages via callback + p.Send)
- [x] Add read_file tool
- [x] Fix empty content null serialization for all OpenAI message types

## TODO

Theme: "works" → "trustworthy and inspectable"

### P1: Tool confirmation UI
- [ ] Before running bash, show command and ask approve/deny
- Highest leverage for trust and safety

### P3: Refactor agent state/session model
- [ ] Avoid hidden mutable conversation state in `Agent`

### P4: Streaming
- [ ] Streaming responses

### P5: Persistence
- [ ] Save/load chat sessions

### Agentic Automations (save dev time, ship faster)

1. **Tool Skeleton Generator** — CLI command generates tool.go + test + JSON schema + registry wiring from natural language description
   - [ ] Implement `autogen tool "description"` scaffolding command
   - Trigger: CLI command | Complexity: Low

2. **AI Code Review on PR** — Claude reviews PRs for Go idioms, error handling, concurrency bugs; leaves inline comments
   - [ ] Add GitHub Action workflow using Claude API for PR review
   - Trigger: PR opened | Complexity: Medium

3. **Conversation Replay Regression Testing** — Record golden conversations, replay as regression tests to catch behavioral changes
   - [ ] Build test harness that replays recorded agent sessions
   - Trigger: CI on commit | Complexity: Medium

4. **Dogfood Agent** — Use tui-agent itself to implement features from TASKS.md (agent develops itself)
   - [ ] Create meta-workflow invoking agent.Run() on own codebase tasks
   - Trigger: Manual/scheduled | Complexity: High

5. **CLAUDE.md Drift Detector** — Detect code divergence from CLAUDE.md (new tools, changed dirs), auto-update
   - [ ] Add weekly GHA or pre-commit check for doc/code drift
   - Trigger: Weekly cron | Complexity: Low

6. **Auto-Fix Lint + Format CI** — Run golangci-lint, gofmt on PRs, auto-commit fixes
   - [ ] Add lint/format GitHub Action with auto-fix commits
   - Trigger: PR opened | Complexity: Low

7. **PR Description Generator** — Generate PR descriptions from diffs + commit messages using Claude
   - [ ] Add GHA that auto-generates PR body on open
   - Trigger: PR opened | Complexity: Low

8. **Tool Hallucination Detector** — Monitor agent execution, log when LLM invents non-existent tools/params
   - [ ] Add validation in agent loop + structured logging of hallucinations
   - Trigger: Runtime (OnToolEvent) | Complexity: Low

9. **Auto-Update TASKS.md from PRs** — Watch merged PRs, mark matching TASKS.md items as done
   - [ ] Add post-merge GHA that parses commits and updates checkboxes
   - Trigger: Post-merge | Complexity: Low

10. **Agent Behavior Monitoring** — Track success rates, tool usage frequency, iteration depth, LLM latency per session
    - [ ] Add metrics collection in agent loop + summary output
    - Trigger: Continuous | Complexity: Medium

### Minor Polish
- [ ] Sort tool names in `Registry.List()` / `ToToolDefs()` for deterministic behavior
- [ ] Show current provider/model in TUI title/status
- [ ] Add `/clear` or `/reset` command
- [x] Add file read tool before file write tool
- [ ] Consider limiting bash environment / cwd explicitly
- [ ] More tools (HTTP, etc.)
