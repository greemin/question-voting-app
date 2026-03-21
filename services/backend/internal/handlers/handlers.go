package handlers

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
	"question-voting-app/internal/ws"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var slugInvalidChars = regexp.MustCompile(`[^\p{L}\p{N}\s-]+`)
var consecutiveHyphens = regexp.MustCompile(`-+`)
var spaceOrUnderscore = regexp.MustCompile(`[_\s]+`)

const (
	userSessionIDCookie = "userSessionId"
	authHeader          = "Authorization"
)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "&", " and ")
	s = strings.ReplaceAll(s, "+", " plus ")

	// Replace unsupported characters with a space
	s = slugInvalidChars.ReplaceAllString(s, " ")
	// Replace spaces and underscores with a hyphen
	s = spaceOrUnderscore.ReplaceAllString(s, "-")
	// Collapse consecutive hyphens
	s = consecutiveHyphens.ReplaceAllString(s, "-")
	// Trim leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// API holds the dependencies for the API handlers.
type API struct {
	Storer       storage.Storer
	SecureCookie bool
	Hub          *ws.Hub
}

// New creates a new API instance.
func New(storer storage.Storer, secureCookie bool, hub *ws.Hub) *API {
	return &API{
		Storer:       storer,
		SecureCookie: secureCookie,
		Hub:          hub,
	}
}

// getUserSessionID extracts the userSessionId from the cookie or generates a new one.
func (a *API) getUserSessionID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(userSessionIDCookie)
	if err == nil {
		return cookie.Value
	}

	newID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     userSessionIDCookie,
		Value:    newID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30, // 30 days
		Secure:   a.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	return newID
}

