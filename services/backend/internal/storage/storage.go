package storage

import (
	"context"
	"fmt"
	"question-voting-app/internal/models"
	"time"

	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Storer defines the interface for session data storage.
type Storer interface {
	LoadSessionData(ctx context.Context, sessionID string) (*models.SessionData, error)
	CreateSessionData(ctx context.Context, data *models.SessionData) error
	UpdateSessionData(ctx context.Context, data *models.SessionData) error
	DeleteSessionData(ctx context.Context, sessionID string) error
	ConfigureIndexes(ctx context.Context) error
}

// MongoStorage implements the Storer interface for MongoDB-based storage.
type MongoStorage struct {
	collection *mongo.Collection
}

// NewMongoStorage creates a new instance of MongoStorage.
func NewMongoStorage(client *mongo.Client, dbName, collectionName string) *MongoStorage {
	return &MongoStorage{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

// ConfigureIndexes sets up the necessary indexes in the MongoDB collection.
func (ms *MongoStorage) ConfigureIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Unique index on sessionId
	sessionIdIndex := mongo.IndexModel{
		Keys:    bson.M{"sessionId": 1},
		Options: options.Index().SetUnique(true),
	}

	// TTL index on createdAt, documents expire after 24 hours
	ttlIndex := mongo.IndexModel{
		Keys:    bson.M{"createdAt": 1},
		Options: options.Index().SetExpireAfterSeconds(86400), // 24 hours
	}

	_, err := ms.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{sessionIdIndex, ttlIndex})
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	fmt.Println("MongoDB indexes configured successfully.")
	return nil
}

// LoadSessionData retrieves a session from MongoDB.
func (ms *MongoStorage) LoadSessionData(ctx context.Context, sessionID string) (*models.SessionData, error) {
	var sessionData models.SessionData
	err := ms.collection.FindOne(ctx, bson.M{"sessionId": sessionID}).Decode(&sessionData)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	return &sessionData, nil
}

// CreateSessionData creates a new session document in MongoDB.
func (ms *MongoStorage) CreateSessionData(ctx context.Context, data *models.SessionData) error {
	_, err := ms.collection.InsertOne(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to create session data: %w", err)
	}
	return nil
}

// UpdateSessionData updates an existing session document in MongoDB.
func (ms *MongoStorage) UpdateSessionData(ctx context.Context, data *models.SessionData) error {
	filter := bson.M{"sessionId": data.SessionID}
	update := bson.M{"$set": data}

	_, err := ms.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update session data: %w", err)
	}
	return nil
}

// DeleteSessionData deletes a session from MongoDB.
func (ms *MongoStorage) DeleteSessionData(ctx context.Context, sessionID string) error {
	_, err := ms.collection.DeleteOne(ctx, bson.M{"sessionId": sessionID})
	if err != nil {
		return fmt.Errorf("failed to delete session data: %w", err)
	}
	return nil
}
