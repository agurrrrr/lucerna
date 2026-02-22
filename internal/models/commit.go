package models

import "time"

// Commit represents a Git commit
type Commit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Parents   []string  `json:"parents"`
	Refs      []RefInfo `json:"refs,omitempty"`
}

// RefInfo represents a reference (branch, tag, HEAD) pointing to a commit
type RefInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // "head", "branch", "remote", "tag"
	IsRemote bool   `json:"isRemote"`
}

// CommitDetail represents detailed information about a commit
type CommitDetail struct {
	Hash      string     `json:"hash"`
	ShortHash string     `json:"shortHash"`
	Author    string     `json:"author"`
	Email     string     `json:"email"`
	Message   string     `json:"message"`
	Timestamp time.Time  `json:"timestamp"`
	Parents   []Commit   `json:"parents"`
	Stats     []FileStat `json:"stats"`
}

// FileStat represents file change statistics
type FileStat struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Status    string `json:"status"` // added, modified, deleted
}