// CreateSessionHandler creates a new voting session.
// POST /api/session
func (a *API) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	a.getUserSessionID(w, r)          // Ensure creator also gets a voting cookie
	adminToken := uuid.New().String() // Generate secret admin token

	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.SessionID) > 50 {
		http.Error(w, "Session ID exceeds maximum length of 50 characters", http.StatusBadRequest)
		return
	}

	sessionID := req.SessionID
	if sessionID != "" {
		sessionID = slugify(sessionID)
	}

	if sessionID == "" {
		randomString, err := generateRandomString(8)
		if err != nil {
			http.Error(w, "Failed to generate session ID", http.StatusInternalServerError)
			return
		}
		sessionID = randomString
	}

	newSession := &models.SessionData{
		SessionID:   sessionID,
		AdminToken: adminToken, // Repurposing this field to store the token
		IsActive:    true,
		CreatedAt:   time.Now(),
		Questions:   []models.Question{},
	}

	// Retry logic for session ID collision
	for i := 0; i < 5; i++ {
		err := a.Storer.CreateSessionData(r.Context(), newSession)
		if err == nil {
			break // Success
		}

		if mongo.IsDuplicateKeyError(err) {
			// On the first collision, add a suffix. On subsequent collisions, generate a new random ID.
			if i == 0 {
				suffix, err := generateRandomString(4)
				if err != nil {
					http.Error(w, "Failed to generate session ID suffix", http.StatusInternalServerError)
					return
				}
				// Update both the session object AND the original sessionID variable for the next loop
				sessionID = sessionID + "-" + suffix
				newSession.SessionID = sessionID
			} else {
				randomString, err := generateRandomString(8)
				if err != nil {
					http.Error(w, "Failed to generate session ID", http.StatusInternalServerError)
					return
				}
				sessionID = randomString
				newSession.SessionID = sessionID
			}
		} else {
			http.Error(w, "Failed to create session data", http.StatusInternalServerError)
			return
		}

		if i == 4 {
			http.Error(w, "Failed to create session after multiple retries", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId":  newSession.SessionID,
		"adminToken": adminToken,
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

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
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
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil || strings.TrimSpace(submission.Text) == "" {
		http.Error(w, "Invalid request body or empty question", http.StatusBadRequest)
		return
	}

	if len(submission.Text) > 500 {
		http.Error(w, "Question exceeds maximum length of 500 characters", http.StatusBadRequest)
		return
	}

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
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

	if err := a.Storer.UpdateSessionData(r.Context(), sessionData); err != nil {
		http.Error(w, "Failed to save question", http.StatusInternalServerError)
		return
	}

	// Broadcast update
	if a.Hub != nil {
		event := map[string]interface{}{
			"type":    "QUESTION_ADDED",
			"payload": newQuestion,
		}
		if msg, err := json.Marshal(event); err == nil {
			a.Hub.Broadcast(sessionID, msg)
		}
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

	if _, err := uuid.Parse(questionID); err != nil {
		http.Error(w, "Invalid question ID format", http.StatusBadRequest)
		return
	}

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
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

			if err := a.Storer.UpdateSessionData(r.Context(), sessionData); err != nil {
				http.Error(w, "Failed to record vote", http.StatusInternalServerError)
				return
			}

			// Broadcast update
			if a.Hub != nil {
				event := map[string]interface{}{
					"type":    "VOTE_UPDATED",
					"payload": sessionData.Questions[i],
				}
				if msg, err := json.Marshal(event); err == nil {
					a.Hub.Broadcast(sessionID, msg)
				}
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(sessionData.Questions[i])
			return
		}
	}

	http.Error(w, "Question not found", http.StatusNotFound)
}

// DeleteQuestionHandler allows the admin to delete a question.
// DELETE /api/session/{sessionId}/questions/{questionId}
func (a *API) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]
	questionID := parts[5]

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if sessionData.AdminToken != providedToken || providedToken == "" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	found := false
	var updatedQuestions []models.Question
	for _, q := range sessionData.Questions {
		if q.ID == questionID {
			found = true
		} else {
			updatedQuestions = append(updatedQuestions, q)
		}
	}

	if !found {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	sessionData.Questions = updatedQuestions

	if err := a.Storer.UpdateSessionData(r.Context(), sessionData); err != nil {
		http.Error(w, "Failed to delete question", http.StatusInternalServerError)
		return
	}

	// Broadcast update
	if a.Hub != nil {
		event := map[string]interface{}{
			"type":    "QUESTION_DELETED",
			"payload": map[string]string{"id": questionID},
		}
		if msg, err := json.Marshal(event); err == nil {
			a.Hub.Broadcast(sessionID, msg)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// EndSessionHandler allows the admin to end the session and delete the file.
// DELETE /api/session/{sessionId}
func (a *API) EndSessionHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid session ID in path", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if sessionData.AdminToken != providedToken || providedToken == "" {
		http.Error(w, "Unauthorized: Only the session creator can end the session.", http.StatusForbidden)
		return
	}

	if err := a.Storer.DeleteSessionData(r.Context(), sessionID); err != nil {
		http.Error(w, "Failed to delete session file", http.StatusInternalServerError)
		return
	}

	// Broadcast that the session ended
	if a.Hub != nil {
		event := map[string]interface{}{
			"type": "SESSION_ENDED",
		}
		if msg, err := json.Marshal(event); err == nil {
			a.Hub.Broadcast(sessionID, msg)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckAdminHandler checks if the current user holds the secret admin token.
// GET /api/session/{sessionId}/check-admin
func (a *API) CheckAdminHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid session ID in path", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}

	isAdmin := sessionData.AdminToken == providedToken && providedToken != ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"isAdmin": isAdmin})
}

// ServeWS handles WebSocket requests from the frontend.
// GET /api/session/{sessionId}/ws
func (a *API) ServeWS(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[len(parts)-1] != "ws" {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}
	sessionID := parts[3]

	// Ensure the session exists before allowing a websocket connection
	_, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if a.Hub != nil {
		a.Hub.ServeWS(w, r, sessionID)
	}
}
