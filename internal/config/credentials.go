package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CredentialStore stores provider credentials on disk.
type CredentialStore struct {
	path string
}

// NewCredentialStore returns a credential store rooted at path.
func NewCredentialStore(path string) *CredentialStore {
	return &CredentialStore{path: path}
}

// SetAPIKey stores an API key for provider.
func (s *CredentialStore) SetAPIKey(provider, key string) error {
	if provider == "" {
		return fmt.Errorf("provider is required")
	}
	credentials, err := s.read()
	if err != nil {
		return err
	}
	credentials[provider] = key
	return s.write(credentials)
}

// APIKey returns the stored API key for provider.
func (s *CredentialStore) APIKey(provider string) (string, bool, error) {
	credentials, err := s.read()
	if err != nil {
		return "", false, err
	}
	key, ok := credentials[provider]
	return key, ok, nil
}

func (s *CredentialStore) read() (map[string]string, error) {
	credentials := make(map[string]string)
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return credentials, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return credentials, nil
	}
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	return credentials, nil
}

func (s *CredentialStore) write(credentials map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}
