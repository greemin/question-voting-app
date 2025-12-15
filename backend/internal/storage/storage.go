package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"question-voting-app/internal/models"
	"sync"
)

// dataMutex protects access to the JSON files to prevent race conditions
var dataMutex sync.Mutex

// sessionFilePath generates the full path for a session file.
func sessionFilePath(sessionID string) string {
	return filepath.Join("data", fmt.Sprintf("session%s.json", sessionID))
}

// loadSessionData reads and parses the JSON file for a given session.
func LoadSessionData(sessionID string) (*models.SessionData, error) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filePath := sessionFilePath(sessionID)
	data, err := ioutil.ReadFile(filePath)
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

// saveSessionData writes the given SessionData struct back to its JSON file.
func SaveSessionData(data *models.SessionData) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filePath := sessionFilePath(data.SessionID)
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write the data to the file
	if err := ioutil.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// deleteSessionData deletes the JSON file for a given session.
func DeleteSessionData(sessionID string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filePath := sessionFilePath(sessionID)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
