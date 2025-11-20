package llm

import (
	"context"
	"fmt"
)

// ServerType defines the type of LLM server
type ServerType string

const (
	ServerTypeOllama ServerType = "ollama"
)

// LLMConfig holds configuration for LLM client initialization
type LLMConfig struct {
	// ServerType specifies the LLM server type (e.g., "ollama")
	ServerType ServerType
	// URL is the base URL of the LLM server (e.g., "http://localhost:11434")
	URL string
}

// ModelInfo represents metadata about an available LLM model
type ModelInfo struct {
	Name string
	Size int64
}

// Message represents a single message in a conversation
type Message struct {
	Role    string // "user", "assistant", or "system"
	Content string
}

// RequestOptions contains optional parameters for LLM requests
type RequestOptions struct {
	// Temperature controls randomness (0.0 to 1.0, lower is more deterministic)
	Temperature float64
	// MaxTokens limits the response length (0 means no limit)
	MaxTokens int
}

// DefaultRequestOptions returns default request options
func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{
		Temperature: 0.7,
		MaxTokens:   0, // No limit
	}
}

// Response represents the LLM's response
type Response struct {
	Content      string
	FinishReason string // "stop", "length", "error", etc.
	TokensUsed   int
}

// Model defines the interface for interacting with a specific LLM model
type Model interface {
	// Ask sends a single prompt and returns the response
	Ask(ctx context.Context, prompt string, opts *RequestOptions) (*Response, error)

	// Chat sends a conversation history and returns the response
	Chat(ctx context.Context, messages []Message, opts *RequestOptions) (*Response, error)
}

// LLMClient defines the interface for interacting with LLM services
type LLMClient interface {
	// ListModels returns a list of available models
	ListModels(ctx context.Context) ([]ModelInfo, error)

	// GetModel returns a Model interface for the specified model name
	GetModel(ctx context.Context, name string) (Model, error)

	// Close cleans up any resources used by the client
	Close() error
}

// NewLLMClient creates a new LLM client based on the provided configuration
func NewLLMClient(ctx context.Context, config LLMConfig) (LLMClient, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	switch config.ServerType {
	case ServerTypeOllama:
		return newOllamaClient(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported server type: %s", config.ServerType)
	}
}
