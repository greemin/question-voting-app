package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"question-voting-app/internal/handlers"
	"question-voting-app/internal/models"
	"question-voting-app/internal/testutil"
)

// --- Test Setup Helper ---

// setupTestAPI creates the handler instance with an injected MockStorer.
func setupTestAPI() (*handlers.API, *testutil.MockStorer) {
	storer := testutil.NewMockStorer()
	api := handlers.New(storer)
	return api, storer
}

// createMockSession creates a SessionData object for testing.
func createMockSession(id string, adminID string, isActive bool) *models.SessionData {
	return &models.SessionData{
		SessionID:   id,
		AdminUserID: adminID,
		IsActive:    isActive,
		Questions: []models.Question{
			{ID: "q1", Text: "Question One (10 votes)", Votes: 10, Voters: []string{"u1", "u2"}},
			{ID: "q2", Text: "Question Two (5 votes)", Votes: 5, Voters: []string{"u3"}},
		},
	}
}

// --- Handler Unit Tests ---

func TestCreateSessionHandler(t *testing.T) {
	api, _ := setupTestAPI()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/session", nil)

	api.CreateSessionHandler(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if _, ok := response["sessionId"]; !ok {
		t.Error("Response missing 'sessionId'")
	}
	if _, ok := response["adminId"]; !ok {
		t.Error("Response missing 'adminId'")
	}
}

func TestGetQuestionsHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "test-get-q"
	adminID := "admin-1"

	// Setup: Save a mock session with sorted questions (10 votes, then 5 votes)
	storer.SaveSessionData(createMockSession(sessionID, adminID, true))

	// Test Case 1: Successful Retrieval and Sorting
	t.Run("Success_Sorted", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/questions", nil)

		api.GetQuestionsHandler(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var questions []models.Question
		if err := json.Unmarshal(w.Body.Bytes(), &questions); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Assert Sorting (Highest vote first)
		if len(questions) != 2 {
			t.Fatalf("Expected 2 questions, got %d", len(questions))
		}
		if questions[0].Votes < questions[1].Votes {
			t.Errorf("Questions were not sorted correctly. Q1 Votes: %d, Q2 Votes: %d", questions[0].Votes, questions[1].Votes)
		}
	})

	// Test Case 2: Session Not Found
	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/non-existent/questions", nil)

		api.GetQuestionsHandler(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestSubmitQuestionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "test-submit-q"
	adminID := "admin-2"
	storer.SaveSessionData(createMockSession(sessionID, adminID, true)) // Active session

	tests := []struct {
		name           string
		body           *models.QuestionSubmission
		sessionID      string
		isActive       bool
		expectedStatus int
	}{
		{"Success", &models.QuestionSubmission{Text: "New Test Question"}, sessionID, true, http.StatusCreated},
		{"EmptyBody", &models.QuestionSubmission{Text: ""}, sessionID, true, http.StatusBadRequest},
		{"SessionNotFound", &models.QuestionSubmission{Text: "Valid Q"}, "non-existent", true, http.StatusNotFound},
		{"SessionClosed", &models.QuestionSubmission{Text: "Closed Q"}, sessionID, false, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Ensure session state matches test case (e.g., set to inactive if needed)
			if tt.isActive == false {
				closedSession := createMockSession(sessionID, adminID, false)
				storer.SaveSessionData(closedSession)
			}

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/session/"+tt.sessionID+"/questions", bytes.NewReader(bodyBytes))
			r.AddCookie(&http.Cookie{Name: "userSessionId", Value: "test-user"})

			api.SubmitQuestionHandler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Post-assertion: Revert session back to active for other tests
			if tt.isActive == false {
				storer.SaveSessionData(createMockSession(sessionID, adminID, true))
			}
		})
	}
}

func TestVoteQuestionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "test-vote-q"
	adminID := "admin-3"

	// Setup: Save a mock session (active) where q1 has 10 votes, 2 voters
	storer.SaveSessionData(createMockSession(sessionID, adminID, true))

	validPath := fmt.Sprintf("/api/session/%s/questions/q1/vote", sessionID)
	invalidPath := fmt.Sprintf("/api/session/%s/questions/q99/vote", sessionID)

	tests := []struct {
		name           string
		path           string
		userID         string
		isActive       bool
		expectedStatus int
	}{
		{"Success_FirstVote", validPath, "new-voter-4", true, http.StatusOK},
		{"AlreadyVoted", validPath, "u1", true, http.StatusForbidden}, // u1 is a voter on q1 from setup
		{"QuestionNotFound", invalidPath, "new-voter-5", true, http.StatusNotFound},
		{"SessionClosed", validPath, "new-voter-6", false, http.StatusForbidden},
		{"SessionNotFound", "/api/session/non-existent/questions/q1/vote", "new-voter-7", true, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Set session state for test case
			if !tt.isActive {
				storer.SaveSessionData(createMockSession(sessionID, adminID, false))
			} else if tt.name == "Success_FirstVote" {
				// Ensure fresh state for the first successful vote
				storer.SaveSessionData(createMockSession(sessionID, adminID, true))
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPut, tt.path, nil)
			r.AddCookie(&http.Cookie{Name: "userSessionId", Value: tt.userID})

			api.VoteQuestionHandler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestEndSessionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "test-end-session"
	adminID := "admin-4"
	otherUser := "user-4"

	// Setup: Save the session
	storer.SaveSessionData(createMockSession(sessionID, adminID, true))

	tests := []struct {
		name           string
		userID         string
		sessionExists  bool
		expectedStatus int
	}{
		{"Success_Admin", adminID, true, http.StatusNoContent},
		{"Unauthorized_User", otherUser, true, http.StatusForbidden},
		{"SessionNotFound", adminID, false, http.StatusNoContent}, // Handler returns NoContent if file isn't found
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Ensure session exists or doesn't exist before test
			if tt.sessionExists {
				// Re-save if it was deleted in a previous test
				storer.SaveSessionData(createMockSession(sessionID, adminID, true))
			} else {
				storer.DeleteSessionData(sessionID)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/api/session/"+sessionID, nil)
			r.AddCookie(&http.Cookie{Name: "userSessionId", Value: tt.userID})

			api.EndSessionHandler(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Post-Assertion Check: If status was 204, verify deletion in mock
			if tt.expectedStatus == http.StatusNoContent && tt.userID == adminID && tt.sessionExists {
				_, err := storer.LoadSessionData(sessionID)
				if err == nil {
					t.Error("Expected session to be deleted from storage, but it was found")
				}
			}
		})
	}
}

func TestCheckAdminHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "test-check-admin"
	adminID := "admin-5"
	otherUser := "user-5"

	// Setup: Save the session
	storer.SaveSessionData(createMockSession(sessionID, adminID, true))

	tests := []struct {
		name                string
		userID              string
		sessionExists       bool
		expectedAdminStatus bool
	}{
		{"IsAdmin", adminID, true, true},
		{"IsNotAdmin", otherUser, true, false},
		{"SessionNotFound", adminID, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Ensure session state matches test case
			if !tt.sessionExists {
				storer.DeleteSessionData(sessionID)
			} else {
				storer.SaveSessionData(createMockSession(sessionID, adminID, true))
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/check-admin", nil)
			r.AddCookie(&http.Cookie{Name: "userSessionId", Value: tt.userID})

			api.CheckAdminHandler(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			}

			var response map[string]bool
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["isAdmin"] != tt.expectedAdminStatus {
				t.Errorf("Expected isAdmin=%t, got isAdmin=%t", tt.expectedAdminStatus, response["isAdmin"])
			}
		})
	}
}
