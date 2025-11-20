package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// mockOllamaServer creates a mock HTTP server that simulates Ollama API
func mockOllamaServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock /api/tags endpoint
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		response := ollamaTagsResponse{
			Models: []struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			}{
				{Name: "test-model:latest", Size: 1000000},
				{Name: "another-model:v1", Size: 2000000},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Mock /api/generate endpoint
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		var req ollamaGenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := ollamaGenerateResponse{
			Model:      req.Model,
			CreatedAt:  "2024-01-01T00:00:00Z",
			Response:   "This is a mock response to: " + req.Prompt,
			Done:       true,
			EvalCount:  10,
			DoneReason: "stop",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Mock /api/chat endpoint
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate response based on conversation
		lastMessage := ""
		if len(req.Messages) > 0 {
			lastMessage = req.Messages[len(req.Messages)-1].Content
		}

		response := ollamaChatResponse{
			Model:     req.Model,
			CreatedAt: "2024-01-01T00:00:00Z",
			Message: ollamaChatMessage{
				Role:    "assistant",
				Content: "Mock chat response to: " + lastMessage,
			},
			Done:       true,
			EvalCount:  15,
			DoneReason: "stop",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(mux)
}

// skipIfNoOllama skips the test if Ollama is not available
func skipIfNoOllama(t *testing.T) string {
	url := os.Getenv("OLLAMA_TEST_URL")
	if url == "" {
		t.Skip("Skipping integration test: OLLAMA_TEST_URL not set")
	}
	return url
}

// ============================================================================
// UNIT TESTS WITH MOCK SERVER
// ============================================================================

func TestOllamaClient_ListModels_Mock(t *testing.T) {
	server := mockOllamaServer()
	defer server.Close()

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	expectedNames := map[string]bool{
		"test-model:latest":  true,
		"another-model:v1":   true,
	}

	for _, model := range models {
		if !expectedNames[model.Name] {
			t.Errorf("Unexpected model name: %s", model.Name)
		}
		if model.Size <= 0 {
			t.Errorf("Model %s has invalid size: %d", model.Name, model.Size)
		}
	}
}

func TestOllamaClient_GetModel_Mock(t *testing.T) {
	server := mockOllamaServer()
	defer server.Close()

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name      string
		modelName string
		wantErr   bool
		errType   interface{}
	}{
		{
			name:      "valid model",
			modelName: "test-model:latest",
			wantErr:   false,
		},
		{
			name:      "empty model name",
			modelName: "",
			wantErr:   true,
			errType:   &ValidationError{},
		},
		{
			name:      "non-existent model",
			modelName: "non-existent-model",
			wantErr:   true,
			errType:   &ModelNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := client.GetModel(ctx, tt.modelName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && model == nil {
				t.Error("GetModel() returned nil model")
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("GetModel() error type = %T, want *ValidationError", err)
					}
				case *ModelNotFoundError:
					if _, ok := err.(*ModelNotFoundError); !ok {
						t.Errorf("GetModel() error type = %T, want *ModelNotFoundError", err)
					}
				}
			}
		})
	}
}

func TestOllamaModel_Ask_Mock(t *testing.T) {
	server := mockOllamaServer()
	defer server.Close()

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model, err := client.GetModel(ctx, "test-model:latest")
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}

	tests := []struct {
		name    string
		prompt  string
		opts    *RequestOptions
		wantErr bool
		errType interface{}
	}{
		{
			name:    "simple question",
			prompt:  "What is 2+2?",
			opts:    nil,
			wantErr: false,
		},
		{
			name:   "with custom temperature",
			prompt: "Say hello",
			opts: &RequestOptions{
				Temperature: 0.1,
				MaxTokens:   50,
			},
			wantErr: false,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			opts:    nil,
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:   "invalid temperature",
			prompt: "Hello",
			opts: &RequestOptions{
				Temperature: -1.0,
				MaxTokens:   50,
			},
			wantErr: true,
			errType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := model.Ask(ctx, tt.prompt, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if response == nil {
					t.Error("Ask() returned nil response")
					return
				}
				if response.Content == "" {
					t.Error("Ask() returned empty content")
				}
				if !strings.Contains(response.Content, tt.prompt) {
					t.Errorf("Ask() response doesn't contain prompt. Response: %s", response.Content)
				}
				if response.TokensUsed != 10 {
					t.Errorf("Ask() returned tokens = %d, want 10", response.TokensUsed)
				}
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Ask() error type = %T, want *ValidationError", err)
					}
				}
			}
		})
	}
}

func TestOllamaModel_Chat_Mock(t *testing.T) {
	server := mockOllamaServer()
	defer server.Close()

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model, err := client.GetModel(ctx, "test-model:latest")
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}

	tests := []struct {
		name     string
		messages []Message
		opts     *RequestOptions
		wantErr  bool
		errType  interface{}
	}{
		{
			name: "simple conversation",
			messages: []Message{
				{Role: "user", Content: "What is 2+2?"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name: "multi-turn conversation",
			messages: []Message{
				{Role: "user", Content: "My name is Alice"},
				{Role: "assistant", Content: "Nice to meet you, Alice!"},
				{Role: "user", Content: "What is my name?"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name:     "empty messages",
			messages: []Message{},
			opts:     nil,
			wantErr:  true,
			errType:  &ValidationError{},
		},
		{
			name: "message with empty content",
			messages: []Message{
				{Role: "user", Content: ""},
			},
			opts:    nil,
			wantErr: true,
			errType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := model.Chat(ctx, tt.messages, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if response == nil {
					t.Error("Chat() returned nil response")
					return
				}
				if response.Content == "" {
					t.Error("Chat() returned empty content")
				}
				if response.TokensUsed != 15 {
					t.Errorf("Chat() returned tokens = %d, want 15", response.TokensUsed)
				}
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Chat() error type = %T, want *ValidationError", err)
					}
				}
			}
		})
	}
}

