package git

import (
	"bufio"
	"fmt"
	"lucerna/internal/models"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// RepositoryService handles Git repository operations
type RepositoryService struct {
	mu   sync.RWMutex
	repo *git.Repository
	path string
}

// NewRepositoryService creates a new repository service
func NewRepositoryService() *RepositoryService {
	return &RepositoryService{}
}

// RefreshRepository re-opens the repository to get fresh refs
func (s *RepositoryService) RefreshRepository() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	// Re-open the repository to refresh internal state
	repo, err := git.PlainOpen(s.path)
	if err != nil {
		return fmt.Errorf("failed to refresh repository: %w", err)
	}

	s.repo = repo
	return nil
}

// OpenRepository opens a Git repository at the given path
func (s *RepositoryService) OpenRepository(path string) (*models.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	s.repo = repo
	s.path = path

	// Get repository info
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get repository name from path
	name := filepath.Base(path)

	// Get current branch name
	branchName := head.Name().Short()

	return &models.Repository{
		Path:          path,
		Name:          name,
		CurrentBranch: branchName,
		LastOpened:    time.Now(),
		IsBare:        false,
	}, nil
}

// GetStatus returns the current status of the repository using native git command
func (s *RepositoryService) GetStatus() (*models.RepositoryStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	// Get current branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = s.path
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get status using porcelain format for easy parsing
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = s.path
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var staged, modified, untracked []string

	scanner := bufio.NewScanner(strings.NewReader(string(statusOutput)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}

		indexStatus := line[0]
		workTreeStatus := line[1]
		filePath := strings.TrimSpace(line[3:])

		// Handle renamed files (format: "R  old -> new")
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			if len(parts) == 2 {
				filePath = parts[1]
			}
		}

		// Staged files (index has changes)
		if indexStatus != ' ' && indexStatus != '?' {
			staged = append(staged, filePath)
		}

		// Modified files (worktree has changes, not staged)
		if workTreeStatus == 'M' || workTreeStatus == 'D' {
			modified = append(modified, filePath)
		}

		// Untracked files
		if indexStatus == '?' && workTreeStatus == '?' {
			untracked = append(untracked, filePath)
		}
	}

	return &models.RepositoryStatus{
		Branch:         branch,
		Ahead:          0,
		Behind:         0,
		StagedFiles:    staged,
		ModifiedFiles:  modified,
		UntrackedFiles: untracked,
	}, nil
}

// GetBranches returns all branches in the repository
func (s *RepositoryService) GetBranches() ([]models.Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	branches := []models.Branch{}

	// Get current branch
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	currentBranch := head.Name().Short()

	// Get all branches
	branchIter, err := s.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()

		// Get commit for this branch
		commit, err := s.repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		branches = append(branches, models.Branch{
			Name:      branchName,
			IsHead:    branchName == currentBranch,
			IsRemote:  false,
			Hash:      ref.Hash().String(),
			Author:    commit.Author.Name,
			Message:   commit.Message,
			Timestamp: commit.Author.When,
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branches, nil
}

// IsValidRepository checks if a path contains a valid Git repository
func IsValidRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
