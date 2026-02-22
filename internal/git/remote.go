package git

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5/config"
)

// AuthError represents an authentication error
type AuthError struct {
	RemoteURL string
	Message   string
}

func (e *AuthError) Error() string {
	return e.Message
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	_, ok := err.(*AuthError)
	return ok
}

// isAuthenticationError checks if the error output indicates auth failure
func isAuthenticationError(output string) bool {
	lowerOutput := strings.ToLower(output)
	authIndicators := []string{
		"authentication failed",
		"could not read username",
		"could not read password",
		"invalid username or password",
		"terminal prompts disabled",
		"unable to access",
		"authorization failed",
		"401",
		"403",
		"permission denied",
		"access denied",
		"support for password authentication was removed",
		"the requested url returned error",
	}
	for _, indicator := range authIndicators {
		if strings.Contains(lowerOutput, indicator) {
			return true
		}
	}
	return false
}

// needsAuthForHTTPS checks if a failed HTTPS operation likely needs auth
func needsAuthForHTTPS(remoteURL, output, errMsg string) bool {
	if !strings.HasPrefix(remoteURL, "https://") {
		return false
	}

	// If output is empty and there's an exit status error, likely auth issue
	output = strings.TrimSpace(output)
	if output == "" && strings.Contains(errMsg, "exit status") {
		return true
	}

	// Check for common auth-related patterns
	return isAuthenticationError(output) || isAuthenticationError(errMsg)
}

// buildAuthURL builds a URL with embedded credentials
func buildAuthURL(remoteURL, username, password string) (string, error) {
	// Skip if not HTTPS URL
	if !strings.HasPrefix(remoteURL, "https://") {
		return remoteURL, nil
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return remoteURL, err
	}

	u.User = url.UserPassword(username, password)
	return u.String(), nil
}

// GetRemotes returns a list of all remotes
func (s *RepositoryService) GetRemotes() ([]*config.RemoteConfig, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository opened")
	}

	remotes, err := s.repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	var remoteConfigs []*config.RemoteConfig
	for _, remote := range remotes {
		remoteConfigs = append(remoteConfigs, remote.Config())
	}

	return remoteConfigs, nil
}

// GetRemoteURL gets the URL for a remote
func (s *RepositoryService) GetRemoteURL(remoteName string) (string, error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository opened")
	}

	remote, err := s.repo.Remote(remoteName)
	if err != nil {
		return "", err
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("remote has no URLs")
	}

	return urls[0], nil
}

// Fetch fetches updates from a remote using git command
func (s *RepositoryService) Fetch(remoteName string) error {
	return s.FetchWithAuth(remoteName, "", "")
}

// FetchWithAuth fetches updates from a remote with optional authentication
func (s *RepositoryService) FetchWithAuth(remoteName, username, password string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	// Get remote URL for auth error reporting
	remoteURL, _ := s.GetRemoteURL(remoteName)

	var cmd *exec.Cmd
	if username != "" && password != "" && strings.HasPrefix(remoteURL, "https://") {
		// Use credential helper approach
		authURL, err := buildAuthURL(remoteURL, username, password)
		if err != nil {
			return err
		}
		cmd = exec.Command("git", "fetch", authURL)
	} else {
		cmd = exec.Command("git", "fetch", remoteName)
	}

	cmd.Dir = s.path
	cmd.Env = append(cmd.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		errStr := err.Error()

		// Check if auth is needed
		if needsAuthForHTTPS(remoteURL, outputStr, errStr) {
			return &AuthError{
				RemoteURL: remoteURL,
				Message:   fmt.Sprintf("AUTH_REQUIRED:%s", remoteURL),
			}
		}
		return fmt.Errorf("failed to fetch from '%s': %s", remoteName, outputStr)
	}

	// Refresh the go-git repository to pick up any changes
	s.RefreshRepository()

	return nil
}

// Pull pulls updates from a remote branch using git command
func (s *RepositoryService) Pull(remoteName, branchName string) error {
	return s.PullWithAuth(remoteName, branchName, "", "")
}

// PullWithAuth pulls updates from a remote branch with optional authentication
func (s *RepositoryService) PullWithAuth(remoteName, branchName, username, password string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	// Get remote URL for auth error reporting
	remoteURL, _ := s.GetRemoteURL(remoteName)

	var cmd *exec.Cmd
	if username != "" && password != "" && strings.HasPrefix(remoteURL, "https://") {
		authURL, err := buildAuthURL(remoteURL, username, password)
		if err != nil {
			return err
		}
		args := []string{"pull", authURL}
		if branchName != "" {
			args = append(args, branchName)
		}
		cmd = exec.Command("git", args...)
	} else {
		args := []string{"pull", remoteName}
		if branchName != "" {
			args = append(args, branchName)
		}
		cmd = exec.Command("git", args...)
	}

	cmd.Dir = s.path
	cmd.Env = append(cmd.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		errStr := err.Error()

		if strings.Contains(outputStr, "Already up to date") {
			return nil
		}

		// Check if auth is needed
		if needsAuthForHTTPS(remoteURL, outputStr, errStr) {
			return &AuthError{
				RemoteURL: remoteURL,
				Message:   fmt.Sprintf("AUTH_REQUIRED:%s", remoteURL),
			}
		}
		return fmt.Errorf("failed to pull from '%s': %s", remoteName, outputStr)
	}

	// Refresh the go-git repository to pick up any changes
	s.RefreshRepository()

	return nil
}

