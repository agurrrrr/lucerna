package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"lucerna/internal/models"
)

// GetStashes returns a list of all stashes
func (s *RepositoryService) GetStashes() ([]models.Stash, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Use git CLI for stash operations (go-git doesn't support stash well)
	cmd := exec.Command("git", "stash", "list", "--format=%gd|%s|%gs|%H|%ct")
	cmd.Dir = worktree.Filesystem.Root()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list stashes: %w", err)
	}

	if len(output) == 0 {
		return []models.Stash{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	stashes := make([]models.Stash, 0, len(lines))

	// Parse stash entries
	// Format: stash@{0}|message|branch|hash|timestamp
	stashRegex := regexp.MustCompile(`stash@\{(\d+)\}`)

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		// Extract stash index
		matches := stashRegex.FindStringSubmatch(parts[0])
		if len(matches) < 2 {
			continue
		}

		index, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		// Parse timestamp
		timestampInt, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			continue
		}
		timestamp := time.Unix(timestampInt, 0)

		stashes = append(stashes, models.Stash{
			Index:     index,
			Message:   parts[1],
			Branch:    parts[2],
			Hash:      parts[3],
			Timestamp: timestamp,
		})
	}

	return stashes, nil
}

// CreateStash creates a new stash with a message
func (s *RepositoryService) CreateStash(message string, includeUntracked bool) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	if includeUntracked {
		args = append(args, "-u")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = worktree.Filesystem.Root()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create stash: %w, output: %s", err, string(output))
	}

	return nil
}

// ApplyStash applies a stash by index
func (s *RepositoryService) ApplyStash(index int, pop bool) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	stashRef := fmt.Sprintf("stash@{%d}", index)

	var cmd *exec.Cmd
	if pop {
		cmd = exec.Command("git", "stash", "pop", stashRef)
	} else {
		cmd = exec.Command("git", "stash", "apply", stashRef)
	}

	cmd.Dir = worktree.Filesystem.Root()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply stash: %w, output: %s", err, string(output))
	}

	return nil
}

// DropStash removes a stash by index
func (s *RepositoryService) DropStash(index int) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	stashRef := fmt.Sprintf("stash@{%d}", index)
	cmd := exec.Command("git", "stash", "drop", stashRef)
	cmd.Dir = worktree.Filesystem.Root()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to drop stash: %w, output: %s", err, string(output))
	}

	return nil
}

// ClearStashes removes all stashes
func (s *RepositoryService) ClearStashes() error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	cmd := exec.Command("git", "stash", "clear")
	cmd.Dir = worktree.Filesystem.Root()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clear stashes: %w, output: %s", err, string(output))
	}

	return nil
}
