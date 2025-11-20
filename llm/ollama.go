package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ollamaTagsResponse represents the response from /api/tags endpoint
type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

// ollamaGenerateRequest represents the request to /api/generate endpoint
type ollamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// ollamaGenerateResponse represents the response from /api/generate endpoint
type ollamaGenerateResponse struct {
	Model      string `json:"model"`
	CreatedAt  string `json:"created_at"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	EvalCount  int    `json:"eval_count"`
	DoneReason string `json:"done_reason,omitempty"`
}

// ollamaChatRequest represents the request to /api/chat endpoint
type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaChatMessage    `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ollamaChatMessage represents a message in the chat request
type ollamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatResponse represents the response from /api/chat endpoint
type ollamaChatResponse struct {
	Model      string              `json:"model"`
	CreatedAt  string              `json:"created_at"`
	Message    ollamaChatMessage   `json:"message"`
	Done       bool                `json:"done"`
	EvalCount  int                 `json:"eval_count"`
	DoneReason string              `json:"done_reason,omitempty"`
}

// ollamaClient implements LLMClient for Ollama
type ollamaClient struct {
	config     LLMConfig
	httpClient *http.Client
}

// ollamaModel implements Model interface for Ollama
type ollamaModel struct {
	client    *ollamaClient
	modelName string
}

// newOllamaClient creates a new Ollama client
func newOllamaClient(_ context.Context, config LLMConfig) (*ollamaClient, error) {
	client := &ollamaClient{
		config:     config,
		httpClient: &http.Client{},
	}

	return client, nil
}

// ListModels returns a list of available models from the Ollama server
func (c *ollamaClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/api/tags", c.config.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tagsResp ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, len(tagsResp.Models))
	for i, m := range tagsResp.Models {
		models[i] = ModelInfo{
			Name: m.Name,
			Size: m.Size,
		}
	}

	return models, nil
}

// GetModel returns a Model interface for the specified model name
func (c *ollamaClient) GetModel(ctx context.Context, name string) (Model, error) {
	// Validate model name
	if err := validateModelName(name); err != nil {
		return nil, err
	}

	// Verify the model exists
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	found := false
	for _, m := range models {
		if m.Name == name {
			found = true
			break
		}
	}

	if !found {
		return nil, &ModelNotFoundError{ModelName: name}
	}

	return &ollamaModel{
		client:    c,
		modelName: name,
	}, nil
}

// Close cleans up any resources used by the client
func (c *ollamaClient) Close() error {
	// No resources to clean up for now
	return nil
}

// Ask sends a single prompt and returns the response
func (m *ollamaModel) Ask(ctx context.Context, prompt string, opts *RequestOptions) (*Response, error) {
	// Validate prompt
	if err := validatePrompt(prompt); err != nil {
		return nil, err
	}

	// Use default options if not provided
	if opts == nil {
		opts = DefaultRequestOptions()
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Build request options
	options := make(map[string]interface{})
	options["temperature"] = opts.Temperature
	if opts.MaxTokens > 0 {
		options["num_predict"] = opts.MaxTokens
	}

	// Create request
	reqBody := ollamaGenerateRequest{
		Model:   m.modelName,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", m.client.config.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse response
	var genResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	finishReason := "stop"
	if genResp.DoneReason != "" {
		finishReason = genResp.DoneReason
	}

	return &Response{
		Content:      genResp.Response,
		FinishReason: finishReason,
		TokensUsed:   genResp.EvalCount,
	}, nil
}

// Chat sends a conversation history and returns the response
func (m *ollamaModel) Chat(ctx context.Context, messages []Message, opts *RequestOptions) (*Response, error) {
	// Validate messages
	if err := validateMessages(messages); err != nil {
		return nil, err
	}

	// Use default options if not provided
	if opts == nil {
		opts = DefaultRequestOptions()
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Build request options
	options := make(map[string]interface{})
	options["temperature"] = opts.Temperature
	if opts.MaxTokens > 0 {
		options["num_predict"] = opts.MaxTokens
	}

	// Convert messages to Ollama format
	ollamaMessages := make([]ollamaChatMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	reqBody := ollamaChatRequest{
		Model:    m.modelName,
		Messages: ollamaMessages,
		Stream:   false,
		Options:  options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", m.client.config.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := m.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse response
	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	finishReason := "stop"
	if chatResp.DoneReason != "" {
		finishReason = chatResp.DoneReason
	}

	return &Response{
		Content:      chatResp.Message.Content,
		FinishReason: finishReason,
		TokensUsed:   chatResp.EvalCount,
	}, nil
}