// Push pushes commits to a remote branch using git command
func (s *RepositoryService) Push(remoteName, branchName string, force bool) error {
	return s.PushWithAuth(remoteName, branchName, force, "", "")
}

// PushWithAuth pushes commits to a remote branch with optional authentication
func (s *RepositoryService) PushWithAuth(remoteName, branchName string, force bool, username, password string) error {
	if s.path == "" {
		return fmt.Errorf("no repository opened")
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	// Get current branch if branchName is empty
	if branchName == "" {
		head, err := s.repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}
		branchName = head.Name().Short()
	}

	// Get remote URL for auth error reporting
	remoteURL, _ := s.GetRemoteURL(remoteName)

	var cmd *exec.Cmd
	if username != "" && password != "" && strings.HasPrefix(remoteURL, "https://") {
		authURL, err := buildAuthURL(remoteURL, username, password)
		if err != nil {
			return err
		}
		args := []string{"push"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, authURL, branchName)
		cmd = exec.Command("git", args...)
	} else {
		args := []string{"push"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, remoteName, branchName)
		cmd = exec.Command("git", args...)
	}

	cmd.Dir = s.path
	cmd.Env = append(cmd.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		errStr := err.Error()

		if strings.Contains(outputStr, "Everything up-to-date") {
			return nil
		}

		// Check if auth is needed
		if needsAuthForHTTPS(remoteURL, outputStr, errStr) {
			return &AuthError{
				RemoteURL: remoteURL,
				Message:   fmt.Sprintf("AUTH_REQUIRED:%s", remoteURL),
			}
		}
		return fmt.Errorf("failed to push to '%s/%s': %s", remoteName, branchName, outputStr)
	}

	// Update remote tracking ref after successful push
	s.updateRemoteTrackingRef(remoteName, branchName)

	// Refresh the go-git repository to pick up any changes
	s.RefreshRepository()

	return nil
}

// updateRemoteTrackingRef updates the remote tracking ref after a successful push
func (s *RepositoryService) updateRemoteTrackingRef(remoteName, branchName string) {
	// Get the current HEAD hash
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = s.path
	headOutput, err := headCmd.Output()
	if err != nil {
		return
	}
	headHash := strings.TrimSpace(string(headOutput))

	// Update the remote tracking ref to point to the pushed commit
	refName := fmt.Sprintf("refs/remotes/%s/%s", remoteName, branchName)
	cmd := exec.Command("git", "update-ref", refName, headHash)
	cmd.Dir = s.path
	cmd.Run() // Ignore errors
}

// AddRemote adds a new remote
func (s *RepositoryService) AddRemote(name, url string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if name == "" {
		return fmt.Errorf("remote name cannot be empty")
	}

	if url == "" {
		return fmt.Errorf("remote URL cannot be empty")
	}

	_, err := s.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})

	if err != nil {
		return fmt.Errorf("failed to add remote '%s': %w", name, err)
	}

	return nil
}

// RemoveRemote removes a remote
func (s *RepositoryService) RemoveRemote(name string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository opened")
	}

	if name == "" {
		return fmt.Errorf("remote name cannot be empty")
	}

	err := s.repo.DeleteRemote(name)
	if err != nil {
		return fmt.Errorf("failed to remove remote '%s': %w", name, err)
	}

	return nil
}

// GetAheadBehind returns the number of commits ahead and behind the remote
func (s *RepositoryService) GetAheadBehind(remoteName, branchName string) (ahead, behind int, err error) {
	if s.path == "" {
		return 0, 0, fmt.Errorf("no repository opened")
	}

	if remoteName == "" {
		remoteName = "origin"
	}

	if branchName == "" {
		// Get current branch
		head, err := s.repo.Head()
		if err != nil {
			return 0, 0, err
		}
		branchName = head.Name().Short()
	}

	// Use git rev-list to count ahead/behind
	// Format: git rev-list --left-right --count origin/main...HEAD
	upstream := fmt.Sprintf("%s/%s", remoteName, branchName)
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", upstream+"...HEAD")
	cmd.Dir = s.path

	output, err := cmd.Output()
	if err != nil {
		// If the upstream branch doesn't exist, return 0,0
		return 0, 0, nil
	}

	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)
	if len(parts) != 2 {
		return 0, 0, nil
	}

	fmt.Sscanf(parts[0], "%d", &behind)
	fmt.Sscanf(parts[1], "%d", &ahead)

	return ahead, behind, nil
}
