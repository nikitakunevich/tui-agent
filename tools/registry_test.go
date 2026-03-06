package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTool struct {
	name   string
	result string
	err    error
}

func (m *mockTool) Name() string                  { return m.name }
func (m *mockTool) Description() string            { return "mock tool" }
func (m *mockTool) Parameters() json.RawMessage    { return json.RawMessage(`{}`) }
func (m *mockTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return m.result, m.err
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "test"}
	r.Register(tool)

	got, ok := r.Get("test")
	require.True(t, ok)
	assert.Equal(t, "test", got.Name())
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "a"})
	r.Register(&mockTool{name: "b"})
	names := r.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
}

func TestRegistryToToolDefs(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "bash"})
	defs := r.ToToolDefs()
	require.Len(t, defs, 1)
	assert.Equal(t, "bash", defs[0].Name)
}

func TestRegistryExecute(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "test", result: "hello"})

	result := r.Execute(context.Background(), "test", json.RawMessage(`{}`))
	assert.Equal(t, "hello", result)
}

func TestRegistryExecuteNotFound(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(context.Background(), "missing", json.RawMessage(`{}`))
	assert.Contains(t, result, "tool not found")
}
