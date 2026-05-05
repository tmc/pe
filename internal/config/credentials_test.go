package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialStoreAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	store := NewCredentialStore(path)

	if err := store.SetAPIKey("openai", "sk-test"); err != nil {
		t.Fatalf("SetAPIKey: %v", err)
	}
	key, ok, err := store.APIKey("openai")
	if err != nil {
		t.Fatalf("APIKey: %v", err)
	}
	if !ok || key != "sk-test" {
		t.Fatalf("APIKey = %q, %v; want sk-test, true", key, ok)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("credentials mode = %v, want 0600", got)
	}
}

func TestCredentialStoreMissingAPIKey(t *testing.T) {
	store := NewCredentialStore(filepath.Join(t.TempDir(), "credentials.json"))
	_, ok, err := store.APIKey("anthropic")
	if err != nil {
		t.Fatalf("APIKey: %v", err)
	}
	if ok {
		t.Fatal("missing key reported present")
	}
}

func TestCredentialStoreRequiresProvider(t *testing.T) {
	store := NewCredentialStore(filepath.Join(t.TempDir(), "credentials.json"))
	if err := store.SetAPIKey("", "key"); err == nil {
		t.Fatal("SetAPIKey with empty provider succeeded")
	}
}
