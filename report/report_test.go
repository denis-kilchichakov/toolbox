package report

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/nikoksr/notify"
	"github.com/stretchr/testify/assert"
)

type MockNotifier struct {
	notify.Notifier
	SendFunc func(ctx context.Context, subject, message string) error
}

func (m *MockNotifier) Send(ctx context.Context, subject, message string) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, subject, message)
	}
	return nil
}

func TestSetup(t *testing.T) {
	mockService := &MockNotifier{}
	setupWithService(mockService, "Default Title")

	assert.Equal(t, "Default Title", _defaultTitle, "default title should be set correctly")
}

func TestSetup_InvalidToken(t *testing.T) {
	telegramApiToken := ""
	telegramReceivers := []int64{123456789}
	defaultTitle := "Default Title"

	err := Setup(telegramApiToken, telegramReceivers, defaultTitle)
	assert.Error(t, err, "expected error with invalid token")
}

func TestSetup_EmptyReceivers(t *testing.T) {
	mockService := &MockNotifier{}
	setupWithService(mockService, "Default Title")

	assert.Equal(t, "Default Title", _defaultTitle, "default title should be set correctly")
}

func TestReport(t *testing.T) {
	// Test case: Report with a custom title and message
	mockService := &MockNotifier{
		SendFunc: func(ctx context.Context, subject, message string) error {
			assert.Equal(t, "Custom Title", subject, "title should match")
			assert.Equal(t, "Custom Message", message, "message should match")
			return nil
		},
	}
	setupWithService(mockService, "Default Title")

	Report("Custom Title", "Custom Message")
}

func TestReport_EmptyTitle(t *testing.T) {
	// Test case: Report with an empty title, should use default title
	mockService := &MockNotifier{
		SendFunc: func(ctx context.Context, subject, message string) error {
			assert.Equal(t, "Default Title", subject, "default title should be used")
			assert.Equal(t, "Message with |angle brackets|", message, "message should match with replaced angle brackets")
			return nil
		},
	}
	setupWithService(mockService, "Default Title")

	Report("", "Message with <angle brackets>")
}

func TestReport_UninitializedService(t *testing.T) {
	// Test case: Report when notification service is not initialized
	_notifyService = nil // Ensure service is uninitialized

	// Capture log output
	logOutput := &bytes.Buffer{}
	log.SetOutput(logOutput)
	defer log.SetOutput(os.Stderr) // Restore default output

	Report("Title", "Message")

	assert.Contains(t, logOutput.String(), "Notification service is not initialized", "should log uninitialized service error")
}
