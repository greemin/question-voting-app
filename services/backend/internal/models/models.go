package models

import "time"

// SessionData represents the structure of the data stored in session${sessionId}.json
type SessionData struct {
	SessionTitle string     `json:"sessionTitle" bson:"sessionTitle"`
	SessionID    string     `json:"sessionId" bson:"sessionId"`
	AdminToken   string     `json:"adminToken" bson:"adminToken"`
	IsActive     bool       `json:"isActive" bson:"isActive"`
	CreatedAt    time.Time  `json:"createdAt" bson:"createdAt"`
	Questions    []Question `json:"questions" bson:"questions"`
	BannedIPs    []string   `json:"-" bson:"bannedIPs"`
}

// Question represents a single question submitted by a user
type Question struct {
	ID          string   `json:"id" bson:"id"`
	Text        string   `json:"text" bson:"text"`
	Votes       int      `json:"votes" bson:"votes"`
	Voters      []string `json:"voters" bson:"voters"` // userSessionIds who have voted
	SubmitterIP string   `json:"-" bson:"submitterIP"`
}

// QuestionSubmission is used for the POST request body
type QuestionSubmission struct {
	Text string `json:"text"`
}

// CreateSessionRequest is used for the POST /api/session request body
type CreateSessionRequest struct {
	SessionID string `json:"sessionId"`
}
