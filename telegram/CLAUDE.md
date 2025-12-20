# telegram/CLAUDE.md

Package-specific guidance for working with the `telegram` package.

## Overview

The `telegram` package provides a clean, channel-based interface for building Telegram bots. It wraps `github.com/go-telegram-bot-api/telegram-bot-api` with:

- Custom types decoupled from the underlying library
- Channel-based update delivery (`<-chan Update`)
- Proper command detection using message entities
- Mock implementation for testing
- Graceful shutdown support

## Key Design Patterns

### 1. Interface-Based Design

The package defines `TelegramBot` interface to allow for different implementations:

```go
type TelegramBot interface {
    Updates() <-chan Update
    Close() error
}
```

Implementations:
- `Bot` - Real bot using Telegram Bot API
- `MockBot` - Test implementation

### 2. Custom Types

The package uses its own types (`Update`, `Message`, `User`, `Chat`, `MessageEntity`) instead of exposing `telegram-bot-api` types directly. This:

- Decouples user code from the underlying library
- Makes the API more stable
- Simplifies testing
- Allows for cleaner type definitions

Conversion happens in `convertUpdate()` function in `bot.go`.

### 3. Channel-Based Communication

Updates are delivered via a read-only channel (`<-chan Update`), which:

- Provides a clean, idiomatic Go API
- Enables easy integration with select statements
- Supports graceful shutdown patterns
- Makes concurrent processing straightforward

### 4. Command Detection Using Entities

**IMPORTANT:** Always use message entities for command detection, never `strings.HasPrefix(msg.Text, "/")`.

**Why:**
- Telegram provides entity metadata marking commands as type `bot_command`
- Handles `/command@botname` mentions correctly (removes @botname automatically)
- Only detects commands at the start of messages (offset 0)
- More reliable than string parsing

**Methods:**
- `msg.IsCommand()` - Returns true if message is a command
- `msg.Command()` - Returns command name without "/" and without "@botname"
- `msg.CommandArguments()` - Returns arguments after the command

**Example:**
```go
if msg.IsCommand() {
    cmd := msg.Command()           // "start" for "/start@mybot"
    args := msg.CommandArguments() // "hello" for "/echo hello"

    switch cmd {
    case "start":
        // Handle start command
    }
}
```

### 5. Configuration with Defaults

Use `DefaultConfig()` to avoid specifying default values:

```go
// Recommended
config := telegram.DefaultConfig(token)
config.Debug = true  // Override only what you need

// Not recommended
config := telegram.Config{
    BotToken: token,
    Timeout:  60,    // Repeating default value
    Debug:    false, // Repeating default value
}
```

## Testing Strategy

### Mock Tests (Fast, No External Dependencies)

Use `MockBot` for unit testing:

```go
func TestMyHandler(t *testing.T) {
    bot := telegram.NewMockBot()
    defer bot.Close()

    // Send test update
    go bot.SendUpdate(telegram.Update{
        Message: &telegram.Message{
            Text: "/start",
            Entities: []telegram.MessageEntity{
                {Type: "bot_command", Offset: 0, Length: 6},
            },
        },
    })

    // Test your handler
    update := <-bot.Updates()
    assert.True(t, update.Message.IsCommand())
    assert.Equal(t, "start", update.Message.Command())
}
```

### Integration Tests (Require Real Bot Token)

Integration tests should:
- Check for `TELEGRAM_BOT_TOKEN` environment variable
- Skip if not set
- Test actual Telegram API connection

```go
func TestIntegration_RealBot(t *testing.T) {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        t.Skip("TELEGRAM_BOT_TOKEN not set, skipping integration test")
    }

    config := telegram.DefaultConfig(token)
    bot, err := telegram.NewBot(config)
    require.NoError(t, err)
    defer bot.Close()

    // Test actual functionality
}
```

Run integration tests:
```bash
TELEGRAM_BOT_TOKEN=your_token go test ./telegram -run Integration -v
```

### Testing Command Detection

When testing commands, **always include entities**:

```go
// Correct
msg := &telegram.Message{
    Text: "/start",
    Entities: []telegram.MessageEntity{
        {Type: "bot_command", Offset: 0, Length: 6},
    },
}

// Wrong - will not be detected as command
msg := &telegram.Message{
    Text: "/start",
    // Missing entities!
}
```

## Implementation Patterns

### Basic Bot Structure

```go
func main() {
    config := telegram.DefaultConfig(os.Getenv("TELEGRAM_BOT_TOKEN"))
    bot, err := telegram.NewBot(config)
    if err != nil {
        log.Fatal(err)
    }
    defer bot.Close()

    for update := range bot.Updates() {
        go handleUpdate(update)  // Process concurrently
    }
}

func handleUpdate(update telegram.Update) {
    if update.Message != nil {
        handleMessage(update.Message)
    }
}

func handleMessage(msg *telegram.Message) {
    if msg.IsCommand() {
        handleCommand(msg)
    } else {
        handleText(msg)
    }
}
```

### Graceful Shutdown

```go
func main() {
    config := telegram.DefaultConfig(os.Getenv("TELEGRAM_BOT_TOKEN"))
    bot, err := telegram.NewBot(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cancel()
        bot.Close()
    }()

    // Process updates with context
    for {
        select {
        case <-ctx.Done():
            return
        case update, ok := <-bot.Updates():
            if !ok {
                return
            }
            go handleUpdate(update)
        }
    }
}
```

### Command Router Pattern

