# ui/ — Terminal User Interface

## Bubble Tea Model (ui/app.go)
Chat interface with text input at bottom, messages scrolling above.

### Components
- `viewport` — scrollable message display (bubbles/viewport)
- `textinput` — user input (bubbles/textinput)
- `spinner` — shown while waiting for LLM response (bubbles/spinner)

### Message types
- `agentResponseMsg` — result from agent.Run() goroutine
- `fireworksTickMsg` / `fireworksDoneMsg` — animation control

### Keybindings
- Enter — send message
- Ctrl+C / Esc — quit

## Fireworks (ui/fireworks.go)
Particle-based fireworks animation shown on startup for ~2 seconds.
- Multiple rocket bursts with colored particles
- Gravity simulation
- Title display ("TUI Agent") centered on screen
- Transitions to chat UI after animation completes

## Styling
Uses lipgloss for colors. User messages in blue, assistant in green, errors in red.
