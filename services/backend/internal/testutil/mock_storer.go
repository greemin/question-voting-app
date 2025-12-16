// internal/testutil/mock_storer.go
package testutil

import (
	"fmt"
	"question-voting-app/internal/models"
)

// MockStorer simulates the storage layer using an in-memory map.
// It satisfies the storage.Storer interface for unit testing.
type MockStorer struct {
	sessions map[string]*models.SessionData
}

// NewMockStorer creates and returns a new in-memory Storer implementation.
func NewMockStorer() *MockStorer {
	return &MockStorer{
		sessions: make(map[string]*models.SessionData),
	}
}

// LoadSessionData implements the Storer interface by reading from the map.
func (ms *MockStorer) LoadSessionData(sessionID string) (*models.SessionData, error) {
	data, exists := ms.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	// Return a copy to prevent test side effects
	dataCopy := *data
	return &dataCopy, nil
}

// SaveSessionData implements the Storer interface by writing to the map.
func (ms *MockStorer) SaveSessionData(data *models.SessionData) error {
	ms.sessions[data.SessionID] = data
	return nil
}

// DeleteSessionData implements the Storer interface by removing the entry from the map.
func (ms *MockStorer) DeleteSessionData(sessionID string) error {
	delete(ms.sessions, sessionID)
	return nil
}
