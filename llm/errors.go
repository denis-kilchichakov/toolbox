package llm

import "fmt"

// ValidationError represents an input validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// ModelNotFoundError represents an error when a model is not found
type ModelNotFoundError struct {
	ModelName string
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("model '%s' not found", e.ModelName)
}

// APIError represents an error from the LLM API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}
