package git

import (
	"bufio"
	"fmt"
	"lucerna/internal/models"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MergeBranchWithResult performs a merge and returns detailed result
func (s *RepositoryService) MergeBranchWithResult(branchName string) (*models.MergeResult, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty")
	}

	// Execute git merge --no-commit to allow inspection before commit
	cmd := exec.Command("git", "merge", "--no-commit", "--no-ff", branchName)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// Check if merge was successful
	if err == nil {
		// Check if it was a fast-forward (even with --no-ff, detect if possible)
		isFastForward := strings.Contains(outputStr, "Fast-forward")

		// Auto-commit for clean merges
		commitCmd := exec.Command("git", "commit", "--no-edit")
		commitCmd.Dir = s.path
		commitCmd.Run() // Ignore error, might fail if nothing to commit (fast-forward)

		return &models.MergeResult{
			Success:     true,
			FastForward: isFastForward,
			Message:     fmt.Sprintf("Successfully merged %s", branchName),
		}, nil
	}

	// Check if there are conflicts
	if strings.Contains(outputStr, "CONFLICT") || strings.Contains(outputStr, "Automatic merge failed") {
		// Get list of conflicted files
		conflictFiles, conflictErr := s.GetConflictFiles()
		if conflictErr != nil {
			return nil, fmt.Errorf("merge has conflicts but failed to get conflict list: %w", conflictErr)
		}

		return &models.MergeResult{
			Success:       false,
			HasConflicts:  true,
			ConflictFiles: conflictFiles,
			Message:       "Merge has conflicts that need to be resolved",
		}, nil
	}

	// Other merge error
	return nil, fmt.Errorf("merge failed: %s", outputStr)
}

// GetConflictFiles returns list of files with conflicts
func (s *RepositoryService) GetConflictFiles() ([]models.ConflictFile, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	// Get unmerged files
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict files: %w", err)
	}

	var files []models.ConflictFile
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		if path != "" {
			files = append(files, models.ConflictFile{
				Path:     path,
				Resolved: false,
			})
		}
	}

	return files, nil
}

// GetThreeWayDiff returns 3-way diff for a conflicted file
func (s *RepositoryService) GetThreeWayDiff(filePath string) (*models.ThreeWayDiff, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	diff := &models.ThreeWayDiff{
		Path: filePath,
	}

	// Get base version (stage 1 - common ancestor)
	baseCmd := exec.Command("git", "show", ":1:"+filePath)
	baseCmd.Dir = s.path
	baseOutput, err := baseCmd.Output()
	if err == nil {
		diff.Base = string(baseOutput)
	} else {
		diff.Base = "" // File might not exist in base
	}

	// Get ours version (stage 2 - current branch/HEAD)
	oursCmd := exec.Command("git", "show", ":2:"+filePath)
	oursCmd.Dir = s.path
	oursOutput, err := oursCmd.Output()
	if err == nil {
		diff.Ours = string(oursOutput)
	} else {
		diff.Ours = "" // File might not exist in ours
	}

	// Get theirs version (stage 3 - branch being merged)
	theirsCmd := exec.Command("git", "show", ":3:"+filePath)
	theirsCmd.Dir = s.path
	theirsOutput, err := theirsCmd.Output()
	if err == nil {
		diff.Theirs = string(theirsOutput)
	} else {
		diff.Theirs = "" // File might not exist in theirs
	}

	// Read current working file (with conflict markers)
	workingFilePath := filepath.Join(s.path, filePath)
	mergedContent, err := os.ReadFile(workingFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read working file: %w", err)
	}
	diff.Merged = string(mergedContent)

	return diff, nil
}

// ResolveConflictFile saves the resolved content and stages the file
func (s *RepositoryService) ResolveConflictFile(filePath string, resolvedContent string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	// Validate file path to prevent directory traversal
	workingFilePath := filepath.Join(s.path, filePath)
	absWorkingPath, err := filepath.Abs(workingFilePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	absRepoPath, err := filepath.Abs(s.path)
	if err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}
	if !strings.HasPrefix(absWorkingPath, absRepoPath+string(filepath.Separator)) {
		return fmt.Errorf("file path is outside repository: %s", filePath)
	}

	// Write resolved content to file
	err = os.WriteFile(workingFilePath, []byte(resolvedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write resolved file: %w", err)
	}

	// Stage the file
	cmd := exec.Command("git", "add", filePath)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage resolved file: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// CompleteMerge completes the merge after resolving all conflicts
func (s *RepositoryService) CompleteMerge(message string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	// Check if there are remaining conflicts
	conflictFiles, err := s.GetConflictFiles()
	if err != nil {
		return fmt.Errorf("failed to check for remaining conflicts: %w", err)
	}

	if len(conflictFiles) > 0 {
		return fmt.Errorf("there are still %d unresolved conflict(s)", len(conflictFiles))
	}

	// Commit the merge
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to complete merge: %s", strings.TrimSpace(string(output)))
	}

	// Refresh the repository to update go-git state
	s.RefreshRepository()

	return nil
}

// AbortMerge cancels the merge operation
func (s *RepositoryService) AbortMerge() error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to abort merge: %s", strings.TrimSpace(string(output)))
	}

	// Refresh the repository to update go-git state
	s.RefreshRepository()

	return nil
}

// IsMergeInProgress checks if a merge is currently in progress
func (s *RepositoryService) IsMergeInProgress() (bool, error) {
	if s.path == "" {
		return false, fmt.Errorf("no repository opened")
	}

	// Check if .git/MERGE_HEAD exists
	mergeHeadPath := filepath.Join(s.path, ".git", "MERGE_HEAD")
	_, err := os.Stat(mergeHeadPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check merge status: %w", err)
}

// GetMergeSourceBranch returns the name of the branch being merged (from MERGE_MSG)
func (s *RepositoryService) GetMergeSourceBranch() (string, error) {
	if s.path == "" {
		return "", fmt.Errorf("no repository opened")
	}

	// Read MERGE_MSG to extract branch name
	mergeMsgPath := filepath.Join(s.path, ".git", "MERGE_MSG")
	content, err := os.ReadFile(mergeMsgPath)
	if err != nil {
		return "", fmt.Errorf("failed to read MERGE_MSG: %w", err)
	}

	// Parse "Merge branch 'branchname'" from the message
	msgStr := string(content)
	if strings.HasPrefix(msgStr, "Merge branch '") {
		start := len("Merge branch '")
		end := strings.Index(msgStr[start:], "'")
		if end > 0 {
			return msgStr[start : start+end], nil
		}
	}

	return "", nil
}