```go
type CommandHandler func(*telegram.Message)

var commands = map[string]CommandHandler{
    "start":  handleStart,
    "help":   handleHelp,
    "status": handleStatus,
}

func handleCommand(msg *telegram.Message) {
    cmd := msg.Command()

    if handler, ok := commands[cmd]; ok {
        handler(msg)
    } else {
        log.Printf("Unknown command: %s", cmd)
    }
}
```

## Common Pitfalls

### 1. Using String Prefix Instead of Entities

**Wrong:**
```go
if strings.HasPrefix(msg.Text, "/") {  // DON'T DO THIS
    // Won't handle /command@botname correctly
}
```

**Correct:**
```go
if msg.IsCommand() {  // DO THIS
    cmd := msg.Command()  // Automatically handles @botname
}
```

### 2. Forgetting to Close the Bot

Always ensure the bot is closed, either with `defer` or in shutdown logic:

```go
bot, err := telegram.NewBot(config)
if err != nil {
    log.Fatal(err)
}
defer bot.Close()  // Important!
```

### 3. Blocking the Main Loop

Process updates concurrently to avoid blocking:

```go
// Good - concurrent processing
for update := range bot.Updates() {
    go handleUpdate(update)
}

// Bad - blocks on each update
for update := range bot.Updates() {
    handleUpdate(update)  // Next update waits for this to finish
}
```

### 4. Not Handling Channel Closure

Check if channel is closed:

```go
update, ok := <-bot.Updates()
if !ok {
    // Channel closed, bot stopped
    return
}
```

### 5. Missing Entities in Tests

When creating test messages with commands, include entities:

```go
// Wrong - IsCommand() will return false
testMsg := &telegram.Message{
    Text: "/start",
}

// Correct - IsCommand() will return true
testMsg := &telegram.Message{
    Text: "/start",
    Entities: []telegram.MessageEntity{
        {Type: "bot_command", Offset: 0, Length: 6},
    },
}
```

## File Structure

```
telegram/
├── bot.go              # Bot implementation, DefaultConfig
├── types.go            # Custom types, command detection methods
├── mock.go             # MockBot for testing
├── telegram_test.go    # All tests
├── example_test.go     # Examples for godoc
├── README.md           # User documentation
└── CLAUDE.md           # This file (AI guidance)
```

## Type Conversion

The `convertUpdate()` function converts from `telegram-bot-api` types to our custom types.

**When adding new fields:**
1. Add to custom type in `types.go`
2. Update JSON tags
3. Add conversion in `convertUpdate()` in `bot.go`
4. Add tests in `telegram_test.go`

**Example:**
```go
// 1. Add to custom type (types.go)
type Message struct {
    // ... existing fields
    EditDate int64 `json:"edit_date,omitempty"`  // New field
}

// 2. Add conversion (bot.go)
if tgUpdate.Message != nil {
    update.Message = &Message{
        // ... existing conversions
        EditDate: int64(tgUpdate.Message.EditDate),  // Convert
    }
}

// 3. Add test
func TestConvertMessage_WithEditDate(t *testing.T) {
    // Test the conversion
}
```

## Integration with Other Toolbox Packages

### With sqldb (Storing User Data)

```go
import (
    "github.com/denis-kilchichakov/toolbox/telegram"
    "github.com/denis-kilchichakov/toolbox/sqldb"
)

// Store user chat IDs
func handleStart(msg *telegram.Message) {
    db.Exec("INSERT INTO users (chat_id, username) VALUES (?, ?)",
        msg.Chat.ID, msg.From.Username)
}
```

### With llm (AI Bot)

```go
import (
    "github.com/denis-kilchichakov/toolbox/telegram"
    "github.com/denis-kilchichakov/toolbox/llm"
)

func handleMessage(msg *telegram.Message) {
    // Use LLM to generate response
    response, err := model.Ask(msg.Text, llm.DefaultRequestOptions())
    // Send response back to user
}
```

### With system (Graceful Shutdown)

```go
import (
    "github.com/denis-kilchichakov/toolbox/telegram"
    "github.com/denis-kilchichakov/toolbox/system"
)

func main() {
    bot, _ := telegram.NewBot(telegram.DefaultConfig(token))

    system.HandleSignals(func() {
        log.Println("Shutting down bot...")
        bot.Close()
    })

    for update := range bot.Updates() {
        handleUpdate(update)
    }
}
```

## Running Tests

```bash
# All tests
go test ./telegram -v

# Mock tests only (fast)
go test ./telegram -v -run Mock

# Integration tests only (requires TELEGRAM_BOT_TOKEN)
TELEGRAM_BOT_TOKEN=your_token go test ./telegram -v -run Integration

# Test coverage
go test ./telegram -cover

# Specific test
go test ./telegram -v -run TestMessage_IsCommand
```

## Debugging

Enable debug logging to see Telegram API calls:

```go
config := telegram.DefaultConfig(token)
config.Debug = true  // Logs all API requests/responses

bot, err := telegram.NewBot(config)
```

## Security Considerations

1. **Never commit bot tokens** - Use environment variables
2. **Validate user input** - Don't trust message content
3. **Rate limiting** - Consider rate limiting command handlers
4. **Authorization** - Check user IDs for admin commands
5. **Sanitize output** - Escape special characters when sending responses

## Performance Notes

- **Channel buffer size**: Updates channel has buffer of 100 messages
- **Long polling timeout**: Default 60 seconds (configurable)
- **Concurrent processing**: Updates are processed in separate goroutines
- **Graceful shutdown**: Waits for polling goroutine to finish

## Further Reading

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [telegram-bot-api Go Library](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [Message Entities](https://core.telegram.org/bots/api#messageentity)
- Repository main CLAUDE.md for general patterns
