package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"question-voting-app/internal/handlers"
	"question-voting-app/internal/storage"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	storer := storage.NewMongoStorage(client, "question-voting-app", "sessions")
	if err := storer.ConfigureIndexes(ctx); err != nil {
		log.Fatalf("Failed to configure MongoDB indexes: %v", err)
	}
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
