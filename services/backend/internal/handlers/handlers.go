package handlers

import (
	"encoding/json"
	"net/http"
	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// API holds the dependencies for the API handlers.
type API struct {
	Storer storage.Storer
}

// New creates a new API instance.
func New(storer storage.Storer) *API {
	return &API{Storer: storer}
}

// getUserSessionID extracts the userSessionId from the cookie or generates a new one.
func (a *API) getUserSessionID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("userSessionId")
	if err == nil {
		return cookie.Value
	}

	newID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "userSessionId",
		Value:    newID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30, // 30 days
		Secure:   false,      // Change to true in production
		SameSite: http.SameSiteLaxMode,
	})

	return newID
}

// CreateSessionHandler creates a new voting session.
// POST /api/session
func (a *API) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	adminID := a.getUserSessionID(w, r)
	sessionID := uuid.New().String()

	newSession := &models.SessionData{
		SessionID:   sessionID,
		AdminUserID: adminID,
		IsActive:    true,
		Questions:   []models.Question{},
	}

	if err := a.Storer.SaveSessionData(newSession); err != nil {
		http.Error(w, "Failed to create session data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId": sessionID,
		"adminId":   adminID,
	})
}

// GetQuestionsHandler retrieves and sorts all questions for a session.
// GET /api/session/{sessionId}/questions
func (a *API) GetQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[len(parts)-1] != "questions" {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]

	sessionData, err := a.Storer.LoadSessionData(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	sort.Slice(sessionData.Questions, func(i, j int) bool {
		return sessionData.Questions[i].Votes > sessionData.Questions[j].Votes
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionData.Questions)
}

// SubmitQuestionHandler adds a new question to the session.
// POST /api/session/{sessionId}/questions
func (a *API) SubmitQuestionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.Split(r.URL.Path, "/")[3]
	a.getUserSessionID(w, r) // Ensure user has a session cookie

	var submission models.QuestionSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil || submission.Text == "" {
		http.Error(w, "Invalid request body or empty question", http.StatusBadRequest)
		return
	}

	sessionData, err := a.Storer.LoadSessionData(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !sessionData.IsActive {
		http.Error(w, "Voting session is closed", http.StatusForbidden)
		return
	}

	newQuestion := models.Question{
		ID:     uuid.New().String(),
		Text:   submission.Text,
		Votes:  0,
		Voters: []string{},
	}

	sessionData.Questions = append(sessionData.Questions, newQuestion)

	if err := a.Storer.SaveSessionData(sessionData); err != nil {
		http.Error(w, "Failed to save question", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newQuestion)
}

// VoteQuestionHandler increments the vote count for a question.
// PUT /api/session/{sessionId}/questions/{questionId}/vote
func (a *API) VoteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 || parts[len(parts)-1] != "vote" {
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]
	questionID := parts[5]
	userID := a.getUserSessionID(w, r)

	sessionData, err := a.Storer.LoadSessionData(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !sessionData.IsActive {
		http.Error(w, "Voting session is closed", http.StatusForbidden)
		return
	}

	for i, q := range sessionData.Questions {
		if q.ID == questionID {
			for _, voterID := range q.Voters {
				if voterID == userID {
					http.Error(w, "Already voted on this question in this session", http.StatusForbidden)
					return
				}
			}

			sessionData.Questions[i].Votes++
			sessionData.Questions[i].Voters = append(sessionData.Questions[i].Voters, userID)

			if err := a.Storer.SaveSessionData(sessionData); err != nil {
				http.Error(w, "Failed to record vote", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(sessionData.Questions[i])
			return
		}
	}

	http.Error(w, "Question not found", http.StatusNotFound)
}

// EndSessionHandler allows the admin to end the session and delete the file.
// DELETE /api/session/{sessionId}
func (a *API) EndSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.Split(r.URL.Path, "/")[3]
	userID := a.getUserSessionID(w, r)

	sessionData, err := a.Storer.LoadSessionData(sessionID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if sessionData.AdminUserID != userID {
		http.Error(w, "Unauthorized: Only the session creator can end the session.", http.StatusForbidden)
		return
	}

	if err := a.Storer.DeleteSessionData(sessionID); err != nil {
		http.Error(w, "Failed to delete session file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAdminHandler checks if the current user (via cookie) is the session admin.
// GET /api/session/{sessionId}/check-admin
func (a *API) CheckAdminHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid session ID in path", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]
	currentUserID := a.getUserSessionID(w, r)

	sessionData, err := a.Storer.LoadSessionData(sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}

	isAdmin := sessionData.AdminUserID == currentUserID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"isAdmin": isAdmin})
}
