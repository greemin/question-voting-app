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

// ConfigureIndexes is a mock implementation that does nothing.
func (ms *MockStorer) ConfigureIndexes() error {
	return nil
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

// CreateSessionData simulates creating a document, returning a duplicate key error if it exists.
func (ms *MockStorer) CreateSessionData(data *models.SessionData) error {
	if _, exists := ms.sessions[data.SessionID]; exists {
		// This string is what mongo.IsDuplicateKeyError() checks for.
		return fmt.Errorf("E11000 duplicate key error collection")
	}
	ms.sessions[data.SessionID] = data
	return nil
}

// UpdateSessionData implements the Storer interface by writing to the map.
func (ms *MockStorer) UpdateSessionData(data *models.SessionData) error {
	if _, exists := ms.sessions[data.SessionID]; !exists {
		return fmt.Errorf("session not found, cannot update: %s", data.SessionID)
	}
	ms.sessions[data.SessionID] = data
	return nil
}

// DeleteSessionData implements the Storer interface by removing the entry from the map.
func (ms *MockStorer) DeleteSessionData(sessionID string) error {
	if _, exists := ms.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found, cannot delete: %s", sessionID)
	}
	delete(ms.sessions, sessionID)
	return nil
}

// GetSessionIDs is a test helper to inspect the mock storer's state.
func (ms *MockStorer) GetSessionIDs() []string {
	keys := make([]string, 0, len(ms.sessions))
	for k := range ms.sessions {
		keys = append(keys, k)
	}
	return keys
}

// Clear is a test helper to reset the storer state between tests.
func (ms *MockStorer) Clear() {
	ms.sessions = make(map[string]*models.SessionData)
}

// FindSessionByAdminID is a helper for finding a session for testing purposes.
func (ms *MockStorer) FindSessionByAdminID(adminID string) *models.SessionData {
	for _, s := range ms.sessions {
		if s.AdminUserID == adminID {
			// Return a copy
			sCopy := *s
			return &sCopy
		}
	}
	return nil
}

// IsEmpty checks if the mock storer has any sessions.
func (ms *MockStorer) IsEmpty() bool {
	return len(ms.sessions) == 0
}

// HasSession checks if a session exists by its ID.
func (ms *MockStorer) HasSession(sessionID string) bool {
	_, exists := ms.sessions[sessionID]
	return exists
}

// PreloadSession is a helper to directly add a session for test setup.
func (ms *MockStorer) PreloadSession(data *models.SessionData) {
	ms.sessions[data.SessionID] = data
}
