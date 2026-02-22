package models

import "time"

// FileHistoryEntry represents a commit that modified a file
type FileHistoryEntry struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Changes   string    `json:"changes"` // Added, Modified, Deleted
}

// BlameLine represents a line with its blame information
type BlameLine struct {
	LineNumber int       `json:"lineNumber"`
	Content    string    `json:"content"`
	Hash       string    `json:"hash"`
	Author     string    `json:"author"`
	Email      string    `json:"email"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
}
