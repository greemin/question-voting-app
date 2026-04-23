package storage

import (
	"context"
	"errors"
	"question-voting-app/internal/models"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicateKey = errors.New("duplicate key")

// Storer defines the interface for session data storage.
type Storer interface {
	LoadSessionData(ctx context.Context, sessionID string) (*models.SessionData, error)
	CreateSessionData(ctx context.Context, data *models.SessionData) error
	UpdateSessionData(ctx context.Context, data *models.SessionData) error
	DeleteSessionData(ctx context.Context, sessionID string) error
	ConfigureIndexes(ctx context.Context) error
}
