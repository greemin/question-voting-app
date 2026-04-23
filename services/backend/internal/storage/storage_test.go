//go:build integration

package storage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"question-voting-app/internal/models"
	"question-voting-app/internal/storage"
)

func TestMongoStorageCRUD(t *testing.T) {
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

	testStorerCRUD(t, store)
}

func TestSQLiteStorageCRUD(t *testing.T) {
	store, err := storage.NewSQLiteStorage(":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := store.ConfigureIndexes(context.Background()); err != nil {
		t.Fatalf("failed to configure indexes: %v", err)
	}

	testStorerCRUD(t, store)
}

func testStorerCRUD(t *testing.T, store storage.Storer) {
	t.Helper()
	ctx := context.Background()

	session := &models.SessionData{
		SessionID:    "test-session",
		SessionTitle: "Test Session",
		IsActive:     true,
		CreatedAt:    time.Now().Truncate(time.Second),
		Questions:    []models.Question{},
	}

	t.Run("create", func(t *testing.T) {
		if err := store.CreateSessionData(ctx, session); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("create duplicate returns ErrDuplicateKey", func(t *testing.T) {
		err := store.CreateSessionData(ctx, session)
		if !errors.Is(err, storage.ErrDuplicateKey) {
			t.Fatalf("expected ErrDuplicateKey, got: %v", err)
		}
	})

	t.Run("load", func(t *testing.T) {
		got, err := store.LoadSessionData(ctx, session.SessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SessionID != session.SessionID {
			t.Errorf("SessionID: got %q, want %q", got.SessionID, session.SessionID)
		}
		if got.SessionTitle != session.SessionTitle {
			t.Errorf("SessionTitle: got %q, want %q", got.SessionTitle, session.SessionTitle)
		}
	})

	t.Run("load missing returns ErrNotFound", func(t *testing.T) {
		_, err := store.LoadSessionData(ctx, "does-not-exist")
		if !errors.Is(err, storage.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		session.SessionTitle = "Updated Title"
		session.Questions = []models.Question{
			{ID: "q1", Text: "First question", Votes: 3, Voters: []string{"a", "b", "c"}},
		}
		if err := store.UpdateSessionData(ctx, session); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := store.LoadSessionData(ctx, session.SessionID)
		if err != nil {
			t.Fatalf("load after update failed: %v", err)
		}
		if got.SessionTitle != "Updated Title" {
			t.Errorf("SessionTitle not persisted: got %q", got.SessionTitle)
		}
		if len(got.Questions) != 1 || got.Questions[0].Votes != 3 {
			t.Errorf("Questions not persisted correctly: %+v", got.Questions)
		}
	})

	t.Run("delete", func(t *testing.T) {
		if err := store.DeleteSessionData(ctx, session.SessionID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := store.LoadSessionData(ctx, session.SessionID)
		if !errors.Is(err, storage.ErrNotFound) {
			t.Fatalf("expected ErrNotFound after delete, got: %v", err)
		}
	})
}
