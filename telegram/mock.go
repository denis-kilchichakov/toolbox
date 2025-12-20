package telegram

import "sync"

// MockBot implements TelegramBot for testing
type MockBot struct {
	updates chan Update
	closed  bool
	mu      sync.RWMutex
}

// NewMockBot creates a new mock bot for testing
func NewMockBot() *MockBot {
	return &MockBot{
		updates: make(chan Update, 10), // buffered channel for tests
	}
}

// Updates returns the mock updates channel
func (m *MockBot) Updates() <-chan Update {
	return m.updates
}

// Close closes the mock bot
func (m *MockBot) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.closed {
		close(m.updates)
		m.closed = true
	}
	return nil
}

// SendUpdate sends a mock update (for testing)
func (m *MockBot) SendUpdate(update Update) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if !m.closed {
		m.updates <- update
	}
}