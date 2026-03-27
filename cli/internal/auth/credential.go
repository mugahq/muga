package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	credentialDir  = "muga"
	credentialFile = "credentials.json"
)

// Credential holds the stored authentication tokens.
type Credential struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	UserName     string    `json:"user_name,omitempty"`
	UserEmail    string    `json:"user_email,omitempty"`
}

// CredentialStore reads and writes credentials to the filesystem.
type CredentialStore struct {
	dir string
}

// NewCredentialStore creates a store that uses the default credential path.
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{dir: credentialPath()}
}

// NewCredentialStoreWithDir creates a store that uses a custom directory.
func NewCredentialStoreWithDir(dir string) *CredentialStore {
	return &CredentialStore{dir: dir}
}

// Save persists a credential to disk.
func (s *CredentialStore) Save(cred *Credential) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("creating credential directory: %w", err)
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding credentials: %w", err)
	}

	path := filepath.Join(s.dir, credentialFile)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}

	return nil
}

// Load reads a stored credential from disk. Returns nil if no credential exists.
func (s *CredentialStore) Load() (*Credential, error) {
	path := filepath.Join(s.dir, credentialFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading credentials: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("decoding credentials: %w", err)
	}

	return &cred, nil
}

// Delete removes stored credentials.
func (s *CredentialStore) Delete() error {
	path := filepath.Join(s.dir, credentialFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	return nil
}

func credentialPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, credentialDir)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", credentialDir)
}
