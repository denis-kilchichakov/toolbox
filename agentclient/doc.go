// Package agentclient provides a Go client library for interacting with
// the Ollama-MCP HTTP server.
//
// The client enables AI-powered question answering with intelligent web search
// capabilities. The server automatically decides whether to use web search
// based on the question asked.
//
// # Basic Usage
//
//	client := agentclient.NewClient("http://localhost:8080", "your-api-key")
//	response, err := client.Query(context.Background(), "What is quantum computing?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response.Answer)
//
// # Features
//
//   - Simple HTTP client with automatic API key authentication
//   - Health check support for monitoring
//   - Configurable request timeouts
//   - Full context support for cancellation and deadlines
//   - Detailed error responses
//
// # Server Setup
//
// Before using this client, you need to have the Ollama-MCP server running.
// The server integrates Ollama LLM with Brave Search via Model Context Protocol.
//
// See the main project repository for server setup instructions.
package agentclient
