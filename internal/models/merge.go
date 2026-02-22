package models

// MergeResult represents the result of a merge operation
type MergeResult struct {
	Success       bool           `json:"success"`
	FastForward   bool           `json:"fastForward"`
	HasConflicts  bool           `json:"hasConflicts"`
	ConflictFiles []ConflictFile `json:"conflictFiles,omitempty"`
	Message       string         `json:"message"`
}

// ConflictFile represents a file with merge conflicts
type ConflictFile struct {
	Path     string `json:"path"`
	Resolved bool   `json:"resolved"`
}

// ThreeWayDiff represents 3-way diff for conflict resolution
type ThreeWayDiff struct {
	Path   string `json:"path"`
	Base   string `json:"base"`   // Common ancestor content
	Ours   string `json:"ours"`   // Current branch content (HEAD)
	Theirs string `json:"theirs"` // Branch being merged content
	Merged string `json:"merged"` // Current working file content (with conflict markers)
}
