package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
	model := updated.(Model)
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
	model := updated.(Model)
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
	model := updated.(Model)
	assert.False(t, model.waiting)
	assert.Len(t, model.messages, 1)
	assert.Equal(t, "error", model.messages[0].role)
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
