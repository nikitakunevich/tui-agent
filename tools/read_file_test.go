package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFileToolName(t *testing.T) {
	r := NewReadFileTool()
	assert.Equal(t, "read_file", r.Name())
}

func TestReadFileToolParameters(t *testing.T) {
	r := NewReadFileTool()
	params := r.Parameters()
	var schema map[string]any
	err := json.Unmarshal(params, &schema)
	require.NoError(t, err)
	assert.Equal(t, "object", schema["type"])
}

func TestReadFileToolReadExistingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644))

	r := NewReadFileTool()
	input, _ := json.Marshal(readFileInput{Path: path})
	result, err := r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3\n", result)
}

func TestReadFileToolOffsetAndLimit(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("line1\nline2\nline3\nline4\nline5"), 0644))

	r := NewReadFileTool()

	// offset=2 → start from line 2
	input, _ := json.Marshal(readFileInput{Path: path, Offset: 2})
	result, err := r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "line2\nline3\nline4\nline5", result)

	// offset=2, limit=2 → lines 2-3
	input, _ = json.Marshal(readFileInput{Path: path, Offset: 2, Limit: 2})
	result, err = r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "line2\nline3", result)

	// limit only
	input, _ = json.Marshal(readFileInput{Path: path, Limit: 3})
	result, err = r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestReadFileToolOffsetBeyondEnd(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("line1\nline2"), 0644))

	r := NewReadFileTool()
	input, _ := json.Marshal(readFileInput{Path: path, Offset: 100})
	result, err := r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestReadFileToolFileNotFound(t *testing.T) {
	r := NewReadFileTool()
	input, _ := json.Marshal(readFileInput{Path: "/nonexistent/file.txt"})
	_, err := r.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}

func TestReadFileToolEmptyPath(t *testing.T) {
	r := NewReadFileTool()
	_, err := r.Execute(context.Background(), json.RawMessage(`{"path":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
}

func TestReadFileToolInvalidInput(t *testing.T) {
	r := NewReadFileTool()
	_, err := r.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestReadFileToolTruncation(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "big.txt")
	// Create content larger than maxOutputLen (50KB)
	big := strings.Repeat("x", maxOutputLen+1000)
	require.NoError(t, os.WriteFile(path, []byte(big), 0644))

	r := NewReadFileTool()
	input, _ := json.Marshal(readFileInput{Path: path})
	result, err := r.Execute(context.Background(), input)
	require.NoError(t, err)
	assert.Contains(t, result, "... (output truncated)")
	assert.LessOrEqual(t, len(result), maxOutputLen+50) // truncated marker overhead
}
