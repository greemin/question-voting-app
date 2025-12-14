package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

// dataMutex protects access to the JSON files to prevent race conditions
var dataMutex sync.Mutex

func getFilename(sessionID string) string {
	return fmt.Sprintf("session%s.json", sessionID)
}

// loadSessionData reads and parses the JSON file for a given session.
func loadSessionData(sessionID string) (*SessionData, error) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filename := getFilename(sessionID)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &sessionData, nil
}

// saveSessionData writes the given SessionData struct back to its JSON file.
func saveSessionData(data *SessionData) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filename := getFilename(data.SessionID)
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write the data to the file
	if err := ioutil.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// deleteSessionData deletes the JSON file for a given session.
func deleteSessionData(sessionID string) error {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	filename := getFilename(sessionID)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
