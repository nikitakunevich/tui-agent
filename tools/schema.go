package tools

import (
	"encoding/json"
	"fmt"
)

// ParseInput unmarshals JSON input into a typed struct and runs an optional validation function.
func ParseInput[T any](input json.RawMessage, validate func(*T) error) (*T, error) {
	var v T
	if err := json.Unmarshal(input, &v); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if validate != nil {
		if err := validate(&v); err != nil {
			return nil, err
		}
	}
	return &v, nil
}

// Property describes a single property in a JSON Schema object.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ObjectSchema builds a JSON Schema "object" with the given properties and required fields.
func ObjectSchema(props map[string]Property, required ...string) json.RawMessage {
	schema := map[string]any{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	b, _ := json.Marshal(schema)
	return b
}
