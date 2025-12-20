package agentclient_test

import (
	"context"
	"fmt"
	"log"

	"github.com/denis-kilchichakov/toolbox/agentclient"
)

// Example demonstrates basic usage of the agentclient package
func Example() {
	// Create a new client
	client := agentclient.NewClient(
		"http://localhost:8080",
		"your-api-key-here",
	)

	// Create context
	ctx := context.Background()

	// Check if the server is healthy
	if err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("Server health check failed: %v", err)
	}
	fmt.Println("Server is healthy")

	// Send a query
	response, err := client.Query(ctx, "What is the capital of France?")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	// Display the response
	fmt.Printf("Question: %s\n", response.Question)
	fmt.Printf("Answer: %s\n", response.Answer)
	fmt.Printf("Used Search: %v\n", response.UsedSearch)
	fmt.Printf("Timestamp: %s\n", response.Timestamp)
}

// Example_withCustomTimeout demonstrates using a custom timeout
func Example_withCustomTimeout() {
	client := agentclient.NewClient(
		"http://localhost:8080",
		"your-api-key-here",
	)

	// Set a custom timeout (default is 120 seconds)
	client.SetTimeout(60 * 1000000000) // 60 seconds

	ctx := context.Background()
	response, err := client.Query(ctx, "Explain quantum computing")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Answer: %s\n", response.Answer)
}
