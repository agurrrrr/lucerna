package git

import (
	"bufio"
	"fmt"
	"lucerna/internal/models"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
)

// GetWorkingTreeDiff returns diff for working tree changes (unstaged + staged)
func (s *RepositoryService) GetWorkingTreeDiff(filePath string) (*models.FileDiff, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get HEAD commit
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Get HEAD tree
	headTree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Get worktree root path
	_ = worktree.Filesystem.Root()

	// Get status for this file
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	fileStatus := status.File(filePath)
	if fileStatus.Worktree == git.Unmodified && fileStatus.Staging == git.Unmodified {
		return nil, fmt.Errorf("file has no changes")
	}

	// Determine file status
	var diffStatus string
	if fileStatus.Worktree == git.Untracked {
		diffStatus = "added"
	} else if fileStatus.Staging == git.Deleted || fileStatus.Worktree == git.Deleted {
		diffStatus = "deleted"
	} else {
		diffStatus = "modified"
	}

	// Get file content from HEAD
	var oldContent string
	if diffStatus != "added" {
		file, err := headTree.File(filePath)
		if err == nil {
			oldContent, _ = file.Contents()
		}
	}

	// Get current file content
	var newContent string
	if diffStatus != "deleted" {
		fs := worktree.Filesystem
		if file, err := fs.Open(filePath); err == nil {
			defer file.Close()
			buf := make([]byte, 1024*1024) // 1MB max
			if n, err := file.Read(buf); err == nil {
				newContent = string(buf[:n])
			}
		}
	}

	// Generate diff
	diff := generateDiff(oldContent, newContent, filePath)
	diff.Status = diffStatus

	return diff, nil
}

// GetCommitFileDiff returns diff for a specific file in a commit using native git command
func (s *RepositoryService) GetCommitFileDiff(commitHash, filePath string) (*models.FileDiff, error) {
	if s.path == "" {
		return nil, fmt.Errorf("no repository opened")
	}

	// Use git show to get the diff for a specific file in a commit
	// This shows only the changes, not the entire file
	cmd := exec.Command("git", "show", "--format=", "-p", "--unified=3", commitHash, "--", filePath)
	cmd.Dir = s.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	return parseGitDiff(string(output), filePath)
}

// parseGitDiff parses git diff output into FileDiff structure
func parseGitDiff(diffOutput, filePath string) (*models.FileDiff, error) {
	hunks := []models.DiffHunk{}
	var currentHunk *models.DiffHunk
	additions := 0
	deletions := 0
	status := "modified"

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	for scanner.Scan() {
		line := scanner.Text()

		// Detect file status from diff header
		if strings.HasPrefix(line, "new file") {
			status = "added"
			continue
		}
		if strings.HasPrefix(line, "deleted file") {
			status = "deleted"
			continue
		}

		// Skip diff headers
		if strings.HasPrefix(line, "diff --git") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "+++") {
			continue
		}

		// Parse hunk header
		if strings.HasPrefix(line, "@@") {
			// Save previous hunk
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}

			// Parse @@ -oldStart,oldLines +newStart,newLines @@
			currentHunk = &models.DiffHunk{
				Header: line,
				Lines:  []models.DiffLine{},
			}

			// Extract line numbers from header
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				// Parse old range (-X,Y)
				oldRange := strings.TrimPrefix(parts[1], "-")
				if commaIdx := strings.Index(oldRange, ","); commaIdx != -1 {
					currentHunk.OldStart, _ = strconv.Atoi(oldRange[:commaIdx])
					currentHunk.OldLines, _ = strconv.Atoi(oldRange[commaIdx+1:])
				} else {
					currentHunk.OldStart, _ = strconv.Atoi(oldRange)
					currentHunk.OldLines = 1
				}

				// Parse new range (+X,Y)
				newRange := strings.TrimPrefix(parts[2], "+")
				if commaIdx := strings.Index(newRange, ","); commaIdx != -1 {
					currentHunk.NewStart, _ = strconv.Atoi(newRange[:commaIdx])
					currentHunk.NewLines, _ = strconv.Atoi(newRange[commaIdx+1:])
				} else {
					currentHunk.NewStart, _ = strconv.Atoi(newRange)
					currentHunk.NewLines = 1
				}
			}
			continue
		}

		// Skip if no current hunk
		if currentHunk == nil {
			continue
		}

		// Parse diff lines
		var diffLine models.DiffLine
		if strings.HasPrefix(line, "+") {
			diffLine.Type = "add"
			diffLine.Content = strings.TrimPrefix(line, "+")
			additions++
		} else if strings.HasPrefix(line, "-") {
			diffLine.Type = "delete"
			diffLine.Content = strings.TrimPrefix(line, "-")
			deletions++
		} else if strings.HasPrefix(line, " ") {
			diffLine.Type = "context"
			diffLine.Content = strings.TrimPrefix(line, " ")
		} else {
			// Handle lines without prefix (context lines)
			diffLine.Type = "context"
			diffLine.Content = line
		}

		currentHunk.Lines = append(currentHunk.Lines, diffLine)
	}

	// Add last hunk
	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	// Calculate line numbers for each line in hunks
	for i := range hunks {
		oldLine := hunks[i].OldStart
		newLine := hunks[i].NewStart
		for j := range hunks[i].Lines {
			switch hunks[i].Lines[j].Type {
			case "context":
				hunks[i].Lines[j].OldLine = oldLine
				hunks[i].Lines[j].NewLine = newLine
				oldLine++
				newLine++
			case "add":
				hunks[i].Lines[j].NewLine = newLine
				newLine++
			case "delete":
				hunks[i].Lines[j].OldLine = oldLine
				oldLine++
			}
		}
	}

	return &models.FileDiff{
		Path:      filePath,
		Status:    status,
		Additions: additions,
		Deletions: deletions,
		Hunks:     hunks,
		IsBinary:  false,
	}, nil
}

