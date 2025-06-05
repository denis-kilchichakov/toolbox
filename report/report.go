package report

import (
	"context"
	"log"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
)

var (
	_defaultTitle  string
	_notifyService notify.Notifier
)

func setupWithService(service notify.Notifier, defaultTitle string) {
	_notifyService = service
	_defaultTitle = defaultTitle
}

func Setup(telegramApiToken string, telegramReceivers []int64, defaultTitle string) error {
	telegramService, err := telegram.New(telegramApiToken)
	if err != nil {
		return err
	}
	telegramService.AddReceivers(telegramReceivers...)
	setupWithService(telegramService, defaultTitle)
	return nil
}

func Report(title string, message string) {
	if title == "" {
		title = _defaultTitle
	}
	if _notifyService == nil {
		log.Println("Notification service is not initialized")
		return
	}
	err := _notifyService.Send(
		context.Background(),
		title,
		replaceAngleBrackets(message),
	)
	if err != nil {
		log.Println(err)
	}
}

func replaceAngleBrackets(input string) string {
	result := make([]rune, len(input))

	for i, char := range input {
		switch char {
		case '<':
			result[i] = '|'
		case '>':
			result[i] = '|'
		default:
			result[i] = char
		}
	}

	return string(result)
}
