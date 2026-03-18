package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"question-voting-app/internal/models"
	"question-voting-app/internal/testutil"
)

// --- Test Setup Helper ---

// setupTestAPI creates the handler instance with an injected MockStorer.
func setupTestAPI() (*API, *testutil.MockStorer) {
	storer := testutil.NewMockStorer()
	api := New(storer, false)
	return api, storer
}

// createMockSession creates a SessionData object for testing.
func createMockSession(id string, adminID string, isActive bool) *models.SessionData {
	return &models.SessionData{
		SessionID:   id,
		AdminUserID: adminID,
		IsActive:    isActive,
		Questions: []models.Question{
			{ID: "00000000-0000-0000-0000-000000000011", Text: "Question One (10 votes)", Votes: 10, Voters: []string{"u1", "u2"}},
			{ID: "00000000-0000-0000-0000-000000000012", Text: "Question Two (5 votes)", Votes: 5, Voters: []string{"u3"}},
		},
	}
}

// --- Unit Tests ---

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal", "Hello World", "hello-world"},
		{"With &", "Q&A Session", "q-and-a-session"},
		{"With +", "Intro+Advanced", "intro-plus-advanced"},
		{"Underscores", "my_cool_session", "my-cool-session"},
		{"Extra Spaces", "  leading and trailing  ", "leading-and-trailing"},
		{"Mixed Garbage", "!@#$My%^&*()_Session 123+", "my-and-session-123-plus"},
		{"Uppercase", "ALL CAPS", "all-caps"},
		{"Empty after slugify", "!@#$%-", ""},
		{"No change", "already-a-slug", "already-a-slug"},
		{"Unicode CJK", "こんにちは World", "こんにちは-world"},
		{"Unicode Cyrillic", "Привет Мир", "привет-мир"},
		{"Unicode Accents", "Café au Lait", "café-au-lait"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := slugify(tt.input)
			if actual != tt.expected {
				t.Errorf("slugify(%q): expected %q, got %q", tt.input, tt.expected, actual)
			}
		})
	}
}

// --- Handler Unit Tests ---

