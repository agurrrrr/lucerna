package models

import "time"

// Tag represents a git tag
type Tag struct {
	Name      string    `json:"name"`
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Tagger    string    `json:"tagger"`
	Timestamp time.Time `json:"timestamp"`
	IsAnnotated bool    `json:"isAnnotated"`
}
