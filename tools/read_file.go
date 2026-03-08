package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReadFileTool reads file contents.
type ReadFileTool struct{}

// NewReadFileTool creates a new read_file tool.
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

type readFileInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (r *ReadFileTool) Name() string { return "read_file" }

func (r *ReadFileTool) Description() string {
	return "Read the contents of a file. Optionally specify offset (1-based line number) and limit (max lines) to read a portion."
}

func (r *ReadFileTool) Parameters() json.RawMessage {
	return ObjectSchema(map[string]Property{
		"path":   {Type: "string", Description: "File path to read"},
		"offset": {Type: "integer", Description: "Line number to start from (1-based, default: 1)"},
		"limit":  {Type: "integer", Description: "Maximum number of lines to read (default: all)"},
	}, "path")
}

func (r *ReadFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	params, err := ParseInput(input, func(p *readFileInput) error {
		if p.Path == "" {
			return fmt.Errorf("path is required")
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(params.Path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	// Apply offset (1-based)
	if params.Offset > 0 {
		start := params.Offset - 1
		if start >= len(lines) {
			return "", nil
		}
		lines = lines[start:]
	}

	// Apply limit
	if params.Limit > 0 && params.Limit < len(lines) {
		lines = lines[:params.Limit]
	}

	output := strings.Join(lines, "\n")
	if len(output) > maxOutputLen {
		output = output[:maxOutputLen] + "\n... (output truncated)"
	}

	return output, nil
}
