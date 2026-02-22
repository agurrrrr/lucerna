package main

import (
	"context"
	"encoding/json"
	"fmt"
	"lucerna/internal/git"
	"lucerna/internal/models"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	gitSvc  *git.RepositoryService
	repoDir string // Current repository directory
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		gitSvc: git.NewRepositoryService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to Lucerna!", name)
}

// SelectRepositoryDirectory opens a directory selector dialog
func (a *App) SelectRepositoryDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Git Repository",
	})
	if err != nil {
		return "", err
	}

	if dir == "" {
		return "", nil // User cancelled
	}

	return dir, nil
}

// OpenRepository opens a Git repository
func (a *App) OpenRepository(path string) (string, error) {
	repo, err := a.gitSvc.OpenRepository(path)
	if err != nil {
		return "", err
	}

	a.repoDir = path

	// Convert to JSON
	data, err := json.Marshal(repo)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetRepositoryStatus gets the current repository status
func (a *App) GetRepositoryStatus() (string, error) {
	status, err := a.gitSvc.GetStatus()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(status)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetBranches gets all branches in the repository
func (a *App) GetBranches() (string, error) {
	branches, err := a.gitSvc.GetBranches()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(branches)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// IsValidRepository checks if a path is a valid Git repository
func (a *App) IsValidRepository(path string) bool {
	return git.IsValidRepository(path)
}

// GetRecentRepositories gets recently opened repositories from config
func (a *App) GetRecentRepositories() (string, error) {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "[]", nil
	}

	configPath := filepath.Join(home, ".lucerna", "recent.json")

	// If config doesn't exist, return empty array
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "[]", nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "[]", nil
	}

	return string(data), nil
}

// SaveRecentRepository saves a repository to recent list
func (a *App) SaveRecentRepository(repoJSON string) error {
	var repo models.Repository
	if err := json.Unmarshal([]byte(repoJSON), &repo); err != nil {
		return err
	}

	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".lucerna")
	configPath := filepath.Join(configDir, "recent.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Read existing recent repos
	var recent []models.Repository
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &recent); err != nil {
			recent = []models.Repository{}
		}
	}

	// Check if repo already exists
	exists := false
	for _, r := range recent {
		if r.Path == repo.Path {
			exists = true
			break
		}
	}

	var filtered []models.Repository
	if exists {
		// Keep existing order, just update the repo info
		for _, r := range recent {
			if r.Path == repo.Path {
				filtered = append(filtered, repo)
			} else {
				filtered = append(filtered, r)
			}
		}
	} else {
		// New repo, add to the end
		filtered = append(recent, repo)
		if len(filtered) > 10 {
			filtered = filtered[len(filtered)-10:]
		}
	}

	// Write back to file
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ReorderRecentRepositories reorders the recent repositories list
func (a *App) ReorderRecentRepositories(reposJSON string) error {
	var repos []models.Repository
	if err := json.Unmarshal([]byte(reposJSON), &repos); err != nil {
		return err
	}

	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".lucerna")
	configPath := filepath.Join(configDir, "recent.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Write the new order to file
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetCommits gets commit history with optional limit
func (a *App) GetCommits(limit int) (string, error) {
	commits, err := a.gitSvc.GetCommits(limit)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(commits)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetCommitDetail gets detailed information about a specific commit
func (a *App) GetCommitDetail(hash string) (string, error) {
	detail, err := a.gitSvc.GetCommitDetail(hash)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(detail)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetWorkingTreeDiff gets diff for a file in working tree
func (a *App) GetWorkingTreeDiff(filePath string) (string, error) {
	diff, err := a.gitSvc.GetWorkingTreeDiff(filePath)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(diff)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetCommitFileDiff gets diff for a file in a specific commit
func (a *App) GetCommitFileDiff(commitHash, filePath string) (string, error) {
	diff, err := a.gitSvc.GetCommitFileDiff(commitHash, filePath)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(diff)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// StageFile adds a file to the staging area
func (a *App) StageFile(filePath string) error {
	return a.gitSvc.StageFile(filePath)
}

// UnstageFile removes a file from the staging area
func (a *App) UnstageFile(filePath string) error {
	return a.gitSvc.UnstageFile(filePath)
}

// StageAll adds all changed files to the staging area
func (a *App) StageAll() error {
	return a.gitSvc.StageAll()
}

// UnstageAll removes all files from the staging area
func (a *App) UnstageAll() error {
	return a.gitSvc.UnstageAll()
}

// CreateCommit creates a new commit with the staged changes
func (a *App) CreateCommit(message, authorName, authorEmail string) (string, error) {
	hash, err := a.gitSvc.CreateCommit(message, authorName, authorEmail)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// CreateBranch creates a new branch
func (a *App) CreateBranch(branchName string) error {
	return a.gitSvc.CreateBranch(branchName)
}

// DeleteBranch deletes a branch
func (a *App) DeleteBranch(branchName string) error {
	return a.gitSvc.DeleteBranch(branchName)
}

// CheckoutBranch switches to a different branch
func (a *App) CheckoutBranch(branchName string) error {
	return a.gitSvc.CheckoutBranch(branchName)
}

// CheckoutCommit checks out a specific commit (detached HEAD)
func (a *App) CheckoutCommit(commitHash string) error {
	return a.gitSvc.CheckoutCommit(commitHash)
}

// MergeBranch merges a branch into the current branch
func (a *App) MergeBranch(branchName string) error {
	return a.gitSvc.MergeBranch(branchName)
}

// CreateBranchFromCommit creates a new branch from a specific commit
func (a *App) CreateBranchFromCommit(branchName, commitHash string) error {
	return a.gitSvc.CreateBranchFromCommit(branchName, commitHash)
}

// CherryPickCommit cherry-picks a specific commit onto the current branch
func (a *App) CherryPickCommit(commitHash string) error {
	return a.gitSvc.CherryPickCommit(commitHash)
}

// RevertCommit reverts a specific commit
func (a *App) RevertCommit(commitHash string) error {
	return a.gitSvc.RevertCommit(commitHash)
}

// GetCurrentBranch returns the name of the current branch
func (a *App) GetCurrentBranch() (string, error) {
	return a.gitSvc.GetCurrentBranch()
}

// RenameBranch renames a branch
func (a *App) RenameBranch(oldName, newName string) error {
	return a.gitSvc.RenameBranch(oldName, newName)
}

// GetRemotes returns a list of all remotes
func (a *App) GetRemotes() (string, error) {
	remotes, err := a.gitSvc.GetRemotes()
	if err != nil {
		return "", err
	}

	// Convert to simple model
	var simpleRemotes []models.Remote
	for _, remote := range remotes {
		simpleRemotes = append(simpleRemotes, models.Remote{
			Name: remote.Name,
			URLs: remote.URLs,
		})
	}

	data, err := json.Marshal(simpleRemotes)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Fetch fetches updates from a remote
func (a *App) Fetch(remoteName string) error {
	return a.gitSvc.Fetch(remoteName)
}

// FetchWithAuth fetches updates from a remote with authentication
func (a *App) FetchWithAuth(remoteName, username, password string) error {
	return a.gitSvc.FetchWithAuth(remoteName, username, password)
}

// Pull pulls updates from a remote branch
func (a *App) Pull(remoteName, branchName string) error {
	return a.gitSvc.Pull(remoteName, branchName)
}

// PullWithAuth pulls updates from a remote branch with authentication
func (a *App) PullWithAuth(remoteName, branchName, username, password string) error {
	return a.gitSvc.PullWithAuth(remoteName, branchName, username, password)
}

// Push pushes commits to a remote branch
func (a *App) Push(remoteName, branchName string, force bool) error {
	return a.gitSvc.Push(remoteName, branchName, force)
}

// PushWithAuth pushes commits to a remote branch with authentication
func (a *App) PushWithAuth(remoteName, branchName string, force bool, username, password string) error {
	return a.gitSvc.PushWithAuth(remoteName, branchName, force, username, password)
}

// GetRemoteURL gets the URL for a remote
func (a *App) GetRemoteURL(remoteName string) (string, error) {
	return a.gitSvc.GetRemoteURL(remoteName)
}

// SaveCredentials saves credentials for a remote URL
func (a *App) SaveCredentials(remoteURL, username, password string) error {
	store, err := git.NewCredentialsStore()
	if err != nil {
		return err
	}
	return store.Save(remoteURL, username, password)
}

// GetCredentials gets saved credentials for a remote URL
func (a *App) GetCredentials(remoteURL string) (string, error) {
	store, err := git.NewCredentialsStore()
	if err != nil {
		return "", err
	}

	creds, exists := store.Get(remoteURL)
	if !exists {
		return "", nil
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// DeleteCredentials removes saved credentials for a remote URL
func (a *App) DeleteCredentials(remoteURL string) error {
	store, err := git.NewCredentialsStore()
	if err != nil {
		return err
	}
	return store.Delete(remoteURL)
}

// AddRemote adds a new remote
func (a *App) AddRemote(name, url string) error {
	return a.gitSvc.AddRemote(name, url)
}

// RemoveRemote removes a remote
func (a *App) RemoveRemote(name string) error {
	return a.gitSvc.RemoveRemote(name)
}

// GetAheadBehind returns the number of commits ahead and behind the remote
func (a *App) GetAheadBehind(remoteName, branchName string) (string, error) {
	ahead, behind, err := a.gitSvc.GetAheadBehind(remoteName, branchName)
	if err != nil {
		return "", err
	}

	result := map[string]int{
		"ahead":  ahead,
		"behind": behind,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetStashes returns a list of all stashes
func (a *App) GetStashes() (string, error) {
	stashes, err := a.gitSvc.GetStashes()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(stashes)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// CreateStash creates a new stash
func (a *App) CreateStash(message string, includeUntracked bool) error {
	return a.gitSvc.CreateStash(message, includeUntracked)
}

// ApplyStash applies a stash by index
func (a *App) ApplyStash(index int, pop bool) error {
	return a.gitSvc.ApplyStash(index, pop)
}

// DropStash removes a stash by index
func (a *App) DropStash(index int) error {
	return a.gitSvc.DropStash(index)
}

// ClearStashes removes all stashes
func (a *App) ClearStashes() error {
	return a.gitSvc.ClearStashes()
}

// GetTags returns a list of all tags
func (a *App) GetTags() (string, error) {
	tags, err := a.gitSvc.GetTags()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// CreateTag creates a new tag
func (a *App) CreateTag(tagName, message, commitHash string) error {
	return a.gitSvc.CreateTag(tagName, message, commitHash)
}

// DeleteTag deletes a tag
func (a *App) DeleteTag(tagName string) error {
	return a.gitSvc.DeleteTag(tagName)
}

// GetFileHistory returns the commit history for a specific file
func (a *App) GetFileHistory(filePath string, limit int) (string, error) {
	history, err := a.gitSvc.GetFileHistory(filePath, limit)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(history)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetFileBlame returns blame information for a file
func (a *App) GetFileBlame(filePath string) (string, error) {
	blame, err := a.gitSvc.GetFileBlame(filePath)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(blame)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// MergeBranchWithResult merges a branch and returns detailed result including conflicts
func (a *App) MergeBranchWithResult(branchName string) (string, error) {
	result, err := a.gitSvc.MergeBranchWithResult(branchName)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetConflictFiles returns list of files with merge conflicts
func (a *App) GetConflictFiles() (string, error) {
	files, err := a.gitSvc.GetConflictFiles()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetThreeWayDiff returns 3-way diff for a conflicted file
func (a *App) GetThreeWayDiff(filePath string) (string, error) {
	diff, err := a.gitSvc.GetThreeWayDiff(filePath)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(diff)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ResolveConflictFile saves resolved content and stages the file
func (a *App) ResolveConflictFile(filePath, resolvedContent string) error {
	return a.gitSvc.ResolveConflictFile(filePath, resolvedContent)
}

// CompleteMerge completes the merge after all conflicts are resolved
func (a *App) CompleteMerge(message string) error {
	return a.gitSvc.CompleteMerge(message)
}

// AbortMerge cancels the current merge operation
func (a *App) AbortMerge() error {
	return a.gitSvc.AbortMerge()
}

// IsMergeInProgress checks if a merge is currently in progress
func (a *App) IsMergeInProgress() (bool, error) {
	return a.gitSvc.IsMergeInProgress()
}

// GetMergeSourceBranch returns the name of the branch being merged
func (a *App) GetMergeSourceBranch() (string, error) {
	return a.gitSvc.GetMergeSourceBranch()
}
