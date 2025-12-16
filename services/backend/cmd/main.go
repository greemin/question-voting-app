package main

import (
	"fmt"
	"log"
	"net/http"
	"question-voting-app/internal/handlers"
	"question-voting-app/internal/storage"
	"strings"
)

func main() {
	// --- CORS Middleware ---
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// --- Dependencies Setup ---
	storer := storage.NewFileStorage("data")
	api := handlers.New(storer)

	// --- API Routes Setup ---
	mux := http.NewServeMux()

	mux.Handle("/api/session", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateSessionHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/session/", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == http.MethodGet && strings.HasSuffix(path, "/check-admin") {
			api.CheckAdminHandler(w, r)
		} else if r.Method == http.MethodGet && strings.HasSuffix(path, "/questions") {
			api.GetQuestionsHandler(w, r)
		} else if r.Method == http.MethodPost && strings.HasSuffix(path, "/questions") {
			api.SubmitQuestionHandler(w, r)
		} else if r.Method == http.MethodPut && strings.HasSuffix(path, "/vote") {
			api.VoteQuestionHandler(w, r)
		} else if r.Method == http.MethodDelete && strings.Count(path, "/") == 3 {
			api.EndSessionHandler(w, r)
		} else {
			http.Error(w, "Not found or Method not allowed", http.StatusNotFound)
		}
	})))

	// --- Server Start ---
	port := ":8081"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
