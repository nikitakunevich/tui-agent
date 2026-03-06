package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBashToolName(t *testing.T) {
	b := NewBashTool()
	assert.Equal(t, "bash", b.Name())
}

func TestBashToolEcho(t *testing.T) {
	b := NewBashTool()
	result, err := b.Execute(context.Background(), json.RawMessage(`{"command":"echo hello"}`))
	require.NoError(t, err)
	assert.Equal(t, "hello\n", result)
}

func TestBashToolStderr(t *testing.T) {
	b := NewBashTool()
	result, err := b.Execute(context.Background(), json.RawMessage(`{"command":"echo error >&2"}`))
	require.NoError(t, err)
	assert.Contains(t, result, "error")
}

func TestBashToolNonZeroExit(t *testing.T) {
	b := NewBashTool()
	result, err := b.Execute(context.Background(), json.RawMessage(`{"command":"exit 1"}`))
	require.NoError(t, err) // non-zero exit is not a Go error
	assert.Contains(t, result, "exit status")
}

func TestBashToolTimeout(t *testing.T) {
	b := NewBashTool()
	b.Timeout = 100 * time.Millisecond
	_, err := b.Execute(context.Background(), json.RawMessage(`{"command":"sleep 10"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestBashToolInvalidInput(t *testing.T) {
	b := NewBashTool()
	_, err := b.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestBashToolEmptyCommand(t *testing.T) {
	b := NewBashTool()
	_, err := b.Execute(context.Background(), json.RawMessage(`{"command":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command is required")
}

func TestBashToolParameters(t *testing.T) {
	b := NewBashTool()
	params := b.Parameters()
	var schema map[string]any
	err := json.Unmarshal(params, &schema)
	require.NoError(t, err)
	assert.Equal(t, "object", schema["type"])
}
