package llm

import "context"

// Provider is the interface for LLM backends.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
