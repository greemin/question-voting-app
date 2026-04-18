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
	"question-voting-app/internal/ws"

	"golang.org/x/text/language"
)

// --- Test Setup Helper ---

// setupTestAPI creates the handler instance with an injected MockStorer.
func setupTestAPI() (*API, *testutil.MockStorer) {
	storer := testutil.NewMockStorer()
	hub := ws.NewHub(false)
	api := New(storer, false, hub)
	return api, storer
}

// createMockSession creates a SessionData object for testing.
func createMockSession(id string, adminToken string, isActive bool) *models.SessionData {
	return &models.SessionData{
		SessionID:  id,
		AdminToken: adminToken,
		IsActive:   isActive,
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

func TestDeslugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal", "hello-world", "Hello World"},
		{"Single word", "session", "Session"},
		{"With numbers", "session-123", "Session 123"},
		{"From slugify", "my-awesome-q-and-a", "My Awesome Q And A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := deslugify(tt.input, language.English)
			if actual != tt.expected {
				t.Errorf("deslugify(%q): expected %q, got %q", tt.input, tt.expected, actual)
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
		r := httptest.NewRequest(http.MethodPost, "/api/session", strings.NewReader("{}"))

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

		adminToken := resp["adminToken"]
		if adminToken == "" {
			t.Error("Expected an adminToken to be returned")
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

func TestServeWS(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "ws-session"
	storer.PreloadSession(createMockSession(sessionID, "admin-1", true))

	t.Run("SessionNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/non-existent/ws", nil)
		api.ServeWS(w, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/websock", nil)
		r.SetPathValue("session_id", sessionID)
		api.ServeWS(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("ValidPathButNotWebSocket", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/ws", nil)
		r.SetPathValue("session_id", sessionID)
		api.ServeWS(w, r)

		// Since the request doesn't have proper websocket upgrade headers,
		// upgrader.Upgrade will fail, log the error, and write a 400 Bad Request status.
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestGetSessionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "my-test-session"
	adminID := "admin-1"
	storer.PreloadSession(createMockSession(sessionID, adminID, true))

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/questions", nil)
		r.SetPathValue("session_id", sessionID)
		api.GetSessionHandler(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		var sessionData models.SessionData
		json.Unmarshal(w.Body.Bytes(), &sessionData)
		if len(sessionData.Questions) != 2 {
			t.Fatalf("Expected 2 questions, got %d", len(sessionData.Questions))
		}
		if sessionData.Questions[0].Votes < sessionData.Questions[1].Votes {
			t.Error("Questions were not sorted correctly by votes")
		}
	})

	t.Run("CreatesNewSessionIfNotFound", func(t *testing.T) {
		storer.Clear() // Make sure no sessions exist
		newSessionID := "a-new-session"
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+newSessionID, nil)
		r.SetPathValue("session_id", newSessionID)

		api.GetSessionHandler(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp["sessionId"] != newSessionID {
			t.Errorf("Expected sessionId %q, got %q", newSessionID, resp["sessionId"])
		}

		if resp["sessionTitle"] != "A New Session" {
			t.Errorf("Expected sessionTitle 'A New Session', got %q", resp["sessionTitle"])
		}

		if resp["adminToken"] == "" || resp["adminToken"] == nil {
			t.Error("Expected an adminToken to be returned for a new session")
		}

		if !storer.HasSession(newSessionID) {
			t.Error("Session was not created in the storer")
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
		r.SetPathValue("session_id", sessionID)
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
		r.SetPathValue("session_id", sessionID)
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
		r.SetPathValue("session_id", sessionID)
		r.SetPathValue("question_id", questionID)
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
		r.SetPathValue("session_id", sessionID)
		r.SetPathValue("question_id", questionID)
		r.AddCookie(&http.Cookie{Name: "userSessionId", Value: "u1"}) // u1 already voted in mock
		api.VoteQuestionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestDeleteQuestionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "delete-q-session"
	adminToken := "secret-admin-token"
	questionID := "00000000-0000-0000-0000-000000000011"

	t.Run("Success_Admin", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminToken, true))
		path := fmt.Sprintf("/api/session/%s/questions/%s", sessionID, questionID)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		r.SetPathValue("session_id", sessionID)
		r.SetPathValue("question_id", questionID)
		r.Header.Set("Authorization", "Bearer "+adminToken)
		api.DeleteQuestionHandler(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if len(session.Questions) != 1 {
			t.Errorf("Expected 1 question remaining, got %d", len(session.Questions))
		}
	})

	t.Run("Unauthorized_User", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminToken, true))
		path := fmt.Sprintf("/api/session/%s/questions/%s", sessionID, questionID)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, path, nil)
		r.SetPathValue("session_id", sessionID)
		r.SetPathValue("question_id", questionID)
		r.Header.Set("Authorization", "Bearer invalid-token")
		api.DeleteQuestionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestEndSessionHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "end-this-session"
	adminToken := "secret-admin-token"

	t.Run("Success_Admin", func(t *testing.T) {
		storer.Clear()
		storer.PreloadSession(createMockSession(sessionID, adminToken, true))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/session/"+sessionID, nil)
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer "+adminToken)
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
		storer.PreloadSession(createMockSession(sessionID, adminToken, true))
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/api/session/"+sessionID, nil)
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer invalid-token")
		api.EndSessionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		if !storer.HasSession(sessionID) {
			t.Error("Session was incorrectly deleted by non-admin")
		}
	})
}

func TestSubmitQuestionHandler_BannedIP(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "ban-submit-session"
	bannedIP := "10.0.0.1"
	session := createMockSession(sessionID, "admin-token", true)
	session.BannedIPs = []string{bannedIP}
	storer.PreloadSession(session)

	t.Run("BannedIPRejected", func(t *testing.T) {
		body := `{"text": "spam question"}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/questions", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("X-Real-IP", bannedIP)
		api.SubmitQuestionHandler(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d for banned IP, got %d", http.StatusForbidden, w.Code)
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if len(session.Questions) != 2 {
			t.Errorf("Expected question count to remain 2, got %d", len(session.Questions))
		}
	})

	t.Run("AllowedIPAccepted", func(t *testing.T) {
		body := `{"text": "legitimate question"}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/questions", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("X-Real-IP", "10.0.0.2")
		api.SubmitQuestionHandler(w, r)
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d for allowed IP, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestBanIPHandler(t *testing.T) {
	const (
		sessionID  = "ban-session"
		adminToken = "admin-secret"
		adminIP    = "5.6.7.8"
		spamIP     = "1.2.3.4"
		otherIP    = "9.9.9.9"
		qSpam1     = "00000000-0000-0000-0000-000000000001"
		qSpam2     = "00000000-0000-0000-0000-000000000002"
		qOther     = "00000000-0000-0000-0000-000000000003"
	)

	makeSession := func() *models.SessionData {
		return &models.SessionData{
			SessionID:  sessionID,
			AdminToken: adminToken,
			IsActive:   true,
			BannedIPs:  []string{},
			Questions: []models.Question{
				{ID: qSpam1, Text: "spam 1", SubmitterIP: spamIP},
				{ID: qSpam2, Text: "spam 2", SubmitterIP: spamIP},
				{ID: qOther, Text: "legit question", SubmitterIP: otherIP},
			},
		}
	}

	banRequest := func(api *API, questionID string, requesterIP string) *httptest.ResponseRecorder {
		body := fmt.Sprintf(`{"questionId": "%s"}`, questionID)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/ban", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer "+adminToken)
		r.Header.Set("X-Real-IP", requesterIP)
		api.BanIPHandler(w, r)
		return w
	}

	t.Run("Success_RemovesAllQuestionsFromBannedIP", func(t *testing.T) {
		api, storer := setupTestAPI()
		storer.PreloadSession(makeSession())

		w := banRequest(api, qSpam1, adminIP)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if len(session.Questions) != 1 {
			t.Errorf("Expected 1 question remaining, got %d", len(session.Questions))
		}
		if session.Questions[0].ID != qOther {
			t.Errorf("Expected legit question to remain, got %q", session.Questions[0].ID)
		}
		if len(session.BannedIPs) != 1 || session.BannedIPs[0] != spamIP {
			t.Errorf("Expected BannedIPs to contain %q, got %v", spamIP, session.BannedIPs)
		}
	})

	t.Run("CannotBanYourself", func(t *testing.T) {
		api, storer := setupTestAPI()
		// Admin's question uses the same IP as the admin making the request
		session := makeSession()
		session.Questions[0].SubmitterIP = adminIP
		storer.PreloadSession(session)

		w := banRequest(api, qSpam1, adminIP)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		session, _ = storer.LoadSessionData(context.Background(), sessionID)
		if len(session.BannedIPs) != 0 {
			t.Error("Expected no IPs to be banned after self-ban attempt")
		}
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		api, storer := setupTestAPI()
		storer.PreloadSession(makeSession())

		body := fmt.Sprintf(`{"questionId": "%s"}`, qSpam1)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/ban", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("X-Real-IP", adminIP)
		api.BanIPHandler(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Unauthorized_WrongToken", func(t *testing.T) {
		api, storer := setupTestAPI()
		storer.PreloadSession(makeSession())

		body := fmt.Sprintf(`{"questionId": "%s"}`, qSpam1)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/ban", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer wrong-token")
		r.Header.Set("X-Real-IP", adminIP)
		api.BanIPHandler(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		session, _ := storer.LoadSessionData(context.Background(), sessionID)
		if len(session.BannedIPs) != 0 {
			t.Error("Expected no IPs to be banned after unauthorized request")
		}
	})

	t.Run("QuestionNotFound", func(t *testing.T) {
		api, storer := setupTestAPI()
		storer.PreloadSession(makeSession())

		w := banRequest(api, "00000000-0000-0000-0000-000000000099", adminIP)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("AlreadyBanned_Idempotent", func(t *testing.T) {
		api, storer := setupTestAPI()
		session := makeSession()
		session.BannedIPs = []string{spamIP}
		storer.PreloadSession(session)

		w := banRequest(api, qSpam1, adminIP)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d for already-banned IP, got %d", http.StatusNoContent, w.Code)
		}
		session, _ = storer.LoadSessionData(context.Background(), sessionID)
		if len(session.BannedIPs) != 1 {
			t.Errorf("Expected BannedIPs to still have 1 entry, got %d", len(session.BannedIPs))
		}
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		api, storer := setupTestAPI()
		storer.Clear()

		body := fmt.Sprintf(`{"questionId": "%s"}`, qSpam1)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/session/"+sessionID+"/ban", strings.NewReader(body))
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer "+adminToken)
		r.Header.Set("X-Real-IP", adminIP)
		api.BanIPHandler(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestCheckAdminHandler(t *testing.T) {
	api, storer := setupTestAPI()
	sessionID := "check-admin-session"
	adminToken := "secret-admin-token"
	storer.PreloadSession(createMockSession(sessionID, adminToken, true))

	t.Run("IsAdmin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/session/"+sessionID+"/check-admin", nil)
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer "+adminToken)
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
		r.SetPathValue("session_id", sessionID)
		r.Header.Set("Authorization", "Bearer wrong-token")
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
