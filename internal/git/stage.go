package git

import (
	"fmt"
	"os/exec"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

// StageFile adds a file to the staging area
func (s *RepositoryService) StageFile(filePath string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Add(filePath)
	if err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	return nil
}

// UnstageFile removes a file from the staging area
func (s *RepositoryService) UnstageFile(filePath string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	cmd := exec.Command("git", "reset", "HEAD", "--", filePath)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unstage file: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// StageAll adds all changed files to the staging area
func (s *RepositoryService) StageAll() error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = worktree.Add(".")
	if err != nil {
		return fmt.Errorf("failed to stage all: %w", err)
	}

	return nil
}

// UnstageAll removes all files from the staging area
func (s *RepositoryService) UnstageAll() error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Reset(&gogit.ResetOptions{
		Mode: gogit.MixedReset,
	})
	if err != nil {
		return fmt.Errorf("failed to unstage all: %w", err)
	}

	return nil
}

// CreateCommit creates a new commit with the staged changes
func (s *RepositoryService) CreateCommit(message, authorName, authorEmail string) (string, error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository opened")
	}

	if message == "" {
		return "", fmt.Errorf("commit message cannot be empty")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Check if there are staged changes
	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	hasStaged := false
	for _, fileStatus := range status {
		if fileStatus.Staging != gogit.Unmodified {
			hasStaged = true
			break
		}
	}

	if !hasStaged {
		return "", fmt.Errorf("no staged changes to commit")
	}

	// Create commit
	hash, err := worktree.Commit(message, &gogit.CommitOptions{
		Author: nil, // Will use git config
	})
	if err != nil {
		return "", fmt.Errorf("failed to create commit: %w", err)
	}

	return hash.String(), nil
}
