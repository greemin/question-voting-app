package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"net/http"
	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
	"question-voting-app/internal/ws"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.SplitN(fwd, ",", 2)[0]
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

var slugInvalidChars = regexp.MustCompile(`[^\p{L}\p{N}\s-]+`)
var consecutiveHyphens = regexp.MustCompile(`-+`)
var spaceOrUnderscore = regexp.MustCompile(`[_\s]+`)

const (
	userSessionIDCookie    = "userSessionId"
	authHeader             = "Authorization"
	maxRequestBodyBytes    = 4096
	maxQuestionsPerSession = 200
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

func deslugify(s string, lang language.Tag) string {
	s = strings.ReplaceAll(s, "-", " ")
	return cases.Title(lang).String(s)
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

func newSessionData(sessionID, sessionTitle string) *models.SessionData {
	return &models.SessionData{
		SessionID:    sessionID,
		SessionTitle: sessionTitle,
		AdminToken:   uuid.New().String(),
		IsActive:     true,
		CreatedAt:    time.Now(),
		Questions:    []models.Question{},
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// We keep the "not found" string check here as a fallback for the mock storer used in tests
	return errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(strings.ToLower(err.Error()), "not found")
}

func isDuplicateKeyError(err error) bool {
	if mongo.IsDuplicateKeyError(err) {
		return true
	}
	// Fallback for older mongo versions or weirdly wrapped errors that don't unwrap to a mongo.WriteException.
	// The error code for a duplicate key error is 11000.
	return strings.Contains(err.Error(), "E11000")
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

func (a *API) createSessionWithRetry(ctx context.Context, sessionID, sessionTitle string) (*models.SessionData, error) {
	newSession := newSessionData(sessionID, sessionTitle)

	// Retry logic for session ID collision
	for i := 0; i < 5; i++ {
		err := a.Storer.CreateSessionData(ctx, newSession)
		if err == nil {
			return newSession, nil
		}

		if isDuplicateKeyError(err) {
			// On the first collision, add a suffix. On subsequent collisions, generate a new random ID.
			if i == 0 {
				suffix, err := generateRandomString(4)
				if err != nil {
					return nil, err
				}
				sessionID = sessionID + "-" + suffix
				newSession.SessionID = sessionID
			} else {
				randomString, err := generateRandomString(8)
				if err != nil {
					return nil, err
				}
				sessionID = randomString
				newSession.SessionID = sessionID
			}
		} else {
			return nil, err
		}
	}

	return nil, errors.New("failed to create session after multiple retries")
}

// CreateSessionHandler creates a new voting session.
// POST /api/session
func (a *API) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	a.getUserSessionID(w, r) // Ensure creator also gets a voting cookie

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
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
	sessionTitle := req.SessionID
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

	newSession, err := a.createSessionWithRetry(r.Context(), sessionID, sessionTitle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId":    newSession.SessionID,
		"sessionTitle": newSession.SessionTitle,
		"adminToken":   newSession.AdminToken,
	})
}

// GetSessionHandler retrieves the full session data (excluding sensitive info).
// GET /api/session/{session_id}
func (a *API) GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			a.getUserSessionID(w, r) // Ensure creator also gets a voting cookie

			// Determine language from the Accept-Language header
			tags, _, parseErr := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
			lang := language.English // Fallback to English
			if parseErr == nil && len(tags) > 0 {
				lang = tags[0]
			}

			sessionTitle := deslugify(sessionID, lang)

			newSession, createErr := a.createSessionWithRetry(r.Context(), sessionID, sessionTitle)
			if createErr != nil {
				http.Error(w, "Failed to create session", http.StatusInternalServerError)
				return
			}

			// Session created successfully by this request.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			// Return a response with the full session data and admin token.
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sessionId":    newSession.SessionID,
				"sessionTitle": newSession.SessionTitle,
				"adminToken":   newSession.AdminToken,
				"isActive":     newSession.IsActive,
				"createdAt":    newSession.CreatedAt,
				"questions":    newSession.Questions,
			})
			return
		}

		http.Error(w, "Failed to load session", http.StatusInternalServerError)
		return
	}

	// Session existed, or was created concurrently and is now loaded.
	sort.Slice(sessionData.Questions, func(i, j int) bool {
		return sessionData.Questions[i].Votes > sessionData.Questions[j].Votes
	})

	// Omit AdminToken for security on normal GETs of existing sessions.
	response := struct {
		SessionID    string            `json:"sessionId"`
		SessionTitle string            `json:"sessionTitle"`
		IsActive     bool              `json:"isActive"`
		CreatedAt    time.Time         `json:"createdAt"`
		Questions    []models.Question `json:"questions"`
	}{
		SessionID:    sessionData.SessionID,
		SessionTitle: sessionData.SessionTitle,
		IsActive:     sessionData.IsActive,
		CreatedAt:    sessionData.CreatedAt,
		Questions:    sessionData.Questions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SubmitQuestionHandler adds a new question to the session.
// POST /api/session/{session_id}/questions
func (a *API) SubmitQuestionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	a.getUserSessionID(w, r) // Ensure user has a session cookie

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
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
		if isNotFoundError(err) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
		return
	}

	if !sessionData.IsActive {
		http.Error(w, "Voting session is closed", http.StatusForbidden)
		return
	}

	if len(sessionData.Questions) >= maxQuestionsPerSession {
		http.Error(w, "Session has reached the maximum number of questions", http.StatusForbidden)
		return
	}

	clientIP := getClientIP(r)
	for _, banned := range sessionData.BannedIPs {
		if banned == clientIP {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	newQuestion := models.Question{
		ID:          uuid.New().String(),
		Text:        submission.Text,
		Votes:       0,
		Voters:      []string{},
		SubmitterIP: clientIP,
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
// PUT /api/session/{session_id}/questions/{question_id}/vote
func (a *API) VoteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	questionID := r.PathValue("question_id")

	cookie, err := r.Cookie(userSessionIDCookie)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	userID := cookie.Value

	if _, err := uuid.Parse(questionID); err != nil {
		http.Error(w, "Invalid question ID format", http.StatusBadRequest)
		return
	}

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
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
// DELETE /api/session/{session_id}/questions/{question_id}
func (a *API) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	questionID := r.PathValue("question_id")

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
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
// DELETE /api/session/{session_id}
func (a *API) EndSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
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
// GET /api/session/{session_id}/check-admin
func (a *API) CheckAdminHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
		return
	}

	isAdmin := sessionData.AdminToken == providedToken && providedToken != ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"isAdmin": isAdmin})
}

