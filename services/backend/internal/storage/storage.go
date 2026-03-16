package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"question-voting-app/internal/models"
	"sync"

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

// FileStorage implements the Storer interface for file-based storage.
type FileStorage struct {
	basePath string
	mutex    sync.Mutex
}

// NewFileStorage creates a new instance of FileStorage.
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

// sessionFilePath generates the full path for a session file.
func (fs *FileStorage) sessionFilePath(sessionID string) string {
	return filepath.Join(fs.basePath, fmt.Sprintf("session-%s.json", sessionID))
}

// LoadSessionData reads and parses the JSON file for a given session.
func (fs *FileStorage) LoadSessionData(sessionID string) (*models.SessionData, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := fs.sessionFilePath(sessionID)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var sessionData models.SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &sessionData, nil
}

// SaveSessionData writes the given SessionData struct back to its JSON file.
func (fs *FileStorage) SaveSessionData(data *models.SessionData) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := fs.sessionFilePath(data.SessionID)
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DeleteSessionData deletes the JSON file for a given session.
func (fs *FileStorage) DeleteSessionData(sessionID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := fs.sessionFilePath(sessionID)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, so we can consider it "deleted"
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
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
