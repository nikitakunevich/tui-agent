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
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "File path to read"
			},
			"offset": {
				"type": "integer",
				"description": "Line number to start from (1-based, default: 1)"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum number of lines to read (default: all)"
			}
		},
		"required": ["path"]
	}`)
}

func (r *ReadFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params readFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if params.Path == "" {
		return "", fmt.Errorf("path is required")
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
