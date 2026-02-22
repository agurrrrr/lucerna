package models

import "time"

// Stash represents a git stash entry
type Stash struct {
	Index     int       `json:"index"`
	Message   string    `json:"message"`
	Branch    string    `json:"branch"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}
