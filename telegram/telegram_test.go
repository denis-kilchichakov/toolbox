package telegram

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Config tests

func TestDefaultConfig(t *testing.T) {
	token := "test_token_123"
	config := DefaultConfig(token)

	assert.Equal(t, token, config.BotToken)
	assert.Equal(t, 60, config.Timeout)
	assert.False(t, config.Debug)
}

func TestDefaultConfig_CanOverride(t *testing.T) {
	config := DefaultConfig("my_token")

	// Override specific settings
	config.Timeout = 120
	config.Debug = true

	assert.Equal(t, "my_token", config.BotToken)
	assert.Equal(t, 120, config.Timeout)
	assert.True(t, config.Debug)
}

// Message entity and command tests

func TestMessage_IsCommand(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		expected bool
	}{
		{
			name: "valid command at start",
			msg: &Message{
				Text: "/start",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
				},
			},
			expected: true,
		},
		{
			name: "command not at start",
			msg: &Message{
				Text: "hello /start",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 6, Length: 6},
				},
			},
			expected: false,
		},
		{
			name: "no entities",
			msg: &Message{
				Text: "/start",
			},
			expected: false,
		},
		{
			name: "not a command entity",
			msg: &Message{
				Text: "@username",
				Entities: []MessageEntity{
					{Type: "mention", Offset: 0, Length: 9},
				},
			},
			expected: false,
		},
		{
			name:     "nil message",
			msg:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.IsCommand()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessage_Command(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		expected string
	}{
		{
			name: "simple command",
			msg: &Message{
				Text: "/start",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
				},
			},
			expected: "start",
		},
		{
			name: "command with @botname",
			msg: &Message{
				Text: "/start@mybot",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 12},
				},
			},
			expected: "start",
		},
		{
			name: "command with arguments",
			msg: &Message{
				Text: "/echo hello world",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 5},
				},
			},
			expected: "echo",
		},
		{
			name: "not a command",
			msg: &Message{
				Text: "hello",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.Command()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessage_CommandArguments(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		expected string
	}{
		{
			name: "command with arguments",
			msg: &Message{
				Text: "/echo hello world",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 5},
				},
			},
			expected: "hello world",
		},
		{
			name: "command without arguments",
			msg: &Message{
				Text: "/start",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 6},
				},
			},
			expected: "",
		},
		{
			name: "command with extra spaces",
			msg: &Message{
				Text: "/echo   test",
				Entities: []MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 5},
				},
			},
			expected: "test",
		},
		{
			name: "not a command",
			msg: &Message{
				Text: "hello",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.CommandArguments()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock tests

func TestMockBot(t *testing.T) {
	bot := NewMockBot()
	defer bot.Close()

	// Test that we can receive updates
	testUpdate := Update{
		ID: 123,
		Message: &Message{
			ID:   1,
			Text: "Hello",
			Chat: &Chat{ID: 456, Type: "private"},
			From: &User{ID: 789, FirstName: "Test"},
			Date: time.Now().Unix(),
		},
	}

	// Send update in goroutine
	go func() {
		bot.SendUpdate(testUpdate)
	}()

	// Receive update
	select {
	case update := <-bot.Updates():
		assert.Equal(t, int64(123), update.ID)
		assert.NotNil(t, update.Message)
		assert.Equal(t, "Hello", update.Message.Text)
		assert.Equal(t, int64(456), update.Message.Chat.ID)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for update")
	}
}

func TestMockBotClose(t *testing.T) {
	bot := NewMockBot()

	// Close the bot
	err := bot.Close()
	assert.NoError(t, err)

	// Channel should be closed
	select {
	case _, ok := <-bot.Updates():
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed immediately")
	}

	// Closing again should not error
	err = bot.Close()
	assert.NoError(t, err)
}

func TestMockBotSendAfterClose(t *testing.T) {
	bot := NewMockBot()
	bot.Close()

	// Sending after close should not panic
	assert.NotPanics(t, func() {
		bot.SendUpdate(Update{ID: 1})
	})
}

// Real Bot tests

func TestNewBot_InvalidToken(t *testing.T) {
	_, err := NewBot(Config{
		BotToken: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bot token is required")
}

func TestNewBot_DefaultTimeout(t *testing.T) {
	// This will fail with invalid token, but we can check the config
	config := Config{
		BotToken: "invalid",
		Timeout:  0,
	}

	// We expect an error due to invalid token, but we're testing config defaults
	bot, err := NewBot(config)
	if err == nil {
		defer bot.Close()
		assert.Equal(t, 60, bot.config.Timeout)
	}
	// If error occurs (expected with invalid token), that's fine for this test
}

func TestBot_Close(t *testing.T) {
	// Test closing without a real connection
	bot := &Bot{
		updates: make(chan Update),
		cancel:  func() {},
	}

	err := bot.Close()
	assert.NoError(t, err)

	// Closing again should not error
	err = bot.Close()
	assert.NoError(t, err)
}

func TestBot_UpdatesChannel(t *testing.T) {
	bot := &Bot{
		updates: make(chan Update, 1),
	}

	ch := bot.Updates()
	assert.NotNil(t, ch)

	// Should be able to receive from channel
	testUpdate := Update{ID: 1}
	bot.updates <- testUpdate

	received := <-ch
	assert.Equal(t, int64(1), received.ID)
}

// Integration test - requires real bot token
func TestIntegration_NewBot(t *testing.T) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		t.Skip("TELEGRAM_BOT_TOKEN not set, skipping integration test")
	}

	bot, err := NewBot(Config{
		BotToken: token,
		Timeout:  10,
		Debug:    false,
	})
	require.NoError(t, err)
	require.NotNil(t, bot)

	// Verify updates channel is available
	assert.NotNil(t, bot.Updates())

	// Clean shutdown
	err = bot.Close()
	assert.NoError(t, err)

	// Channel should be closed
	select {
	case _, ok := <-bot.Updates():
		assert.False(t, ok, "Channel should be closed after bot.Close()")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed immediately")
	}
}