package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/denis-kilchichakov/toolbox/telegram"
)

func main() {
	// Get bot token from environment variable
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Create bot with default configuration
	// This uses sensible defaults (Timeout: 60s, Debug: false)
	config := telegram.DefaultConfig(token)

	// Override specific settings if needed
	config.Debug = true

	bot, err := telegram.NewBot(config)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Println("Bot started successfully. Press Ctrl+C to stop.")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping bot...")
		cancel()
		bot.Close()
	}()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopped")
			return

		case update, ok := <-bot.Updates():
			if !ok {
				log.Println("Updates channel closed")
				return
			}

			// Handle different types of updates
			go handleUpdate(update)
		}
	}
}

// handleUpdate processes a single update
func handleUpdate(update telegram.Update) {
	// Handle text messages
	if update.Message != nil {
		handleMessage(update.Message)
		return
	}

	// Handle callback queries (from inline keyboards)
	if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery)
		return
	}
}

// handleMessage processes text messages
func handleMessage(msg *telegram.Message) {
	// Log the message
	log.Printf("[Message #%d] From: %s (@%s) | Chat ID: %d | Text: %s",
		msg.ID,
		msg.From.FirstName,
		msg.From.Username,
		msg.Chat.ID,
		msg.Text,
	)

	// Handle commands using proper entity detection
	if msg.IsCommand() {
		handleCommand(msg)
		return
	}

	// Handle regular text messages
	handleTextMessage(msg)
}

// handleCommand processes bot commands
func handleCommand(msg *telegram.Message) {
	// Use the proper Command() method which handles @botname mentions
	command := msg.Command()
	args := msg.CommandArguments()

	log.Printf("Command: %s | Args: %s", command, args)

	switch command {
	case "start":
		log.Printf("User %s started the bot", msg.From.FirstName)
		// In a real bot, you would send a welcome message here
		fmt.Printf("Welcome message would be sent to chat %d\n", msg.Chat.ID)

	case "help":
		log.Printf("User %s requested help", msg.From.FirstName)
		// In a real bot, you would send help information
		fmt.Printf("Help message would be sent to chat %d\n", msg.Chat.ID)

	case "status":
		log.Printf("User %s requested status", msg.From.FirstName)
		// Example: show bot status
		fmt.Printf("Bot status: Running since startup. Current time: %s\n", time.Now().Format(time.RFC3339))

	case "echo":
		// Echo back the arguments
		if args != "" {
			fmt.Printf("Would echo to chat %d: %s\n", msg.Chat.ID, args)
		} else {
			fmt.Printf("Echo command received but no text provided\n")
		}

	default:
		log.Printf("Unknown command: %s", command)
		fmt.Printf("Unknown command message would be sent to chat %d\n", msg.Chat.ID)
	}
}

// handleTextMessage processes regular (non-command) text messages
func handleTextMessage(msg *telegram.Message) {
	// Example: log and respond to regular messages
	log.Printf("Received regular message from %s: %s", msg.From.FirstName, msg.Text)

	// You could implement logic here, such as:
	// - Keyword detection
	// - AI/LLM integration
	// - Database lookups
	// - State machine for conversations

	// For demo purposes, just log some statistics
	wordCount := len(strings.Fields(msg.Text))
	log.Printf("Message has %d words", wordCount)
}

// handleCallbackQuery processes callback queries from inline keyboards
func handleCallbackQuery(query *telegram.CallbackQuery) {
	log.Printf("[Callback Query] ID: %s | From: %s | Data: %s",
		query.ID,
		query.From.FirstName,
		query.Data,
	)

	// In a real bot, you would:
	// 1. Answer the callback query (to stop the loading animation)
	// 2. Process the callback data
	// 3. Update the message or send a new one

	fmt.Printf("Would process callback data: %s\n", query.Data)
}