func TestCreateSessionHandler(t *testing.T) {
	api, storer := setupTestAPI()

	t.Run("NoSlugProvided", func(t *testing.T) {
		storer.Clear()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session", nil)

		api.CreateSessionHandler(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		sessionID := resp["sessionId"]
		if len(sessionID) != 8 {
			t.Errorf("Expected a random 8-character slug, got %q", sessionID)
		}
		if !storer.HasSession(sessionID) {
			t.Error("Session was not created in the storer")
		}
	})

	t.Run("ValidSlugProvided", func(t *testing.T) {
		storer.Clear()
		slug := "my-cool-event"
		body := fmt.Sprintf(`{"sessionId": "%s"}`, slug)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(body))

		api.CreateSessionHandler(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["sessionId"] != slug {
			t.Errorf("Expected sessionId %q, got %q", slug, resp["sessionId"])
		}
		if !storer.HasSession(slug) {
			t.Errorf("Session with slug %q was not created", slug)
		}
	})

	t.Run("SlugIsSlugified", func(t *testing.T) {
		storer.Clear()
		inputSlug := "My Awesome Q&A"
		expectedSlug := "my-awesome-q-and-a"
		body := fmt.Sprintf(`{"sessionId": "%s"}`, inputSlug)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(body))

		api.CreateSessionHandler(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["sessionId"] != expectedSlug {
			t.Errorf("Expected slugified sessionId %q, got %q", expectedSlug, resp["sessionId"])
		}
		if !storer.HasSession(expectedSlug) {
			t.Errorf("Session with slugified slug %q was not created", expectedSlug)
		}
	})

	t.Run("SlugCollision", func(t *testing.T) {
		storer.Clear()
		collidingSlug := "i-exist"
		storer.PreloadSession(&models.SessionData{SessionID: collidingSlug})

		body := fmt.Sprintf(`{"sessionId": "%s"}`, collidingSlug)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(body))

		api.CreateSessionHandler(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d and message \"%s\"", http.StatusCreated, w.Code, w.Body.String())
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		newSessionID := resp["sessionId"]

		if newSessionID == collidingSlug {
			t.Fatalf("Session ID was not changed after collision")
		}
		if !strings.HasPrefix(newSessionID, collidingSlug+"-") {
			t.Errorf("Expected new session ID to be prefixed with %q, but got %q", collidingSlug+"-", newSessionID)
		}
		if len(newSessionID) != len(collidingSlug)+1+4 { // slug + hyphen + 4 chars
			t.Errorf("Expected new session ID to have a 4-char suffix, but got %q", newSessionID)
		}
		if !storer.HasSession(newSessionID) {
			t.Errorf("The new suffixed session %q was not created in the storer", newSessionID)
		}
	})

	t.Run("SlugTooLong", func(t *testing.T) {
		storer.Clear()
		longSlug := strings.Repeat("a", 51)
		body := fmt.Sprintf(`{"sessionId": "%s"}`, longSlug)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader(body))

		api.CreateSessionHandler(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestGetQuestionsHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "my-test-session"
	adminID := "admin-1"
	storer.PreloadSession(createMockSession(sessionID, adminID, true))

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/questions", nil)
		api.GetQuestionsHandler(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var questions []models.Question
		json.Unmarshal(w.Body.Bytes(), &questions)
		if len(questions) != 2 {
			t.Fatalf("Expected 2 questions, got %d", len(questions))
		}
		if questions[0].Votes < questions[1].Votes {
			t.Error("Questions were not sorted correctly by votes")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/non-existent-session/questions", nil)
		api.GetQuestionsHandler(w, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestSubmitQuestionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "submit-question-session"
	adminID := "admin-2"
	storer.PreloadSession(createMockSession(sessionID, adminID, true)) // Active session

	t.Run("Success", func(t *testing.T) {
		body := `{"text": "A new valid question?"}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/questions", strings.NewReader(body))
		api.SubmitQuestionHandler(w, r)
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if len(session.Questions) != 3 {
			t.Errorf("Expected 3 questions after submission, got %d", len(session.Questions))
		}
	})

	t.Run("SessionClosed", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminID, false)) // Inactive
		body := `{"text": "A question for a closed session?"}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/questions", strings.NewReader(body))
		api.SubmitQuestionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestVoteQuestionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "vote-session"
	adminID := "admin-3"
	questionID := "00000000-0000-0000-0000-000000000011"
	voterCookie := &http.Cookie{Name: "userSessionId", Value: "new-voter"}

	t.Run("Success", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminID, true))
		path := fmt.Sprintf("/api/session/%s/questions/%s/vote", sessionID, questionID)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, path, nil)
		r.AddCookie(voterCookie)
		api.VoteQuestionHandler(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if session.Questions[0].Votes != 11 {
			t.Errorf("Expected vote count to be 11, got %d", session.Questions[0].Votes)
		}
	})

	t.Run("AlreadyVoted", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminID, true))
		path := fmt.Sprintf("/api/session/%s/questions/%s/vote", sessionID, questionID)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, path, nil)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: "u1"}) // u1 already voted in mock
		api.VoteQuestionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestEndSessionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "end-this-session"
	adminID := "admin-4"
	otherUser := "user-4"

	t.Run("Success_Admin", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminID, true))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/session/"+sessionID, nil)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: adminID})
		api.EndSessionHandler(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
		if storer.HasSession(sessionID) {
			t.Error("Session was not deleted from storer")
		}
	})

	t.Run("Unauthorized_User", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminID, true))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/session/"+sessionID, nil)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: otherUser})
		api.EndSessionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		if !storer.HasSession(sessionID) {
			t.Error("Session was incorrectly deleted by non-admin")
		}
	})
}

func TestCheckAdminHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "check-admin-session"
	adminID := "admin-5"
	otherUser := "user-5"
	storer.PreloadSession(createMockSession(sessionID, adminID, true))

	t.Run("IsAdmin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/check-admin", nil)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: adminID})
		api.CheckAdminHandler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var resp map[string]bool
		json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp["isAdmin"] {
			t.Error("Expected isAdmin to be true")
		}
	})

	t.Run("IsNotAdmin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/check-admin", nil)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: otherUser})
		api.CheckAdminHandler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var resp map[string]bool
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["isAdmin"] {
			t.Error("Expected isAdmin to be false")
		}
	})
}
