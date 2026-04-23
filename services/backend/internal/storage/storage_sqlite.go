package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"question-voting-app/internal/models"

	_ "modernc.org/sqlite"
)

// SQLiteStorage implements the Storer interface using SQLite.
// Session data is stored as a JSON blob in a single table, matching the
// whole-document-replace update pattern used by MongoStorage.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage opens (or creates) a SQLite database at the given DSN.
func NewSQLiteStorage(dsn string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	// SQLite allows only one concurrent writer; serialise all access through one connection.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	return &SQLiteStorage{db: db}, nil
}

// ConfigureIndexes creates the sessions table if it does not exist and starts
// the background TTL cleanup goroutine.
func (s *SQLiteStorage) ConfigureIndexes(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			created_at DATETIME NOT NULL,
			data       TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}
	go s.runCleanup()
	fmt.Println("SQLite storage configured successfully.")
	return nil
}

// runCleanup periodically deletes sessions older than 24 hours.
func (s *SQLiteStorage) runCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		if _, err := s.db.Exec(`DELETE FROM sessions WHERE created_at < ?`, cutoff); err != nil {
			fmt.Printf("SQLite TTL cleanup error: %v\n", err)
		}
	}
}

func (s *SQLiteStorage) LoadSessionData(ctx context.Context, sessionID string) (*models.SessionData, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT data FROM sessions WHERE session_id = ?`, sessionID).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to load session: %w", err)
	}
	var data models.SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	return &data, nil
}

func (s *SQLiteStorage) CreateSessionData(ctx context.Context, data *models.SessionData) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	createdAt := data.CreatedAt.UTC().Format(time.RFC3339)
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (session_id, created_at, data) VALUES (?, ?, ?)`,
		data.SessionID, createdAt, string(raw))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("session already exists: %w", ErrDuplicateKey)
		}
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) UpdateSessionData(ctx context.Context, data *models.SessionData) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET data = ? WHERE session_id = ?`,
		string(raw), data.SessionID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("session not found: %w", ErrNotFound)
	}
	return nil
}

func (s *SQLiteStorage) DeleteSessionData(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
