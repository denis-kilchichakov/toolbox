# CLAUDE.md - LLM Package

This file provides guidance for Claude Code when working with the `llm` package.

## Package Overview

A modular Go library for interacting with LLM services. Currently supports Ollama with a clean interface that can be extended to other providers (OpenAI, Anthropic, etc.).

## Architecture

```
llm/
├── llm.go          # Core types, interfaces, factory functions
├── errors.go       # Custom error types (ValidationError, ModelNotFoundError, APIError)
├── validation.go   # All input validation logic
├── ollama.go       # Ollama-specific implementation
├── llm_test.go     # Validation unit tests
└── ollama_test.go  # Mock and integration tests
```

### Key Design Patterns

1. **Interface-based design**: `LLMClient` and `Model` interfaces allow multiple implementations
2. **Factory pattern**: `NewLLMClient()` creates appropriate client based on `ServerType`
3. **Two-level API**: Client manages models, Model handles conversations
4. **Custom errors**: Type-safe error handling with specific error types
5. **Validation first**: All inputs validated before API calls

## Testing Strategy

### Two Types of Tests

1. **Mock Tests** (fast, run by default):
   - Use `mockOllamaServer()` to simulate Ollama API
   - Test actual code paths without external dependencies
   - Run in < 1 second
   - Suffix: `_Mock`

2. **Integration Tests** (require real Ollama, skip by default):
   - Test against actual Ollama instance
   - Require environment variables
   - Take several seconds
   - Suffix: `_Integration`

### Running Tests

```bash
# Run all tests (mock only, integration skips)
go test ./llm
go test -v ./llm  # verbose

# Run only mock tests
go test ./llm -run Mock

# Run only integration tests (requires Ollama)
OLLAMA_TEST_URL=http://localhost:11434 OLLAMA_TEST_MODEL=llama3.2:latest go test ./llm -run Integration

# Run all tests including integration
OLLAMA_TEST_URL=http://localhost:11434 OLLAMA_TEST_MODEL=llama3.2:latest go test -v ./llm
```

### Test Coverage Goals

- All validation functions must have tests
- All error paths must be tested
- Mock tests should cover success and failure cases
- Integration tests verify real API compatibility

## Implementation Guidelines

### Adding New Validation

1. Add validation function to `validation.go`
2. Return `*ValidationError` with clear field and message
3. Add unit test to `llm_test.go`
4. Use validation in `ollama.go` methods

Example:
```go
// validation.go
func validateSomething(value string) error {
	if value == "" {
		return &ValidationError{
			Field:   "something",
			Message: "cannot be empty",
		}
	}
	return nil
}

// llm_test.go
func TestValidateSomething(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "test", wantErr: false},
		{name: "empty", value: "", wantErr: true},
	}
	// ... test implementation
}
```

### Adding New LLM Provider

1. Create new file (e.g., `openai.go`)
2. Define provider-specific types (requests/responses)
3. Implement `LLMClient` interface with `*openaiClient`
4. Implement `Model` interface with `*openaiModel`
5. Add factory function `newOpenaiClient()`
6. Update `NewLLMClient()` switch statement in `llm.go`
7. Add new `ServerType` constant
8. Create mock server and tests in `provider_test.go`

### Error Handling

- **ValidationError**: Input validation failures
- **ModelNotFoundError**: Model doesn't exist on server
- **APIError**: HTTP errors from API (includes status code)
- Always wrap errors with context: `fmt.Errorf("context: %w", err)`

### Request Options

- **Temperature**: Controls randomness (0.0+ for Ollama, no upper limit)
- **MaxTokens**: Limits response length (0 = unlimited)
- Default temperature: 0.7
- Always validate: temp >= 0, maxTokens >= 0

### API Patterns

```go
// Always validate inputs first
if err := validatePrompt(prompt); err != nil {
	return nil, err
}

// Default options if not provided
if opts == nil {
	opts = DefaultRequestOptions()
}

// Validate options
if err := opts.Validate(); err != nil {
	return nil, err
}

// Return custom errors, not generic errors
if resp.StatusCode != http.StatusOK {
	return nil, &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}
```

## Important Notes

### Temperature Ranges
- Ollama: 0.0 to infinity (no enforced upper limit)
- Other providers may have different ranges (e.g., OpenAI: 0.0-2.0)
- Current validation only checks >= 0

### Message Roles
- Valid roles: "user", "assistant", "system"
- Ollama supports all three
- Chat messages must alternate user/assistant (per Ollama docs)
- Validation only checks non-empty, not role validity

### Context Management
- All methods accept `context.Context` as first parameter
- Use for timeouts, cancellation, and graceful shutdown
- Mock tests use 5s timeout, integration tests use 60s

### Thread Safety
- Clients are **NOT** thread-safe by default
- Use separate client per goroutine or add mutex if sharing

## Common Tasks

### Add New Method to Model Interface

1. Add method signature to `Model` interface in `llm.go`
2. Implement in `ollamaModel` in `ollama.go`
3. Add validation for inputs
4. Add mock test in `ollama_test.go`
5. Add integration test (if applicable)

### Update Mock Server

Mock server in `ollama_test.go` should:
- Return predictable, testable responses
- Echo inputs to verify they were sent correctly
- Use fixed token counts for assertions
- Handle all endpoints that real implementations use

### Debug Integration Tests

If integration tests fail:
1. Verify Ollama is running: `curl http://localhost:11434/api/tags`
2. Check model exists: `ollama list`
3. Set environment variables correctly
4. Check Ollama logs for errors
5. Increase timeout if model is slow

## DO NOT

- ❌ Don't skip validation - always validate inputs first
- ❌ Don't return generic errors - use custom error types
- ❌ Don't hardcode timeouts - accept context from caller
- ❌ Don't add temperature upper limit (Ollama allows high values)
- ❌ Don't batch requests - keep API simple and stateless
- ❌ Don't add logging - let callers decide how to log
- ❌ Don't modify `interface{}` types to `any` - maintain compatibility

## Future Enhancements

Ideas for future improvements (don't implement unless requested):
- Streaming support for long responses
- Retry logic with exponential backoff
- Token counting utilities
- Response caching
- System message helpers
- Conversation history management
- Model capability detection
- Cost tracking
