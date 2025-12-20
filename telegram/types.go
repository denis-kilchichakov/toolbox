package telegram

import "strings"

// Update represents an incoming update from Telegram
type Update struct {
	ID            int64          `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// Message represents a message from Telegram
type Message struct {
	ID       int              `json:"message_id"`
	From     *User            `json:"from,omitempty"`
	Chat     *Chat            `json:"chat"`
	Date     int64            `json:"date"`
	Text     string           `json:"text,omitempty"`
	Entities []MessageEntity  `json:"entities,omitempty"`
}

// MessageEntity represents a special entity in a text message (e.g., commands, mentions, URLs)
type MessageEntity struct {
	Type   string `json:"type"`   // Type of the entity (bot_command, mention, url, etc.)
	Offset int    `json:"offset"` // Offset in UTF-16 code units to the start of the entity
	Length int    `json:"length"` // Length of the entity in UTF-16 code units
}

// IsCommand returns true if the message is a bot command
func (m *Message) IsCommand() bool {
	if m == nil || len(m.Entities) == 0 {
		return false
	}
	// Command must start at the beginning of the message
	return m.Entities[0].Type == "bot_command" && m.Entities[0].Offset == 0
}

// Command returns the command without the leading slash, or empty string if not a command
func (m *Message) Command() string {
	if !m.IsCommand() {
		return ""
	}

	// Extract command text using the entity length
	entity := m.Entities[0]
	if entity.Length > len(m.Text) {
		return ""
	}

	command := m.Text[:entity.Length]

	// Remove leading slash
	if len(command) > 0 && command[0] == '/' {
		command = command[1:]
	}

	// Remove @botname suffix if present
	if idx := strings.Index(command, "@"); idx != -1 {
		command = command[:idx]
	}

	return command
}

// CommandArguments returns the text after the command, or empty string if not a command
func (m *Message) CommandArguments() string {
	if !m.IsCommand() {
		return ""
	}

	entity := m.Entities[0]
	// Skip the command and any whitespace after it
	if entity.Length >= len(m.Text) {
		return ""
	}

	args := m.Text[entity.Length:]
	return strings.TrimSpace(args)
}

// CallbackQuery represents an incoming callback query from inline keyboard
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

// User represents a Telegram user
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}