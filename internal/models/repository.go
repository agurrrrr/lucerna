package models

import "time"

// Repository represents a Git repository
type Repository struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	CurrentBranch string   `json:"currentBranch"`
	LastOpened   time.Time `json:"lastOpened"`
	IsBare       bool      `json:"isBare"`
}

// RepositoryStatus represents the current status of a repository
type RepositoryStatus struct {
	Branch         string   `json:"branch"`
	Ahead          int      `json:"ahead"`
	Behind         int      `json:"behind"`
	StagedFiles    []string `json:"stagedFiles"`
	ModifiedFiles  []string `json:"modifiedFiles"`
	UntrackedFiles []string `json:"untrackedFiles"`
}
