package models

// FileDiff represents diff information for a file
type FileDiff struct {
	Path      string      `json:"path"`
	OldPath   string      `json:"oldPath,omitempty"`
	Status    string      `json:"status"` // added, modified, deleted, renamed
	Additions int         `json:"additions"`
	Deletions int         `json:"deletions"`
	Hunks     []DiffHunk  `json:"hunks"`
	IsBinary  bool        `json:"isBinary"`
}

// DiffHunk represents a hunk in a diff
type DiffHunk struct {
	OldStart int        `json:"oldStart"`
	OldLines int        `json:"oldLines"`
	NewStart int        `json:"newStart"`
	NewLines int        `json:"newLines"`
	Header   string     `json:"header"`
	Lines    []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    string `json:"type"`    // add, delete, context
	Content string `json:"content"`
	OldLine int    `json:"oldLine"` // 0 if not applicable
	NewLine int    `json:"newLine"` // 0 if not applicable
}
