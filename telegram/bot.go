package telegram

import (
	"context"
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TelegramBot defines the interface for receiving updates from Telegram
type TelegramBot interface {
	// Updates returns a channel that receives incoming updates
	Updates() <-chan Update

	// Close stops the bot and closes the updates channel
	Close() error
}

// Config holds the configuration for the Telegram bot
type Config struct {
	// BotToken is the Telegram bot token obtained from @BotFather
	BotToken string

	// Timeout is the timeout for long polling in seconds (default: 60)
	Timeout int

	// Debug enables debug logging (default: false)
	Debug bool
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig(botToken string) Config {
	return Config{
		BotToken: botToken,
		Timeout:  60,
		Debug:    false,
	}
}

// Bot implements TelegramBot using the Telegram Bot API
type Bot struct {
	api     *tgbotapi.BotAPI
	updates chan Update
	config  Config
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	closed  bool
}

// NewBot creates a new Telegram bot with the given configuration
func NewBot(config Config) (*Bot, error) {
	if config.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	if config.Timeout == 0 {
		config.Timeout = 60
	}

	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	api.Debug = config.Debug

	if config.Debug {
		log.Printf("Authorized on account %s", api.Self.UserName)
	}

	ctx, cancel := context.WithCancel(context.Background())

	bot := &Bot{
		api:     api,
		updates: make(chan Update, 100),
		config:  config,
		cancel:  cancel,
	}

	bot.wg.Add(1)
	go bot.pollUpdates(ctx)

	return bot, nil
}

// Updates returns the channel that receives incoming updates
func (b *Bot) Updates() <-chan Update {
	return b.updates
}

// Close stops the bot and closes the updates channel
func (b *Bot) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()

	b.cancel()
	b.wg.Wait()
	close(b.updates)

	return nil
}

// pollUpdates continuously polls for updates from Telegram
func (b *Bot) pollUpdates(ctx context.Context) {
	defer b.wg.Done()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.config.Timeout

	updatesChan, err := b.api.GetUpdatesChan(u)
	if err != nil {
		if b.config.Debug {
			log.Printf("Error getting updates channel: %v", err)
		}
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case tgUpdate, ok := <-updatesChan:
			if !ok {
				return
			}

			update := convertUpdate(tgUpdate)

			select {
			case b.updates <- update:
			case <-ctx.Done():
				return
			}
		}
	}
}

// convertUpdate converts a telegram-bot-api Update to our custom Update type
func convertUpdate(tgUpdate tgbotapi.Update) Update {
	update := Update{
		ID: int64(tgUpdate.UpdateID),
	}

	if tgUpdate.Message != nil {
		update.Message = &Message{
			ID:   tgUpdate.Message.MessageID,
			Text: tgUpdate.Message.Text,
			Date: int64(tgUpdate.Message.Date),
		}

		if tgUpdate.Message.From != nil {
			update.Message.From = &User{
				ID:        int64(tgUpdate.Message.From.ID),
				FirstName: tgUpdate.Message.From.FirstName,
				Username:  tgUpdate.Message.From.UserName,
			}
		}

		if tgUpdate.Message.Chat != nil {
			update.Message.Chat = &Chat{
				ID:   tgUpdate.Message.Chat.ID,
				Type: tgUpdate.Message.Chat.Type,
			}
		}

		// Convert entities
		if tgUpdate.Message.Entities != nil && len(*tgUpdate.Message.Entities) > 0 {
			entities := *tgUpdate.Message.Entities
			update.Message.Entities = make([]MessageEntity, len(entities))
			for i, entity := range entities {
				update.Message.Entities[i] = MessageEntity{
					Type:   entity.Type,
					Offset: entity.Offset,
					Length: entity.Length,
				}
			}
		}
	}

	if tgUpdate.CallbackQuery != nil {
		update.CallbackQuery = &CallbackQuery{
			ID:   tgUpdate.CallbackQuery.ID,
			Data: tgUpdate.CallbackQuery.Data,
		}

		if tgUpdate.CallbackQuery.From != nil {
			update.CallbackQuery.From = &User{
				ID:        int64(tgUpdate.CallbackQuery.From.ID),
				FirstName: tgUpdate.CallbackQuery.From.FirstName,
				Username:  tgUpdate.CallbackQuery.From.UserName,
			}
		}

		if tgUpdate.CallbackQuery.Message != nil {
			update.CallbackQuery.Message = &Message{
				ID:   tgUpdate.CallbackQuery.Message.MessageID,
				Text: tgUpdate.CallbackQuery.Message.Text,
				Date: int64(tgUpdate.CallbackQuery.Message.Date),
			}

			if tgUpdate.CallbackQuery.Message.Chat != nil {
				update.CallbackQuery.Message.Chat = &Chat{
					ID:   tgUpdate.CallbackQuery.Message.Chat.ID,
					Type: tgUpdate.CallbackQuery.Message.Chat.Type,
				}
			}
		}
	}

	return update
}