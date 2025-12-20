package agentclient

import "time"

// QueryRequest represents a question sent to the server
type QueryRequest struct {
	Question string `json:"question"`
}

// QueryResponse represents the server's answer
type QueryResponse struct {
	Question   string    `json:"question"`
	Answer     string    `json:"answer"`
	UsedSearch bool      `json:"used_search"`
	Timestamp  time.Time `json:"timestamp"`
}

// ErrorResponse represents an error from the server
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
