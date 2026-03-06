package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nikitakunevich/tui-agent/agent"
)

// Styles
var (
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DC4E4")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6DA95")).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ED8796")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#6E738D"))
	titleStyle     = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#1E1E2E")).
			Bold(true).
			Padding(0, 1)
)

type chatMessage struct {
	role    string
	content string
}

// agentResponseMsg is sent when the agent finishes processing.
type agentResponseMsg struct {
	content string
	err     error
}

// Model is the Bubble Tea model for the TUI.
type Model struct {
	agent     *agent.Agent
	messages  []chatMessage
	input     textinput.Model
	viewport  viewport.Model
	spinner   spinner.Model
	fireworks fireworksModel
	showFireworks bool
	waiting   bool
	width     int
	height    int
	ready     bool
	err       error
}

// New creates a new TUI model.
func New(a *agent.Agent) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 4096

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6DA95"))

	return Model{
		agent:         a,
		input:         ti,
		spinner:       sp,
		showFireworks: true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		fireworksTickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.waiting || m.input.Value() == "" {
				break
			}
			userMsg := m.input.Value()
			m.messages = append(m.messages, chatMessage{role: "user", content: userMsg})
			m.input.Reset()
			m.waiting = true
			m.updateViewport()
			cmds = append(cmds, m.sendMessage(userMsg))
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 1
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4) // 4 lines for input + status
			m.viewport.SetContent(m.renderMessages())
			m.fireworks = newFireworks(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
			m.updateViewport()
		}

	case fireworksTickMsg:
		if m.showFireworks {
			cmd := m.fireworks.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case fireworksDoneMsg:
		m.showFireworks = false

	case agentResponseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{role: "error", content: msg.err.Error()})
		} else {
			m.messages = append(m.messages, chatMessage{role: "assistant", content: msg.content})
		}
		m.updateViewport()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update text input
	var tiCmd tea.Cmd
	m.input, tiCmd = m.input.Update(msg)
	cmds = append(cmds, tiCmd)

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.showFireworks {
		return m.fireworks.view(m.width, m.height)
	}

	var b strings.Builder

	// Title bar
	title := titleStyle.Render(" TUI Agent ")
	b.WriteString(title)
	b.WriteString("\n")

	// Messages viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Status line
	if m.waiting {
		b.WriteString(m.spinner.View() + " Thinking...")
	} else {
		b.WriteString(dimStyle.Render("Press Enter to send, Esc to quit"))
	}
	b.WriteString("\n")

	// Input
	b.WriteString(m.input.View())

	return b.String()
}

func (m *Model) sendMessage(msg string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.agent.Run(context.Background(), msg)
		return agentResponseMsg{content: result, err: err}
	}
}

func (m *Model) updateViewport() {
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

func (m *Model) renderMessages() string {
	if len(m.messages) == 0 {
		return dimStyle.Render("  No messages yet. Start chatting!")
	}

	wrapStyle := lipgloss.NewStyle().Width(m.viewport.Width)

	var b strings.Builder
	for _, msg := range m.messages {
		var line string
		switch msg.role {
		case "user":
			line = userStyle.Render("You: ") + msg.content
		case "assistant":
			line = assistantStyle.Render("Agent: ") + msg.content
		case "error":
			line = errorStyle.Render(fmt.Sprintf("Error: %s", msg.content))
		}
		b.WriteString(wrapStyle.Render(line))
		b.WriteString("\n\n")
	}
	return b.String()
}
