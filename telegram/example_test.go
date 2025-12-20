package telegram_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/denis-kilchichakov/toolbox/telegram"
)

// Example demonstrates basic usage of the telegram bot
func Example() {
	// Create bot with default configuration
	config := telegram.DefaultConfig("YOUR_BOT_TOKEN_FROM_BOTFATHER")

	// Override specific settings if needed
	config.Debug = false

	bot, err := telegram.NewBot(config)
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()

	// Listen for updates
	for update := range bot.Updates() {
		if update.Message != nil {
			fmt.Printf("Message from %s: %s\n",
				update.Message.From.FirstName,
				update.Message.Text)
		}
	}
}

// Example_gracefulShutdown demonstrates graceful shutdown
func Example_gracefulShutdown() {
	config := telegram.DefaultConfig(os.Getenv("TELEGRAM_BOT_TOKEN"))
	bot, err := telegram.NewBot(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
		bot.Close()
	}()

	// Process updates
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown complete")
			return
		case update, ok := <-bot.Updates():
			if !ok {
				return
			}
			if update.Message != nil {
				log.Printf("Received: %s", update.Message.Text)
			}
		}
	}
}

// Example_mockBot demonstrates testing with MockBot
func Example_mockBot() {
	mock := telegram.NewMockBot()
	defer mock.Close()

	// Simulate sending an update
	go func() {
		mock.SendUpdate(telegram.Update{
			ID: 1,
			Message: &telegram.Message{
				ID:   1,
				Text: "Hello, bot!",
				From: &telegram.User{
					ID:        123,
					FirstName: "TestUser",
				},
				Chat: &telegram.Chat{
					ID:   456,
					Type: "private",
				},
			},
		})
	}()

	// Receive and process the update
	update := <-mock.Updates()
	fmt.Printf("Received: %s\n", update.Message.Text)
	// Output: Received: Hello, bot!
}
