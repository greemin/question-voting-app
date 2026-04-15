//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
)

// TestCreateSessionData_DuplicateKeyError verifies that inserting two sessions
// with the same sessionId returns an error that mongo.IsDuplicateKeyError
// correctly identifies — even after the storage layer wraps it with fmt.Errorf("%w").
// This guards against regressions where the wrapping changes from %w to %s,
// which would silently break the collision-retry logic in the handler.
func TestCreateSessionData_DuplicateKeyError(t *testing.T) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("failed to start mongodb container: %v", err)
	}
	t.Cleanup(func() { mongoContainer.Terminate(ctx) })

	connString, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		t.Fatalf("failed to connect to mongodb: %v", err)
	}
	t.Cleanup(func() { client.Disconnect(ctx) })

	store := storage.NewMongoStorage(client, "testdb", "sessions")
	if err := store.ConfigureIndexes(ctx); err != nil {
		t.Fatalf("failed to configure indexes: %v", err)
	}

	session := &models.SessionData{
		SessionID:    "my-session",
		SessionTitle: "My Session",
		IsActive:     true,
		CreatedAt:    time.Now(),
		Questions:    []models.Question{},
	}

	if err := store.CreateSessionData(ctx, session); err != nil {
		t.Fatalf("first insert failed unexpectedly: %v", err)
	}

	err = store.CreateSessionData(ctx, session)
	if err == nil {
		t.Fatal("expected a duplicate key error on second insert, got nil")
	}

	if !mongo.IsDuplicateKeyError(err) {
		t.Errorf("mongo.IsDuplicateKeyError returned false on the wrapped error; "+
			"the error wrapping in CreateSessionData may have broken unwrapping — error was: %v", err)
	}
}
