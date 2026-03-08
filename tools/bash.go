package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	maxOutputLen     = 50000 // 50KB
)

// BashTool executes bash commands.
type BashTool struct {
	Timeout    time.Duration
	WorkingDir string
}

// NewBashTool creates a new bash tool with default settings.
func NewBashTool() *BashTool {
	return &BashTool{
		Timeout: defaultTimeout,
	}
}

type bashInput struct {
	Command string `json:"command"`
}

func (b *BashTool) Name() string { return "bash" }

func (b *BashTool) Description() string {
	return "Execute a bash command and return stdout+stderr. Use this to run shell commands."
}

func (b *BashTool) Parameters() json.RawMessage {
	return ObjectSchema(map[string]Property{
		"command": {Type: "string", Description: "The bash command to execute"},
	}, "command")
}

func (b *BashTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	params, err := ParseInput(input, func(p *bashInput) error {
		if p.Command == "" {
			return fmt.Errorf("command is required")
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	timeout := b.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", params.Command)
	if b.WorkingDir != "" {
		cmd.Dir = b.WorkingDir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()

	output := out.String()
	if len(output) > maxOutputLen {
		output = output[:maxOutputLen] + "\n... (output truncated)"
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return output + "\n(command timed out)", fmt.Errorf("command timed out after %v", timeout)
		}
		// Include the error in output but still return it (non-zero exit is common)
		return fmt.Sprintf("%s\nexit status: %v", output, err), nil
	}

	return output, nil
}
