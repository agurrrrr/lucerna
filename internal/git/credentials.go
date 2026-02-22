package git

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"lucerna/internal/models"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CredentialsStore manages credentials storage
type CredentialsStore struct {
	configPath string
	key        []byte
}

// NewCredentialsStore creates a new credentials store
func NewCredentialsStore() (*CredentialsStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".lucerna")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	key := deriveKey(home)

	return &CredentialsStore{
		configPath: filepath.Join(configDir, "credentials.json"),
		key:        key,
	}, nil
}

// deriveKey creates a machine-specific encryption key
func deriveKey(home string) []byte {
	// Combine multiple machine-specific values to create a stable key
	seed := home + "|" + runtime.GOOS + "|" + runtime.GOARCH
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

// encryptPassword encrypts a password using AES-GCM
func (cs *CredentialsStore) encryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts a password using AES-GCM
func (cs *CredentialsStore) decryptPassword(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode: %w", err)
	}

	block, err := aes.NewCipher(cs.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Fallback: try legacy base64 decoding for migration
		decoded, decErr := base64.StdEncoding.DecodeString(encrypted)
		if decErr == nil && len(decoded) > 0 {
			return string(decoded), nil
		}
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// normalizeURL extracts the host from a URL for matching
func normalizeURL(remoteURL string) string {
	// Parse the URL
	u, err := url.Parse(remoteURL)
	if err != nil {
		// Try to handle git@ style URLs
		if strings.HasPrefix(remoteURL, "git@") {
			// git@github.com:user/repo.git -> github.com/user/repo
			parts := strings.SplitN(remoteURL, ":", 2)
			if len(parts) == 2 {
				host := strings.TrimPrefix(parts[0], "git@")
				path := strings.TrimSuffix(parts[1], ".git")
				return host + "/" + path
			}
		}
		return remoteURL
	}

	// Remove .git suffix
	path := strings.TrimSuffix(u.Path, ".git")
	return u.Host + path
}

// Save saves credentials for a remote URL
func (cs *CredentialsStore) Save(remoteURL, username, password string) error {
	credentials, err := cs.loadAll()
	if err != nil {
		credentials = make(map[string]models.Credentials)
	}

	encryptedPassword, err := cs.encryptPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	normalizedURL := normalizeURL(remoteURL)
	credentials[normalizedURL] = models.Credentials{
		RemoteURL: normalizedURL,
		Username:  username,
		Password:  encryptedPassword,
	}

	return cs.saveAll(credentials)
}

// Get retrieves credentials for a remote URL
func (cs *CredentialsStore) Get(remoteURL string) (*models.Credentials, bool) {
	credentials, err := cs.loadAll()
	if err != nil {
		return nil, false
	}

	normalizedURL := normalizeURL(remoteURL)
	cred, exists := credentials[normalizedURL]
	if !exists {
		return nil, false
	}

	// Decrypt password before returning
	decrypted, err := cs.decryptPassword(cred.Password)
	if err != nil {
		return nil, false
	}
	cred.Password = decrypted
	return &cred, true
}

// Delete removes credentials for a remote URL
func (cs *CredentialsStore) Delete(remoteURL string) error {
	credentials, err := cs.loadAll()
	if err != nil {
		return nil
	}

	normalizedURL := normalizeURL(remoteURL)
	delete(credentials, normalizedURL)

	return cs.saveAll(credentials)
}

// loadAll loads all credentials from file
func (cs *CredentialsStore) loadAll() (map[string]models.Credentials, error) {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]models.Credentials), nil
		}
		return nil, err
	}

	var credentials map[string]models.Credentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return make(map[string]models.Credentials), nil
	}

	return credentials, nil
}

// saveAll saves all credentials to file
func (cs *CredentialsStore) saveAll(credentials map[string]models.Credentials) error {
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	// Use restrictive permissions for credentials file
	return os.WriteFile(cs.configPath, data, 0600)
}
