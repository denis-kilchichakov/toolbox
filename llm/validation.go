package llm

import "fmt"

// Validate checks if the RequestOptions are valid
func (o *RequestOptions) Validate() error {
	if o.Temperature < 0 {
		return &ValidationError{
			Field:   "Temperature",
			Message: "must be >= 0",
		}
	}
	if o.MaxTokens < 0 {
		return &ValidationError{
			Field:   "MaxTokens",
			Message: "must be >= 0",
		}
	}
	return nil
}

// validatePrompt checks if a prompt is valid
func validatePrompt(prompt string) error {
	if prompt == "" {
		return &ValidationError{
			Field:   "prompt",
			Message: "cannot be empty",
		}
	}
	return nil
}

// validateModelName checks if a model name is valid
func validateModelName(name string) error {
	if name == "" {
		return &ValidationError{
			Field:   "model name",
			Message: "cannot be empty",
		}
	}
	return nil
}

// validateMessages checks if messages are valid
func validateMessages(messages []Message) error {
	if len(messages) == 0 {
		return &ValidationError{
			Field:   "messages",
			Message: "cannot be empty",
		}
	}
	for i, msg := range messages {
		if msg.Role == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("messages[%d].Role", i),
				Message: "cannot be empty",
			}
		}
		if msg.Content == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("messages[%d].Content", i),
				Message: "cannot be empty",
			}
		}
	}
	return nil
}
