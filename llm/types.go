package llm

import "encoding/json"

// Role constants for messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Arguments string `json:"arguments"` // raw JSON
}

// ToolDef defines a tool that can be passed to the LLM.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

// ChatRequest is the input to Provider.Chat.
type ChatRequest struct {
	Messages []Message
	Tools    []ToolDef
}

// ChatResponse is the output from Provider.Chat.
type ChatResponse struct {
	Content    string     // text content (if any)
	ToolCalls  []ToolCall // tool calls (if any)
	StopReason string     // "end_turn" | "tool_use"
}

// StopReasonFromToolCalls returns "tool_use" if there are tool calls, "end_turn" otherwise.
func StopReasonFromToolCalls(calls []ToolCall) string {
	if len(calls) > 0 {
		return "tool_use"
	}
	return "end_turn"
}
