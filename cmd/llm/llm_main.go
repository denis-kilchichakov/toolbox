package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/denis-kilchichakov/toolbox/llm"
)

func main() {
	config := llm.LLMConfig{
		ServerType: llm.ServerTypeOllama,
		URL:        "http://localhost:11434",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Printf("Connecting to Ollama at %s...\n", config.URL)

	client, err := llm.NewLLMClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	defer client.Close()

	fmt.Println("✓ Successfully connected to Ollama")

	// List available models
	fmt.Println("\nFetching available models...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatalf("Failed to list models: %v", err)
	}

	fmt.Printf("\nFound %d models:\n", len(models))
	for _, modelInfo := range models {
		sizeMB := float64(modelInfo.Size) / (1024 * 1024)
		fmt.Printf("  - %s (%.1f MB)\n", modelInfo.Name, sizeMB)
	}

	// Get a specific model
	modelName := "llama3.2:latest"
	fmt.Printf("\nGetting model '%s'...\n", modelName)
	model, err := client.GetModel(ctx, modelName)
	if err != nil {
		log.Fatalf("Failed to get model: %v", err)
	}

	fmt.Printf("✓ Model '%s' is ready to use\n", modelName)

	// Test the same question with different temperatures
	question := "Give me slogan for my new IT company"
	temperatures := []float64{0.1, 0.4, 0.7, 1.5, 2, 3, 5, 10}

	fmt.Println("\n" + "===============================================================================")
	fmt.Printf("Testing question: \"%s\"\n", question)
	fmt.Println("===============================================================================")

	for _, temp := range temperatures {
		fmt.Printf("\n--- Temperature: %.1f ---\n", temp)

		opts := &llm.RequestOptions{
			Temperature: temp,
			MaxTokens:   200,
		}

		response, err := model.Ask(ctx, question, opts)
		if err != nil {
			log.Printf("Error with temperature %.1f: %v", temp, err)
			continue
		}

		fmt.Printf("Response: %s\n", response.Content)
		fmt.Printf("Tokens used: %d\n", response.TokensUsed)
	}

	fmt.Println("\n" + "===============================================================================")
	fmt.Println("Note: Lower temperature = more deterministic, Higher temperature = more creative")

	// Test Chat() with multi-turn conversation
	fmt.Println("\n\n" + "===============================================================================")
	fmt.Println("Testing Chat() - Multi-turn Conversation")
	fmt.Println("===============================================================================")

	messages := []llm.Message{
		{Role: "user", Content: "My favorite color is blue."},
		{Role: "assistant", Content: "That's nice! Blue is a calming color."},
		{Role: "user", Content: "What was my favorite color?"},
	}

	chatOpts := &llm.RequestOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}

	fmt.Println("\nConversation:")
	for _, msg := range messages {
		fmt.Printf("  [%s]: %s\n", msg.Role, msg.Content)
	}

	fmt.Println("\nSending to model...")
	chatResponse, err := model.Chat(ctx, messages, chatOpts)
	if err != nil {
		log.Printf("Chat error: %v", err)
	} else {
		fmt.Printf("\n  [assistant]: %s\n", chatResponse.Content)
		fmt.Printf("\nTokens used: %d\n", chatResponse.TokensUsed)
	}

	fmt.Println("\n" + "===============================================================================")
}
