package main

import (
	"os"
	"path/filepath"

	"github.com/tmc/pe/internal/promptfoo/storage"
)

// defaultRunStoreDir returns the directory where evaluation runs are persisted.
// It keeps the historical ~/.promptfoo/evals location so runs saved by earlier
// versions still resolve. If the home directory cannot be determined, it falls
// back to a relative .promptfoo/evals.
func defaultRunStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".promptfoo", "evals")
	}
	return filepath.Join(home, ".promptfoo", "evals")
}

// defaultRunStore returns the run store backed by the default directory.
func defaultRunStore() *storage.RunStore {
	return storage.NewRunStore(storage.NewDirStore(defaultRunStoreDir()))
}
