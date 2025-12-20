# AgentClient

A Go client library for interacting with the Ollama-MCP HTTP server that provides AI-powered question answering with intelligent web search capabilities.

## Features

- Simple HTTP client for querying the Ollama-MCP server
- Automatic API key authentication
- Health check support
- Configurable timeouts
- Context support for cancellation and deadlines

## Installation

The package is part of the toolbox module:

```go
import "github.com/denis-kilchichakov/toolbox/agentclient"
```

## Prerequisites

You need a running instance of the Ollama-MCP server. See the main project documentation for setup instructions.

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/denis-kilchichakov/toolbox/agentclient"
)

func main() {
    // Create a new client
    client := agentclient.NewClient(
        "http://localhost:8080",
        "your-api-key-here",
    )

    ctx := context.Background()

    // Check server health
    if err := client.HealthCheck(ctx); err != nil {
        log.Fatalf("Server health check failed: %v", err)
    }

    // Send a query
    response, err := client.Query(ctx, "What is quantum computing?")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    fmt.Printf("Answer: %s\n", response.Answer)
    fmt.Printf("Used Search: %v\n", response.UsedSearch)
}
```

### Custom Timeout

```go
client := agentclient.NewClient(baseURL, apiKey)

// Set custom timeout (default is 120 seconds)
client.SetTimeout(60 * time.Second)

response, err := client.Query(ctx, "Your question")
```

### With Context Cancellation

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := client.Query(ctx, "Your question")
```

## API Reference

### Types

#### `Client`

The main client struct for interacting with the server.

#### `QueryRequest`

```go
type QueryRequest struct {
    Question string `json:"question"`
}
```

#### `QueryResponse`

```go
type QueryResponse struct {
    Question   string    `json:"question"`
    Answer     string    `json:"answer"`
    UsedSearch bool      `json:"used_search"`
    Timestamp  time.Time `json:"timestamp"`
}
```

#### `ErrorResponse`

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}
```

### Functions

#### `NewClient(baseURL, apiKey string) *Client`

Creates a new client instance.

**Parameters:**
- `baseURL` - The base URL of the server (e.g., "http://localhost:8080")
- `apiKey` - The API key for authentication

**Returns:** A new `*Client` instance

#### `(c *Client) Query(ctx context.Context, question string) (*QueryResponse, error)`

Sends a question to the server and returns the AI-generated response.

**Parameters:**
- `ctx` - Context for cancellation and timeout
- `question` - The question to ask

**Returns:** A `*QueryResponse` containing the answer, or an error

#### `(c *Client) HealthCheck(ctx context.Context) error`

Checks if the server is healthy and accessible.

**Parameters:**
- `ctx` - Context for cancellation and timeout

**Returns:** An error if the server is unhealthy, nil otherwise

#### `(c *Client) SetTimeout(timeout time.Duration)`

Sets a custom timeout for HTTP requests.

**Parameters:**
- `timeout` - The timeout duration

## Error Handling

The client returns descriptive errors for various failure scenarios:

```go
response, err := client.Query(ctx, question)
if err != nil {
    // Handle errors
    log.Printf("Query failed: %v", err)
    return
}
```

Common error cases:
- Network errors (server unreachable)
- Authentication errors (invalid API key)
- Server errors (internal server issues)
- Timeout errors (request took too long)

## Configuration

Configure the server via environment variables on the server side:

- `API_KEY` - The API key for authentication
- `OLLAMA_URL` - Ollama API endpoint (default: http://localhost:11434)
- `MODEL_NAME` - Ollama model to use (default: gpt-oss:20b)
- `SERVER_PORT` - HTTP server port (default: 8080)

## License

Part of the toolbox project.
