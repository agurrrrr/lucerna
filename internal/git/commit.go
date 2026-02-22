package git

import (
	"bufio"
	"fmt"
	"lucerna/internal/models"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetCommits returns commit history with optional limit using native git command
func (s *RepositoryService) GetCommits(limit int) ([]models.Commit, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	// First, get all refs (branches, tags, HEAD) mapped to commit hashes
	refMap := s.getRefMap()

	// Use git log with custom format for easy parsing
	// Format: hash|short_hash|author|email|timestamp|parents|message
	format := "%H|%h|%an|%ae|%at|%P|%s"
	args := []string{"log", "--all", fmt.Sprintf("--format=%s", format)}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-n%d", limit))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	commits := []models.Commit{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 7)
		if len(parts) < 7 {
			continue
		}

		hash := parts[0]
		shortHash := parts[1]
		author := parts[2]
		email := parts[3]
		timestampStr := parts[4]
		parentsStr := parts[5]
		message := parts[6]

		// Parse timestamp
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		commitTime := time.Unix(timestamp, 0)

		// Parse parents
		var parents []string
		if parentsStr != "" {
			parents = strings.Split(parentsStr, " ")
		}

		// Get refs for this commit
		refs := refMap[hash]

		commits = append(commits, models.Commit{
			Hash:      hash,
			ShortHash: shortHash,
			Author:    author,
			Email:     email,
			Message:   message,
			Timestamp: commitTime,
			Parents:   parents,
			Refs:      refs,
		})
	}

	return commits, nil
}

// getRefMap returns a map of commit hash to refs pointing to it
func (s *RepositoryService) getRefMap() map[string][]models.RefInfo {
	refMap := make(map[string][]models.RefInfo)

	// Get HEAD hash and current branch name
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = s.path
	headOutput, err := headCmd.Output()
	var headHash string
	var currentBranch string
	if err == nil {
		headHash = strings.TrimSpace(string(headOutput))

		// Get current branch name (empty if detached HEAD)
		branchCmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
		branchCmd.Dir = s.path
		branchOutput, branchErr := branchCmd.Output()
		if branchErr == nil {
			currentBranch = strings.TrimSpace(string(branchOutput))
		}
	}

	// Add HEAD marker at HEAD position (for detached HEAD or to show current position)
	if headHash != "" {
		refMap[headHash] = append(refMap[headHash], models.RefInfo{
			Name: "HEAD",
			Type: "head",
		})
	}

	// Get all refs using for-each-ref
	// Format: hash refname reftype
	cmd := exec.Command("git", "for-each-ref", "--format=%(objectname) %(refname:short) %(refname)", "refs/heads", "refs/remotes", "refs/tags")
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return refMap
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}

		hash := parts[0]
		shortName := parts[1]
		fullName := parts[2]

		var refInfo models.RefInfo
		refInfo.Name = shortName

		if strings.HasPrefix(fullName, "refs/heads/") {
			refInfo.Type = "branch"
			refInfo.IsRemote = false
			// Mark current branch as head type (shows with HEAD styling)
			if currentBranch != "" && shortName == currentBranch {
				refInfo.Type = "head"
			}
		} else if strings.HasPrefix(fullName, "refs/remotes/") {
			refInfo.Type = "remote"
			refInfo.IsRemote = true
		} else if strings.HasPrefix(fullName, "refs/tags/") {
			refInfo.Type = "tag"
			refInfo.IsRemote = false
		}

		refMap[hash] = append(refMap[hash], refInfo)
	}

	return refMap
}

// GetCommitDetail returns detailed information about a specific commit using native git command
func (s *RepositoryService) GetCommitDetail(hash string) (*models.CommitDetail, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	// Get commit info using git show
	format := "%H|%h|%an|%ae|%at|%P|%B"
	cmd := exec.Command("git", "show", "-s", fmt.Sprintf("--format=%s", format), hash)
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	lines := strings.SplitN(string(output), "|", 7)
	if len(lines) < 7 {
		return nil, fmt.Errorf("invalid commit format")
	}

	commitHash := lines[0]
	shortHash := lines[1]
	author := lines[2]
	email := lines[3]
	timestampStr := lines[4]
	parentsStr := lines[5]
	message := strings.TrimSpace(lines[6])

	// Parse timestamp
	timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
	commitTime := time.Unix(timestamp, 0)

	// Parse parent hashes
	var parentHashes []string
	if parentsStr != "" {
		parentHashes = strings.Split(parentsStr, " ")
	}

	// Get parent commits info
	parents := []models.Commit{}
	for _, parentHash := range parentHashes {
		parentCmd := exec.Command("git", "show", "-s", "--format=%H|%h|%an|%ae|%at|%s", parentHash)
		parentCmd.Dir = s.path
		parentOutput, err := parentCmd.Output()
		if err != nil {
			continue
		}

		parentParts := strings.SplitN(strings.TrimSpace(string(parentOutput)), "|", 6)
		if len(parentParts) >= 6 {
			parentTs, _ := strconv.ParseInt(parentParts[4], 10, 64)
			parents = append(parents, models.Commit{
				Hash:      parentParts[0],
				ShortHash: parentParts[1],
				Author:    parentParts[2],
				Email:     parentParts[3],
				Message:   parentParts[5],
				Timestamp: time.Unix(parentTs, 0),
			})
		}
	}

	// Get file stats using git diff-tree
	stats := s.getCommitStatsNative(hash)

	return &models.CommitDetail{
		Hash:      commitHash,
		ShortHash: shortHash,
		Author:    author,
		Email:     email,
		Message:   message,
		Timestamp: commitTime,
		Parents:   parents,
		Stats:     stats,
	}, nil
}

// getCommitStatsNative gets file change statistics for a commit using native git command
func (s *RepositoryService) getCommitStatsNative(hash string) []models.FileStat {
	stats := []models.FileStat{}

	// Use git diff-tree to get file changes with stats
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--numstat", "-r", hash)
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return stats
	}

	// Get status (added/modified/deleted) separately
	statusCmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-status", "-r", hash)
	statusCmd.Dir = s.path
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return stats
	}

	// Build status map
	statusMap := make(map[string]string)
	statusScanner := bufio.NewScanner(strings.NewReader(string(statusOutput)))
	for statusScanner.Scan() {
		line := statusScanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			statusCode := parts[0]
			filePath := parts[len(parts)-1]
			switch statusCode[0] {
			case 'A':
				statusMap[filePath] = "added"
			case 'D':
				statusMap[filePath] = "deleted"
			case 'M':
				statusMap[filePath] = "modified"
			case 'R':
				statusMap[filePath] = "renamed"
			default:
				statusMap[filePath] = "modified"
			}
		}
	}

	// Parse numstat output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			additions, _ := strconv.Atoi(parts[0])
			deletions, _ := strconv.Atoi(parts[1])
			filePath := parts[2]

			status := statusMap[filePath]
			if status == "" {
				status = "modified"
			}

			stats = append(stats, models.FileStat{
				Path:      filePath,
				Additions: additions,
				Deletions: deletions,
				Status:    status,
			})
		}
	}

	return stats
}
