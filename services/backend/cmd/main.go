package main

import (
	"fmt"
	"log"
	"net/http"
	"question-voting-app/internal/handlers"
	"strings"
)

func main() {
	// --- CORS Middleware (Crucial for development with React) ---
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Your Vite Dev Server
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true") // Must be true for cookies (userSessionId)

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// --- API Routes Setup ---
	mux := http.NewServeMux()

	// POST /api/session - Create a new session
	mux.Handle("/api/session", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateSessionHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Dynamic handler for /api/session/{sessionId}/... routes
	mux.Handle("/api/session/", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check for the new admin route BEFORE others
		if r.Method == http.MethodGet && strings.HasSuffix(path, "/check-admin") {
			// GET /api/session/{sessionId}/check-admin
			handlers.CheckAdminHandler(w, r)
		} else if r.Method == http.MethodGet && strings.HasSuffix(path, "/questions") {
			// GET /api/session/{sessionId}/questions
			handlers.GetQuestionsHandler(w, r)
		} else if r.Method == http.MethodPost && strings.HasSuffix(path, "/questions") {
			// POST /api/session/{sessionId}/questions
			handlers.SubmitQuestionHandler(w, r)
		} else if r.Method == http.MethodPut && strings.HasSuffix(path, "/vote") {
			// PUT /api/session/{sessionId}/questions/{questionId}/vote
			handlers.VoteQuestionHandler(w, r)
		} else if r.Method == http.MethodDelete && strings.Count(path, "/") == 3 {
			// DELETE /api/session/{sessionId}
			handlers.EndSessionHandler(w, r)
		} else {
			http.Error(w, "Not found or Method not allowed", http.StatusNotFound)
		}
	})))

	// --- Server Start ---
	port := ":8081"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
