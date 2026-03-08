package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nikitakunevich/tui-agent/llm"
)

// Tool is the interface that all tools must implement.
type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage // JSON Schema
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

// Registry holds registered tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
	slog.Debug("registered tool", "name", t.Name())
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ToToolDefs converts all registered tools to LLM tool definitions.
func (r *Registry) ToToolDefs() []llm.ToolDef {
	defs := make([]llm.ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, llm.ToolDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		})
	}
	return defs
}

// Execute runs a tool by name with the given input JSON.
func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	slog.Debug("executing tool", "name", name, "input", string(input))
	result, err := t.Execute(ctx, input)
	if err != nil {
		return "", fmt.Errorf("tool %s: %w", name, err)
	}

	slog.Debug("tool result", "name", name, "output_len", len(result))
	return result, nil
}
