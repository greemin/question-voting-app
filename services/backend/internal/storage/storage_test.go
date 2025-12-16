package storage_test

import (
	"strings"
	"testing"

	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
	"question-voting-app/internal/testutil"
)

// --- Test Setup and Helpers ---

// setupStorer returns the Storer interface implemented by the MockStorer.
func setupStorer() storage.Storer {
	return testutil.NewMockStorer()
}

// createMockSession creates a SessionData object for testing, using a slice of Questions.
func createMockSession(id string) *models.SessionData {
	return &models.SessionData{
		SessionID:   id,
		AdminUserID: id + "-admin",
		IsActive:    true,
		Questions: []models.Question{ // 👈 CORRECTED to use SLICE
			{ID: "q1", Text: "Question One?", Votes: 5, Voters: []string{"user-a"}},
		},
	}
}

// --- Interface Tests ---

// TestStorerSaveAndLoad verifies successful creation and retrieval of session data.
func TestStorerSaveAndLoad(t *testing.T) {
	storer := setupStorer()

	tests := []struct {
		name      string
		sessionID string
	}{
		{name: "First Session", sessionID: "test-save-load-1"},
		{name: "Second Session", sessionID: "test-save-load-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange: Create the session data
			expected := createMockSession(tt.sessionID)

			// 2. Act: Save the data
			if err := storer.SaveSessionData(expected); err != nil {
				t.Fatalf("SaveSessionData failed: %v", err)
			}

			// 3. Act: Load the data
			actual, err := storer.LoadSessionData(tt.sessionID)

			// 4. Assert: Check for errors and data integrity
			if err != nil {
				t.Fatalf("LoadSessionData failed: %v", err)
			}
			if actual.SessionID != expected.SessionID {
				t.Errorf("Loaded ID mismatch. Got %s, Expected %s", actual.SessionID, expected.SessionID)
			}
			if len(actual.Questions) != 1 { // Assertions changed for slice length
				t.Errorf("Loaded questions count mismatch. Got %d, Expected 1", len(actual.Questions))
			}
			if actual.Questions[0].ID != expected.Questions[0].ID {
				t.Errorf("Loaded question ID mismatch. Got %s, Expected %s", actual.Questions[0].ID, expected.Questions[0].ID)
			}
		})
	}
}

// TestStorerLoadNotFound verifies that loading a non-existent session returns the expected error.
func TestStorerLoadNotFound(t *testing.T) {
	storer := setupStorer()
	_, err := storer.LoadSessionData("non-existent-id")

	if err == nil {
		t.Fatal("Expected an error for non-existent session, got nil")
	}

	expectedSubstring := "session not found"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message mismatch. Expected to contain '%s', got '%s'", expectedSubstring, err.Error())
	}
}

// TestStorerDelete verifies that data can be deleted and subsequently cannot be loaded.
func TestStorerDelete(t *testing.T) {
	storer := setupStorer()
	sessionID := "test-session-to-delete"

	// 1. Arrange: Create and save a session
	sessionData := createMockSession(sessionID)
	if err := storer.SaveSessionData(sessionData); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 2. Act: Delete the session data
	if err := storer.DeleteSessionData(sessionID); err != nil {
		t.Fatalf("DeleteSessionData failed: %v", err)
	}

	// 3. Assert: Try to load the session; it should return a "not found" error
	_, err := storer.LoadSessionData(sessionID)
	if err == nil {
		t.Fatal("Expected an error after deletion, got nil (session still loaded)")
	}

	// 4. Act/Assert: Deleting a non-existent session should not return an error
	if err := storer.DeleteSessionData("already-deleted-id"); err != nil {
		t.Errorf("Deleting a non-existent session returned an error: %v", err)
	}
}
