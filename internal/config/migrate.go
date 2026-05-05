package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// MigrateFile reads src with the normal config loader and writes canonical PE YAML to dst.
func MigrateFile(src, dst string) error {
	if src == "" {
		return fmt.Errorf("source config path is required")
	}
	if dst == "" {
		return fmt.Errorf("destination config path is required")
	}
	manager, err := NewManager(WithConfigPaths(src))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := manager.SaveToFile(dst); err != nil {
		return err
	}
	return nil
}
