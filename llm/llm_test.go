package llm

import (
	"testing"
)

func TestRequestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    RequestOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: RequestOptions{
				Temperature: 0.7,
				MaxTokens:   100,
			},
			wantErr: false,
		},
		{
			name: "valid with zero temperature",
			opts: RequestOptions{
				Temperature: 0.0,
				MaxTokens:   0,
			},
			wantErr: false,
		},
		{
			name: "valid with high temperature",
			opts: RequestOptions{
				Temperature: 10.0,
				MaxTokens:   0,
			},
			wantErr: false,
		},
		{
			name: "invalid negative temperature",
			opts: RequestOptions{
				Temperature: -0.1,
				MaxTokens:   100,
			},
			wantErr: true,
			errMsg:  "Temperature",
		},
		{
			name: "invalid negative max tokens",
			opts: RequestOptions{
				Temperature: 0.7,
				MaxTokens:   -1,
			},
			wantErr: true,
			errMsg:  "MaxTokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if verr, ok := err.(*ValidationError); ok {
					if verr.Field != tt.errMsg {
						t.Errorf("Validate() error field = %v, want %v", verr.Field, tt.errMsg)
					}
				} else {
					t.Errorf("Validate() error type = %T, want *ValidationError", err)
				}
			}
		})
	}
}

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
	}{
		{
			name:    "valid prompt",
			prompt:  "Hello, world!",
			wantErr: false,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePrompt(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModelName(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		wantErr   bool
	}{
		{
			name:      "valid model name",
			modelName: "llama3.2:latest",
			wantErr:   false,
		},
		{
			name:      "empty model name",
			modelName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModelName(tt.modelName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModelName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		wantErr  bool
		errField string
	}{
		{
			name: "valid messages",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			wantErr: false,
		},
		{
			name:     "empty messages",
			messages: []Message{},
			wantErr:  true,
			errField: "messages",
		},
		{
			name: "message with empty role",
			messages: []Message{
				{Role: "", Content: "Hello"},
			},
			wantErr:  true,
			errField: "messages[0].Role",
		},
		{
			name: "message with empty content",
			messages: []Message{
				{Role: "user", Content: ""},
			},
			wantErr:  true,
			errField: "messages[0].Content",
		},
		{
			name: "second message with empty content",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: ""},
			},
			wantErr:  true,
			errField: "messages[1].Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if verr, ok := err.(*ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("validateMessages() error field = %v, want %v", verr.Field, tt.errField)
					}
				} else {
					t.Errorf("validateMessages() error type = %T, want *ValidationError", err)
				}
			}
		})
	}
}

func TestDefaultRequestOptions(t *testing.T) {
	opts := DefaultRequestOptions()
	if opts == nil {
		t.Fatal("DefaultRequestOptions() returned nil")
	}
	if opts.Temperature != 0.7 {
		t.Errorf("DefaultRequestOptions().Temperature = %v, want 0.7", opts.Temperature)
	}
	if opts.MaxTokens != 0 {
		t.Errorf("DefaultRequestOptions().MaxTokens = %v, want 0", opts.MaxTokens)
	}
}
