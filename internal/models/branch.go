package models

import "time"

// Branch represents a Git branch
type Branch struct {
	Name      string    `json:"name"`
	IsHead    bool      `json:"isHead"`
	IsRemote  bool      `json:"isRemote"`
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