// generateDiff creates a FileDiff from old and new content
func generateDiff(oldContent, newContent, path string) *models.FileDiff {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	hunks := []models.DiffHunk{}
	currentHunk := models.DiffHunk{
		Lines: []models.DiffLine{},
	}

	oldLine := 1
	newLine := 1
	additions := 0
	deletions := 0

	// Collect pending changes to group them (Fork-style: adds first, then deletes)
	var pendingAdds []models.DiffLine
	var pendingDeletes []models.DiffLine

	flushPending := func() {
		// Output adds first, then deletes
		currentHunk.Lines = append(currentHunk.Lines, pendingAdds...)
		currentHunk.Lines = append(currentHunk.Lines, pendingDeletes...)
		pendingAdds = nil
		pendingDeletes = nil
	}

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		var oldText, newText string
		hasOld := i < len(oldLines)
		hasNew := i < len(newLines)

		if hasOld {
			oldText = oldLines[i]
		}
		if hasNew {
			newText = newLines[i]
		}

		if hasOld && hasNew {
			if oldText == newText {
				// Context line - flush any pending changes first
				flushPending()
				currentHunk.Lines = append(currentHunk.Lines, models.DiffLine{
					Type:    "context",
					Content: oldText,
					OldLine: oldLine,
					NewLine: newLine,
				})
				oldLine++
				newLine++
			} else {
				// Modified line - collect for grouping
				pendingAdds = append(pendingAdds, models.DiffLine{
					Type:    "add",
					Content: newText,
					OldLine: 0,
					NewLine: newLine,
				})
				pendingDeletes = append(pendingDeletes, models.DiffLine{
					Type:    "delete",
					Content: oldText,
					OldLine: oldLine,
					NewLine: 0,
				})
				additions++
				deletions++
				oldLine++
				newLine++
			}
		} else if hasOld {
			// Deleted line
			pendingDeletes = append(pendingDeletes, models.DiffLine{
				Type:    "delete",
				Content: oldText,
				OldLine: oldLine,
				NewLine: 0,
			})
			deletions++
			oldLine++
		} else if hasNew {
			// Added line
			pendingAdds = append(pendingAdds, models.DiffLine{
				Type:    "add",
				Content: newText,
				OldLine: 0,
				NewLine: newLine,
			})
			additions++
			newLine++
		}
	}

	// Flush remaining pending changes
	flushPending()

	if len(currentHunk.Lines) > 0 {
		currentHunk.OldStart = 1
		currentHunk.NewStart = 1
		currentHunk.OldLines = len(oldLines)
		currentHunk.NewLines = len(newLines)
		currentHunk.Header = fmt.Sprintf("@@ -%d,%d +%d,%d @@", 1, len(oldLines), 1, len(newLines))
		hunks = append(hunks, currentHunk)
	}

	return &models.FileDiff{
		Path:      path,
		Additions: additions,
		Deletions: deletions,
		Hunks:     hunks,
		IsBinary:  false,
	}
}
