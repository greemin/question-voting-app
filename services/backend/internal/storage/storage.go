package storage

import (
	"context"
	"fmt"
	"question-voting-app/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Storer defines the interface for session data storage.
type Storer interface {
	LoadSessionData(sessionID string) (*models.SessionData, error)
	SaveSessionData(data *models.SessionData) error
	DeleteSessionData(sessionID string) error
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

// LoadSessionData retrieves a session from MongoDB.
func (ms *MongoStorage) LoadSessionData(sessionID string) (*models.SessionData, error) {
	var sessionData models.SessionData
	err := ms.collection.FindOne(context.Background(), bson.M{"sessionId": sessionID}).Decode(&sessionData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	return &sessionData, nil
}

// SaveSessionData saves a session to MongoDB, using upsert to create or update.
func (ms *MongoStorage) SaveSessionData(data *models.SessionData) error {
	opts := options.UpdateOne().SetUpsert(true)
	filter := bson.M{"sessionId": data.SessionID}
	update := bson.M{"$set": data}

	_, err := ms.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save session data: %w", err)
	}
	return nil
}

// DeleteSessionData deletes a session from MongoDB.
func (ms *MongoStorage) DeleteSessionData(sessionID string) error {
	_, err := ms.collection.DeleteOne(context.Background(), bson.M{"sessionId": sessionID})
	if err != nil {
		return fmt.Errorf("failed to delete session data: %w", err)
	}
	return nil
}