// ============================================================================
// INTEGRATION TESTS WITH REAL OLLAMA (requires OLLAMA_TEST_URL env var)
// ============================================================================

func TestOllamaClient_ListModels_Integration(t *testing.T) {
	url := skipIfNoOllama(t)

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	if len(models) == 0 {
		t.Error("Expected at least one model, got none")
	}

	// Check that models have required fields
	for _, model := range models {
		if model.Name == "" {
			t.Error("Model has empty name")
		}
		if model.Size <= 0 {
			t.Errorf("Model %s has invalid size: %d", model.Name, model.Size)
		}
	}
}

func TestOllamaClient_GetModel_Integration(t *testing.T) {
	url := skipIfNoOllama(t)
	modelName := os.Getenv("OLLAMA_TEST_MODEL")
	if modelName == "" {
		t.Skip("Skipping: OLLAMA_TEST_MODEL not set")
	}

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name      string
		modelName string
		wantErr   bool
		errType   interface{}
	}{
		{
			name:      "valid model",
			modelName: modelName,
			wantErr:   false,
		},
		{
			name:      "empty model name",
			modelName: "",
			wantErr:   true,
			errType:   &ValidationError{},
		},
		{
			name:      "non-existent model",
			modelName: "this-model-does-not-exist-12345",
			wantErr:   true,
			errType:   &ModelNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := client.GetModel(ctx, tt.modelName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && model == nil {
				t.Error("GetModel() returned nil model")
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("GetModel() error type = %T, want *ValidationError", err)
					}
				case *ModelNotFoundError:
					if _, ok := err.(*ModelNotFoundError); !ok {
						t.Errorf("GetModel() error type = %T, want *ModelNotFoundError", err)
					}
				}
			}
		})
	}
}

func TestOllamaModel_Ask_Integration(t *testing.T) {
	url := skipIfNoOllama(t)
	modelName := os.Getenv("OLLAMA_TEST_MODEL")
	if modelName == "" {
		t.Skip("Skipping: OLLAMA_TEST_MODEL not set")
	}

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model, err := client.GetModel(ctx, modelName)
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}

	tests := []struct {
		name    string
		prompt  string
		opts    *RequestOptions
		wantErr bool
		errType interface{}
	}{
		{
			name:    "simple question",
			prompt:  "What is 2+2?",
			opts:    nil,
			wantErr: false,
		},
		{
			name:   "with custom temperature",
			prompt: "Say hello",
			opts: &RequestOptions{
				Temperature: 0.1,
				MaxTokens:   50,
			},
			wantErr: false,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			opts:    nil,
			wantErr: true,
			errType: &ValidationError{},
		},
		{
			name:   "invalid temperature",
			prompt: "Hello",
			opts: &RequestOptions{
				Temperature: -1.0,
				MaxTokens:   50,
			},
			wantErr: true,
			errType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := model.Ask(ctx, tt.prompt, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if response == nil {
					t.Error("Ask() returned nil response")
					return
				}
				if response.Content == "" {
					t.Error("Ask() returned empty content")
				}
				if response.TokensUsed < 0 {
					t.Errorf("Ask() returned negative tokens: %d", response.TokensUsed)
				}
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Ask() error type = %T, want *ValidationError", err)
					}
				}
			}
		})
	}
}

func TestOllamaModel_Chat_Integration(t *testing.T) {
	url := skipIfNoOllama(t)
	modelName := os.Getenv("OLLAMA_TEST_MODEL")
	if modelName == "" {
		t.Skip("Skipping: OLLAMA_TEST_MODEL not set")
	}

	config := LLMConfig{
		ServerType: ServerTypeOllama,
		URL:        url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := NewLLMClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model, err := client.GetModel(ctx, modelName)
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}

	tests := []struct {
		name     string
		messages []Message
		opts     *RequestOptions
		wantErr  bool
		errType  interface{}
	}{
		{
			name: "simple conversation",
			messages: []Message{
				{Role: "user", Content: "What is 2+2?"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name: "multi-turn conversation",
			messages: []Message{
				{Role: "user", Content: "My name is Alice"},
				{Role: "assistant", Content: "Nice to meet you, Alice!"},
				{Role: "user", Content: "What is my name?"},
			},
			opts:    nil,
			wantErr: false,
		},
		{
			name:     "empty messages",
			messages: []Message{},
			opts:     nil,
			wantErr:  true,
			errType:  &ValidationError{},
		},
		{
			name: "message with empty content",
			messages: []Message{
				{Role: "user", Content: ""},
			},
			opts:    nil,
			wantErr: true,
			errType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := model.Chat(ctx, tt.messages, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if response == nil {
					t.Error("Chat() returned nil response")
					return
				}
				if response.Content == "" {
					t.Error("Chat() returned empty content")
				}
				if response.TokensUsed < 0 {
					t.Errorf("Chat() returned negative tokens: %d", response.TokensUsed)
				}
			}
			if tt.wantErr && tt.errType != nil {
				switch tt.errType.(type) {
				case *ValidationError:
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Chat() error type = %T, want *ValidationError", err)
					}
				}
			}
		})
	}
}
