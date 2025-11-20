package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/denis-kilchichakov/toolbox/agentclient"
)

// Example program demonstrating how to use agentclient within the toolbox project
func main() {
	// Get configuration from environment or use defaults
	serverURL := getEnv("AGENT_SERVER_URL", "http://localhost:8080")
	apiKey := getEnv("AGENT_API_KEY", "your-secret-api-key-here")

	// Create client
	client := agentclient.NewClient(serverURL, apiKey)

	ctx := context.Background()

	// Check server health
	fmt.Println("Checking server health...")
	if err := client.HealthCheck(ctx); err != nil {
		log.Fatalf("Server is not healthy: %v", err)
	}
	fmt.Println("✓ Server is healthy\n")

	// Example questions
	questions := []string{
		"What is the capital of France?",
		"What is 2 + 2?",
		"What are the latest developments in AI?",
	}

	for i, question := range questions {
		fmt.Printf("[%d] Question: %s\n", i+1, question)

		response, err := client.Query(ctx, question)
		if err != nil {
			log.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("    Answer: %s\n", response.Answer)
		fmt.Printf("    Used Search: %v\n", response.UsedSearch)
		fmt.Printf("    Timestamp: %s\n\n", response.Timestamp.Format("2006-01-02 15:04:05"))
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
