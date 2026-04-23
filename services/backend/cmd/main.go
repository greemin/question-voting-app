package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"question-voting-app/internal/config"
	"question-voting-app/internal/handlers"
	"question-voting-app/internal/storage"
	"question-voting-app/internal/ws"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	var storer storage.Storer
	switch cfg.DBDriver {
	case "mongodb":
		if cfg.MongoURI == "" {
			log.Fatal("MONGO_URI environment variable not set")
		}
		connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(cfg.MongoURI)
		client, err := mongo.Connect(clientOptions)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
		if err = client.Ping(connCtx, nil); err != nil {
			log.Fatalf("Failed to ping MongoDB: %v", err)
		}
		fmt.Println("Connected to MongoDB!")
		storer = storage.NewMongoStorage(client, "question-voting-app", "sessions")

	default: // "sqlite"
		sqliteStorer, err := storage.NewSQLiteStorage(cfg.SQLiteFile)
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %v", err)
		}
		storer = sqliteStorer
	}

	if err := storer.ConfigureIndexes(ctx); err != nil {
		log.Fatalf("Failed to configure storage: %v", err)
	}

	isProduction := os.Getenv("ENV") == "production"
	hub := ws.NewHub(isProduction)
	api := handlers.New(storer, cfg.SecureCookie, hub)

	mux := SetupRouter(api, cfg.CORSOrigins)

	// --- Server Start ---
	fmt.Printf("Starting server on http://localhost%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}

// SetupRouter configures the API routes and applies CORS middleware.
func SetupRouter(api *handlers.API, corsOrigins string) http.Handler {
	// --- Request Logging Middleware ---
	loggingHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Millisecond))
		})
	}

	// --- CORS Middleware ---
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", corsOrigins)
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

	mux := http.NewServeMux()

	// Session Management
	mux.HandleFunc("POST /api/session", api.CreateSessionHandler)
	mux.HandleFunc("GET /api/session/{session_id}", api.GetSessionHandler)
	mux.HandleFunc("DELETE /api/session/{session_id}", api.EndSessionHandler)

	// Session Sub-resources
	mux.HandleFunc("GET /api/session/{session_id}/check-admin", api.CheckAdminHandler)
	mux.HandleFunc("GET /api/session/{session_id}/ws", api.ServeWS)

	// Questions & Voting
	mux.HandleFunc("POST /api/session/{session_id}/questions", api.SubmitQuestionHandler)
	mux.HandleFunc("DELETE /api/session/{session_id}/questions/{question_id}", api.DeleteQuestionHandler)
	mux.HandleFunc("PUT /api/session/{session_id}/questions/{question_id}/vote", api.VoteQuestionHandler)

	// Moderation
	mux.HandleFunc("POST /api/session/{session_id}/ban", api.BanIPHandler)

	return loggingHandler(corsHandler(mux))
}