// BanIPHandler bans the submitter IP of a question and removes all their questions.
// POST /api/session/{session_id}/ban
func (a *API) BanIPHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")

	authHeader := r.Header.Get(authHeader)
	providedToken := strings.TrimPrefix(authHeader, "Bearer ")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req struct {
		QuestionID string `json:"questionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.QuestionID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionData, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
		return
	}

	if sessionData.AdminToken != providedToken || providedToken == "" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var targetIP string
	for _, q := range sessionData.Questions {
		if q.ID == req.QuestionID {
			targetIP = q.SubmitterIP
			break
		}
	}

	if targetIP == "" {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	if targetIP == getClientIP(r) {
		http.Error(w, "Cannot ban yourself", http.StatusForbidden)
		return
	}

	for _, ip := range sessionData.BannedIPs {
		if ip == targetIP {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	sessionData.BannedIPs = append(sessionData.BannedIPs, targetIP)

	var removedIDs []string
	remaining := []models.Question{}
	for _, q := range sessionData.Questions {
		if q.SubmitterIP == targetIP {
			removedIDs = append(removedIDs, q.ID)
		} else {
			remaining = append(remaining, q)
		}
	}
	sessionData.Questions = remaining

	if err := a.Storer.UpdateSessionData(r.Context(), sessionData); err != nil {
		http.Error(w, "Failed to ban submitter", http.StatusInternalServerError)
		return
	}

	if a.Hub != nil {
		event := map[string]interface{}{
			"type":    "IP_BANNED",
			"payload": map[string]interface{}{"questionIds": removedIDs},
		}
		if msg, err := json.Marshal(event); err == nil {
			a.Hub.Broadcast(sessionID, msg)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// ServeWS handles WebSocket requests from the frontend.
// GET /api/session/{session_id}/ws
func (a *API) ServeWS(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")

	// Ensure the session exists before allowing a websocket connection
	_, err := a.Storer.LoadSessionData(r.Context(), sessionID)
	if err != nil {
		if isNotFoundError(err) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load session", http.StatusInternalServerError)
		return
	}

	if a.Hub != nil {
		a.Hub.ServeWS(w, r, sessionID)
	}
}
