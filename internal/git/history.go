package git

import (
	"fmt"
	"io"

	"lucerna/internal/models"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GetFileHistory returns the commit history for a specific file
func (s *RepositoryService) GetFileHistory(filePath string, limit int) ([]models.FileHistoryEntry, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	if limit <= 0 {
		limit = 100
	}

	// Get commit log
	commits, err := s.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		PathFilter: func(path string) bool {
			return path == filePath
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}

	var history []models.FileHistoryEntry
	count := 0

	err = commits.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return io.EOF
		}

		// Check what happened to the file in this commit
		changes := "Modified"

		// Check if this is the first commit (file was added)
		if c.NumParents() == 0 {
			changes = "Added"
		} else {
			// Check if file was deleted
			parent, err := c.Parent(0)
			if err == nil {
				_, err1 := c.File(filePath)
				_, err2 := parent.File(filePath)

				if err1 != nil && err2 == nil {
					changes = "Deleted"
				} else if err1 == nil && err2 != nil {
					changes = "Added"
				}
			}
		}

		entry := models.FileHistoryEntry{
			Hash:      c.Hash.String(),
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Message:   c.Message,
			Timestamp: c.Author.When,
			Changes:   changes,
		}

		history = append(history, entry)
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return history, nil
}

// GetFileBlame returns blame information for a file
func (s *RepositoryService) GetFileBlame(filePath string) ([]models.BlameLine, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
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

	// Get blame for the file
	blame, err := git.Blame(commit, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame: %w", err)
	}

	var blameLines []models.BlameLine

	for i, line := range blame.Lines {
		// Get commit info for this line
		lineCommit, err := s.repo.CommitObject(line.Hash)
		if err != nil {
			continue
		}

		blameLine := models.BlameLine{
			LineNumber: i + 1,
			Content:    line.Text,
			Hash:       line.Hash.String(),
			Author:     line.Author,
			Email:      lineCommit.Author.Email,
			Timestamp:  line.Date,
			Message:    lineCommit.Message,
		}

		blameLines = append(blameLines, blameLine)
	}

	return blameLines, nil
}
