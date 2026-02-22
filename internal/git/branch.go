package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CreateBranch creates a new branch
func (s *RepositoryService) CreateBranch(branchName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Get HEAD reference
	head, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, head.Hash())

	err = s.repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch
func (s *RepositoryService) DeleteBranch(branchName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check if it's the current branch
	head, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().Short() == branchName {
		return fmt.Errorf("cannot delete current branch")
	}

	// Delete branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	err = s.repo.Storer.RemoveReference(refName)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// CheckoutBranch switches to a different branch
func (s *RepositoryService) CheckoutBranch(branchName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// MergeBranch merges a branch into the current branch
func (s *RepositoryService) MergeBranch(branchName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Get current HEAD
	head, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get branch reference
	branchRef, err := s.repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		return fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get commit objects
	headCommit, err := s.repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	branchCommit, err := s.repo.CommitObject(branchRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get branch commit: %w", err)
	}

	// Get worktree
	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// go-git doesn't have a built-in Merge method on worktree
	// For now, we'll do a simple fast-forward merge by updating the branch reference
	// Full 3-way merge requires more complex implementation

	// Check if we can fast-forward (HEAD is ancestor of branch)
	isAncestor, err := headCommit.IsAncestor(branchCommit)
	if err != nil {
		return fmt.Errorf("failed to check ancestry: %w", err)
	}

	if !isAncestor {
		return fmt.Errorf("cannot fast-forward merge: branches have diverged")
	}

	// Update HEAD to point to merged commit (fast-forward)
	err = s.repo.Storer.SetReference(plumbing.NewHashReference(head.Name(), branchRef.Hash()))
	if err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}

	// Reset worktree to the new HEAD
	err = worktree.Reset(&git.ResetOptions{
		Commit: branchRef.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset worktree: %w", err)
	}

	return nil
}

// CreateBranchFromCommit creates a new branch from a specific commit
func (s *RepositoryService) CreateBranchFromCommit(branchName, commitHash string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Parse commit hash
	hash := plumbing.NewHash(commitHash)

	// Verify commit exists
	_, err := s.repo.CommitObject(hash)
	if err != nil {
		return fmt.Errorf("commit not found: %w", err)
	}

	// Create branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, hash)

	err = s.repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// GetCurrentBranch returns the name of the current branch
func (s *RepositoryService) GetCurrentBranch() (string, error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository opened")
	}

	head, err := s.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Name().Short(), nil
}

// CheckoutCommit checks out a specific commit (detached HEAD)
func (s *RepositoryService) CheckoutCommit(commitHash string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if commitHash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	// Use native git checkout
	cmd := exec.Command("git", "checkout", commitHash)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout commit: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// RenameBranch renames a branch
func (s *RepositoryService) RenameBranch(oldName, newName string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if oldName == "" || newName == "" {
		return fmt.Errorf("branch names cannot be empty")
	}

	// Get old branch reference
	oldRef, err := s.repo.Reference(plumbing.NewBranchReferenceName(oldName), true)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	// Create new reference
	newRefName := plumbing.NewBranchReferenceName(newName)
	newRef := plumbing.NewHashReference(newRefName, oldRef.Hash())

	err = s.repo.Storer.SetReference(newRef)
	if err != nil {
		return fmt.Errorf("failed to create new branch: %w", err)
	}

	// Check if it's the current branch and update HEAD if needed
	head, err := s.repo.Head()
	if err == nil && head.Name().Short() == oldName {
		// Update HEAD to point to new branch
		err = s.repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, newRefName))
		if err != nil {
			return fmt.Errorf("failed to update HEAD: %w", err)
		}
	}

	// Delete old reference
	oldRefName := plumbing.NewBranchReferenceName(oldName)
	err = s.repo.Storer.RemoveReference(oldRefName)
	if err != nil {
		return fmt.Errorf("failed to delete old branch: %w", err)
	}

	return nil
}

// CherryPickCommit cherry-picks a specific commit onto the current branch
func (s *RepositoryService) CherryPickCommit(commitHash string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if commitHash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	cmd := exec.Command("git", "cherry-pick", commitHash)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to cherry-pick commit: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

// RevertCommit reverts a specific commit
func (s *RepositoryService) RevertCommit(commitHash string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if commitHash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	cmd := exec.Command("git", "revert", "--no-edit", commitHash)
	cmd.Dir = s.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to revert commit: %s", strings.TrimSpace(string(output)))
	}

	return nil
}
