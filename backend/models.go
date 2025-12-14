package main

// SessionData represents the structure of the data stored in session${sessionId}.json
type SessionData struct {
	SessionID   string     `json:"sessionId"`
	AdminUserID string     `json:"adminUserId"`
	IsActive    bool       `json:"isActive"`
	Questions   []Question `json:"questions"`
}

// Question represents a single question submitted by a user
type Question struct {
	ID     string   `json:"id"`
	Text   string   `json:"text"`
	Votes  int      `json:"votes"`
	Voters []string `json:"voters"` // userSessionIds who have voted
}

// QuestionSubmission is used for the POST request body
type QuestionSubmission struct {
	Text string `json:"text"`
}
