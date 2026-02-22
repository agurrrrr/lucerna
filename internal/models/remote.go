package models

// Remote represents a git remote
type Remote struct {
	Name string   `json:"name"`
	URLs []string `json:"urls"`
}
