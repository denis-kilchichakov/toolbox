# telegram

Simple Telegram bot integration for Go applications.

## Features

- Clean channel-based API for receiving updates
- Built on top of the mainstream `github.com/go-telegram-bot-api/telegram-bot-api` library
- Custom types for type-safe message handling
- Mock implementation for testing
- Graceful shutdown support
- Configurable long polling timeout

## Installation

```bash
go get github.com/denis-kilchichakov/toolbox/telegram
```

## Usage

### Basic Example

```go
package main

import (
    "log"
    "github.com/denis-kilchichakov/toolbox/telegram"
)

func main() {
    // Create bot with default configuration (recommended)
    config := telegram.DefaultConfig("YOUR_BOT_TOKEN_FROM_BOTFATHER")

    // Override specific settings if needed
    // config.Debug = true
    // config.Timeout = 120

    bot, err := telegram.NewBot(config)
    if err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    // Listen for updates on the channel
    for update := range bot.Updates() {
        if update.Message != nil {
            log.Printf("Received message: %s from %s",
                update.Message.Text,
                update.Message.From.FirstName)
        }

        if update.CallbackQuery != nil {
            log.Printf("Received callback: %s from %s",
                update.CallbackQuery.Data,
                update.CallbackQuery.From.FirstName)
        }
    }
}
```

### Integration with Graceful Shutdown

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/denis-kilchichakov/toolbox/telegram"
)

func main() {
    config := telegram.DefaultConfig(os.Getenv("TELEGRAM_BOT_TOKEN"))
    bot, err := telegram.NewBot(config)
    if err != nil {
        log.Fatal(err)
    }

    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cancel()
        bot.Close()
    }()

    // Process updates
    for {
        select {
        case <-ctx.Done():
            return
        case update, ok := <-bot.Updates():
            if !ok {
                return
            }
            handleUpdate(update)
        }
    }
}

func handleUpdate(update telegram.Update) {
    // Your update handling logic
}
```

### Testing with MockBot

```go
package myapp

import (
    "testing"
    "time"

    "github.com/denis-kilchichakov/toolbox/telegram"
    "github.com/stretchr/testify/assert"
)

func TestMessageHandler(t *testing.T) {
    mock := telegram.NewMockBot()
    defer mock.Close()

    // Send a test update
    go func() {
        mock.SendUpdate(telegram.Update{
            ID: 1,
            Message: &telegram.Message{
                ID:   1,
                Text: "/start",
                Chat: &telegram.Chat{ID: 123, Type: "private"},
                From: &telegram.User{ID: 456, FirstName: "TestUser"},
            },
        })
    }()

    // Test your handler
    select {
    case update := <-mock.Updates():
        assert.Equal(t, "/start", update.Message.Text)
    case <-time.After(time.Second):
        t.Fatal("Timeout waiting for update")
    }
}
```

## Configuration

### Using DefaultConfig (Recommended)

The easiest way to configure the bot is using `DefaultConfig()`:

```go
// Use defaults (Timeout: 60s, Debug: false)
config := telegram.DefaultConfig("your_bot_token")
bot, err := telegram.NewBot(config)

// Or override specific settings
config := telegram.DefaultConfig("your_bot_token")
config.Debug = true
config.Timeout = 120
bot, err := telegram.NewBot(config)
```

### Manual Configuration

You can also create a `Config` struct manually:

```go
bot, err := telegram.NewBot(telegram.Config{
    BotToken: "your_bot_token",
    Timeout:  60,   // Long polling timeout in seconds
    Debug:    false, // Enable debug logging
})
```

### Config Options

- `BotToken` (required): Telegram bot token from @BotFather
- `Timeout` (optional): Long polling timeout in seconds (default: 60)
- `Debug` (optional): Enable debug logging (default: false)

## Types

The package defines its own types for Telegram updates:

- `Update`: Represents an incoming update
- `Message`: Represents a message with command detection helpers
- `MessageEntity`: Represents special entities in text (commands, mentions, URLs, etc.)
- `CallbackQuery`: Represents a callback query from inline keyboard
- `User`: Represents a Telegram user
- `Chat`: Represents a Telegram chat

These types are decoupled from the underlying `telegram-bot-api` library, making your code more maintainable and testable.

### Command Detection

The package provides proper command detection using Telegram's message entities (not simple string prefix checking):

```go
if msg.IsCommand() {
    command := msg.Command()           // Returns "start" for "/start" or "/start@botname"
    args := msg.CommandArguments()     // Returns arguments after the command

    switch command {
    case "start":
        // Handle /start command
    case "echo":
        // Handle /echo with args
    }
}
```

**Why use entities instead of `strings.HasPrefix(msg.Text, "/")`?**

1. **Accurate detection**: Telegram marks commands with entity type `bot_command`
2. **Handles @botname mentions**: `/start@mybot` is properly parsed as command "start"
3. **Position-aware**: Only detects commands at the start of messages
4. **Official API**: Uses Telegram's built-in metadata instead of string parsing

## Testing

Run tests with:

```bash
# Run all tests (mock tests only)
go test ./telegram

# Run with integration tests (requires TELEGRAM_BOT_TOKEN)
TELEGRAM_BOT_TOKEN=your_token_here go test ./telegram -v
```

## Architecture

The package follows the toolbox repository patterns:

- Interface-based design (`TelegramBot` interface)
- Channel-based communication for updates
- Mock implementation for testing
- Graceful shutdown support with context
- Comprehensive test coverage

## License

Part of the [toolbox](https://github.com/denis-kilchichakov/toolbox) utility library.
