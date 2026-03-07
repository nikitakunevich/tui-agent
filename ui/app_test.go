package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nikitakunevich/tui-agent/agent"
	"github.com/stretchr/testify/assert"
)

func TestModelInitialState(t *testing.T) {
	m := New(nil)
	assert.False(t, m.waiting)
	assert.Empty(t, m.messages)
	assert.True(t, m.showFireworks)
}

func TestModelViewBeforeReady(t *testing.T) {
	m := New(nil)
	assert.Contains(t, m.View(), "Initializing")
}

func TestFireworksDoneHidesFireworks(t *testing.T) {
	m := New(nil)
	m.showFireworks = true
	m.ready = true
	updated, _ := m.Update(fireworksDoneMsg{})
	model := updated.(*Model)
	assert.False(t, model.showFireworks)
}

func TestModelCtrlCQuits(t *testing.T) {
	m := New(nil)
	m.ready = true
	m.showFireworks = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	// cmd should be tea.Quit
	assert.NotNil(t, cmd)
}

func TestRenderMessagesEmpty(t *testing.T) {
	m := New(nil)
	result := m.renderMessages()
	assert.Contains(t, result, "No messages yet")
}

func TestRenderMessagesWithChat(t *testing.T) {
	m := New(nil)
	m.messages = []chatMessage{
		{role: "user", content: "hello"},
		{role: "assistant", content: "hi there"},
	}
	result := m.renderMessages()
	assert.Contains(t, result, "hello")
	assert.Contains(t, result, "hi there")
}

func TestAgentResponseMsgUpdatesState(t *testing.T) {
	m := New(nil)
	m.ready = true
	m.showFireworks = false
	m.waiting = true
	m.viewport.Width = 80
	m.viewport.Height = 20

	updated, _ := m.Update(agentResponseMsg{content: "test response"})
	model := updated.(*Model)
	assert.False(t, model.waiting)
	assert.Len(t, model.messages, 1)
	assert.Equal(t, "assistant", model.messages[0].role)
	assert.Equal(t, "test response", model.messages[0].content)
}

func TestAgentResponseMsgError(t *testing.T) {
	m := New(nil)
	m.ready = true
	m.showFireworks = false
	m.waiting = true
	m.viewport.Width = 80
	m.viewport.Height = 20

	updated, _ := m.Update(agentResponseMsg{err: assert.AnError})
	model := updated.(*Model)
	assert.False(t, model.waiting)
	assert.Len(t, model.messages, 1)
	assert.Equal(t, "error", model.messages[0].role)
}

func TestFormatToolEventStart(t *testing.T) {
	result := formatToolEvent(agent.ToolEvent{
		Type:      "start",
		Name:      "bash",
		Arguments: `{"command":"ls -la"}`,
	})
	assert.Contains(t, result, "▶")
	assert.Contains(t, result, "bash")
	assert.Contains(t, result, `{"command":"ls -la"}`)
}

func TestFormatToolEventStartTruncatesLongArgs(t *testing.T) {
	longArgs := strings.Repeat("x", 250)
	result := formatToolEvent(agent.ToolEvent{
		Type:      "start",
		Name:      "bash",
		Arguments: longArgs,
	})
	assert.Contains(t, result, "...")
	// The displayed args should be at most 200 chars + "..."
	assert.NotContains(t, result, strings.Repeat("x", 201))
}

func TestFormatToolEventEnd(t *testing.T) {
	result := formatToolEvent(agent.ToolEvent{
		Type:   "end",
		Name:   "bash",
		Result: "file1.go\nfile2.go\n",
	})
	assert.Contains(t, result, "✓")
	assert.Contains(t, result, "bash")
	assert.Contains(t, result, "18 chars")
}

func TestFormatToolEventEndLongResult(t *testing.T) {
	longResult := strings.Repeat("a", 500)
	result := formatToolEvent(agent.ToolEvent{
		Type:   "end",
		Name:   "bash",
		Result: longResult,
	})
	assert.Contains(t, result, "500 chars")
}

func TestFormatToolEventUnknownType(t *testing.T) {
	result := formatToolEvent(agent.ToolEvent{
		Type: "unknown",
		Name: "bash",
	})
	assert.Contains(t, result, "?")
	assert.Contains(t, result, "bash")
}

func TestToolEventMsgAppendsMessage(t *testing.T) {
	m := New(nil)
	m.ready = true
	m.showFireworks = false
	m.waiting = true
	m.viewport.Width = 80
	m.viewport.Height = 20

	updated, _ := m.Update(toolEventMsg{
		Type:      "start",
		Name:      "bash",
		Arguments: `{"command":"ls"}`,
	})
	model := updated.(*Model)
	assert.Len(t, model.messages, 1)
	assert.Equal(t, "tool", model.messages[0].role)
	assert.Contains(t, model.messages[0].content, "bash")
	assert.Contains(t, model.messages[0].content, "▶")
	// waiting should still be true (tool event doesn't clear it)
	assert.True(t, model.waiting)
}

func TestToolEventMsgEndAppendsMessage(t *testing.T) {
	m := New(nil)
	m.ready = true
	m.showFireworks = false
	m.waiting = true
	m.viewport.Width = 80
	m.viewport.Height = 20

	updated, _ := m.Update(toolEventMsg{
		Type:   "end",
		Name:   "bash",
		Result: "output",
	})
	model := updated.(*Model)
	assert.Len(t, model.messages, 1)
	assert.Equal(t, "tool", model.messages[0].role)
	assert.Contains(t, model.messages[0].content, "✓")
}

func TestRenderMessagesWithToolMessages(t *testing.T) {
	m := New(nil)
	m.viewport.Width = 80
	m.viewport.Height = 20
	m.messages = []chatMessage{
		{role: "user", content: "run ls"},
		{role: "tool", content: "  ▶ bash: ls"},
		{role: "tool", content: "  ✓ bash done (20 chars)"},
		{role: "assistant", content: "here are your files"},
	}
	result := m.renderMessages()
	assert.Contains(t, result, "run ls")
	assert.Contains(t, result, "▶ bash")
	assert.Contains(t, result, "✓ bash")
	assert.Contains(t, result, "here are your files")
}

func TestToolMessagesGetSingleLineSpacing(t *testing.T) {
	m := New(nil)
	m.viewport.Width = 80
	m.viewport.Height = 20
	m.messages = []chatMessage{
		{role: "tool", content: "  ▶ bash: ls"},
		{role: "tool", content: "  ✓ bash done"},
	}
	result := m.renderMessages()
	// Tool messages should NOT have double newlines between them
	assert.NotContains(t, result, "bash: ls\n\n")
}

func TestRenderMessagesLongLine(t *testing.T) {
	m := New(nil)
	m.viewport.Width = 40
	m.viewport.Height = 20
	longContent := strings.Repeat("word ", 50) // 250 chars
	m.messages = []chatMessage{
		{role: "user", content: longContent},
	}
	result := m.renderMessages()
	// Wrapped output should contain newlines within the message
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 2, "long message should be wrapped into multiple lines")
}

func TestRenderMessagesLongLinePreservesContent(t *testing.T) {
	m := New(nil)
	m.viewport.Width = 40
	m.viewport.Height = 20
	longContent := strings.Repeat("hello ", 40) // 240 chars
	m.messages = []chatMessage{
		{role: "assistant", content: longContent},
	}
	result := m.renderMessages()
	// All words should be present (no truncation)
	assert.Equal(t, 40, strings.Count(result, "hello"), "all words should be preserved after wrapping")
}
