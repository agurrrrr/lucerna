package git

import (
	"fmt"
	"time"

	"lucerna/internal/models"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GetTags returns a list of all tags
func (s *RepositoryService) GetTags() ([]models.Tag, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	tagrefs, err := s.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var tags []models.Tag

	err = tagrefs.ForEach(func(ref *plumbing.Reference) error {
		tag := models.Tag{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		}

		// Try to get annotated tag object
		obj, err := s.repo.TagObject(ref.Hash())
		if err == nil {
			// This is an annotated tag
			tag.IsAnnotated = true
			tag.Message = obj.Message
			tag.Tagger = obj.Tagger.Name
			tag.Timestamp = obj.Tagger.When
		} else {
			// This is a lightweight tag, get the commit it points to
			tag.IsAnnotated = false
			commit, err := s.repo.CommitObject(ref.Hash())
			if err == nil {
				tag.Tagger = commit.Author.Name
				tag.Timestamp = commit.Author.When
			}
		}

		tags = append(tags, tag)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// CreateTag creates a new tag
func (s *RepositoryService) CreateTag(tagName, message, commitHash string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Get the commit to tag
	var hash plumbing.Hash
	if commitHash == "" {
		// Tag HEAD
		head, err := s.repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}
		hash = head.Hash()
	} else {
		hash = plumbing.NewHash(commitHash)
	}

	// Create tag reference
	tagRef := plumbing.NewHashReference(
		plumbing.NewTagReferenceName(tagName),
		hash,
	)

	// If message is provided, create an annotated tag
	if message != "" {
		commit, err := s.repo.CommitObject(hash)
		if err != nil {
			return fmt.Errorf("failed to get commit: %w", err)
		}

		// Get git config for tagger info
		cfg, err := s.repo.Config()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		tagger := &object.Signature{
			Name:  cfg.User.Name,
			Email: cfg.User.Email,
			When:  time.Now(),
		}

		// Create annotated tag object
		tagObj := &object.Tag{
			Name:       tagName,
			Tagger:     *tagger,
			Message:    message,
			TargetType: plumbing.CommitObject,
			Target:     commit.Hash,
		}

		// Encode and store the tag object
		obj := s.repo.Storer.NewEncodedObject()
		if err := tagObj.Encode(obj); err != nil {
			return fmt.Errorf("failed to encode tag: %w", err)
		}

		tagHash, err := s.repo.Storer.SetEncodedObject(obj)
		if err != nil {
			return fmt.Errorf("failed to store tag: %w", err)
		}

		// Update the reference to point to the tag object
		tagRef = plumbing.NewHashReference(
			plumbing.NewTagReferenceName(tagName),
			tagHash,
		)
	}

	// Set the tag reference
	err := s.repo.Storer.SetReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// DeleteTag deletes a tag
func (s *RepositoryService) DeleteTag(tagName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Remove the tag reference
	err := s.repo.Storer.RemoveReference(plumbing.NewTagReferenceName(tagName))
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}
