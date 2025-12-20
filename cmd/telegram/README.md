# Telegram Bot Example

This is a complete example demonstrating how to use the `telegram` package to build a Telegram bot.

## Features Demonstrated

- Creating a bot with configuration
- Receiving updates via channel
- Graceful shutdown with signal handling
- Processing text messages
- **Proper command detection using message entities** (not `strings.HasPrefix`)
- Handling bot commands (`/start`, `/help`, `/status`, `/echo`)
- Command argument parsing
- Handling callback queries (from inline keyboards)
- Concurrent update processing with goroutines
- Logging and error handling

## Prerequisites

1. Create a Telegram bot and get a token:
   - Open Telegram and search for `@BotFather`
   - Send `/newbot` and follow the instructions
   - Copy the bot token you receive

2. Set the bot token as an environment variable:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   ```

## Running the Example

From the repository root:

```bash
# Run directly
TELEGRAM_BOT_TOKEN="your_token" go run ./cmd/telegram

# Or build and run
go build -o telegram-bot ./cmd/telegram
TELEGRAM_BOT_TOKEN="your_token" ./telegram-bot
```

## Testing the Bot

1. Start the bot (see above)
2. Open Telegram and find your bot
3. Try these commands:
   - `/start` - Start the bot
   - `/help` - Get help information
   - `/status` - Check bot status
   - `/echo Hello World` - Echo a message
   - Send any regular text message

## What Happens

The example bot will:

1. Connect to Telegram using the Bot API
2. Start listening for updates on a channel
3. Log all received messages and commands
4. Process commands (currently just logs them)
5. Handle graceful shutdown on Ctrl+C

## Example Output

```
2025/11/30 12:00:00 Authorized on account YourBotName
2025/11/30 12:00:00 Bot started successfully. Press Ctrl+C to stop.
2025/11/30 12:00:05 [Message #1] From: John (@john_doe) | Chat ID: 123456789 | Text: /start
2025/11/30 12:00:05 User John started the bot
Welcome message would be sent to chat 123456789
2025/11/30 12:00:10 [Message #2] From: John (@john_doe) | Chat ID: 123456789 | Text: Hello bot!
2025/11/30 12:00:10 Received regular message from John: Hello bot!
2025/11/30 12:00:10 Message has 2 words
```

## Extending the Example

This is a read-only example that logs updates. To make it interactive, you would need to:

1. Add the Telegram Bot API client for sending messages
2. Implement message sending in the command handlers
3. Add state management for conversations
4. Integrate with databases, APIs, or other services

### Example: Sending Messages

To send messages, you'd need to use the underlying `telegram-bot-api` library directly:

```go
import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// Keep a reference to the API client
var api *tgbotapi.BotAPI

// In handleCommand:
case "/start":
    msg := tgbotapi.NewMessage(msg.Chat.ID, "Welcome to the bot!")
    api.Send(msg)
```

## Architecture Notes

The example demonstrates clean architecture:

- **Main loop**: Receives updates from channel
- **Handlers**: Process updates concurrently with goroutines
- **Command Detection**: Uses proper message entities (via `msg.IsCommand()`) instead of string prefix checking
- **Command Parsing**: Automatically handles @botname mentions and extracts arguments
- **Commands**: Separated from regular messages
- **Graceful shutdown**: Uses context and signal handling
- **Logging**: Comprehensive logging for debugging

### Why Proper Command Detection Matters

The example uses `msg.IsCommand()`, `msg.Command()`, and `msg.CommandArguments()` instead of checking `strings.HasPrefix(msg.Text, "/")`. This is the mainstream approach because:

- Telegram provides entity metadata that marks commands
- Handles `/command@botname` mentions correctly
- Only detects commands at the start of messages
- More reliable than string parsing

## Integration with Other Toolbox Packages

You can combine this with other packages:

```go
import (
    "github.com/denis-kilchichakov/toolbox/telegram"
    "github.com/denis-kilchichakov/toolbox/sqldb"
    "github.com/denis-kilchichakov/toolbox/llm"
    "github.com/denis-kilchichakov/toolbox/system"
)

// Use sqldb for storing user data
// Use llm for AI-powered responses
// Use system for signal handling
```

## Common Use Cases

1. **Notification Bot**: Receive updates and send notifications
2. **Command Bot**: Execute commands and return results
3. **Chat Bot**: Conversational AI integration
4. **Admin Bot**: Manage servers or services via Telegram
5. **Alert Bot**: Monitor systems and send alerts

## Security Considerations

- Never commit your bot token to version control
- Use environment variables for sensitive data
- Validate and sanitize all user input
- Implement rate limiting for commands
- Use allowlists for admin commands

## Troubleshooting

**Bot not receiving updates:**
- Check your bot token is correct
- Ensure the bot is not already running elsewhere
- Check network connectivity

**Permission denied:**
- Make sure TELEGRAM_BOT_TOKEN environment variable is set
- Verify the token is valid

**Build errors:**
- Run `go mod tidy` from repository root
- Ensure all dependencies are installed
