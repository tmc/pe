package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// peInitCmd returns the init command for PE repository initialization
func peInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a PE repository",
		Long: `Initialize a PE repository with a .pe directory structure.

Creates the necessary directories and configuration files for a PE project,
similar to 'git init' or 'go mod init'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if .pe already exists
			if _, err := os.Stat(".pe"); err == nil && !force {
				return fmt.Errorf("already initialized")
			}

			// Create .pe directory structure
			dirs := []string{
				".pe",
				".pe/cache",
				".pe/modules",
				".pe/config",
				".pe/attestations",
				".pe/keys",
			}

			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}

			// Create default config
			configContent := `# PE Configuration
version: 1.0

# Default provider settings
providers:
  default: openai:gpt-4

# Cache settings
cache:
  enabled: true
  ttl: 24h
  max_size: 1GB

# Security settings
security:
  verify_signatures: true
  require_attestations: false
`
			configPath := filepath.Join(".pe", "config", "pe.yaml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}

			// Create .peignore file
			ignoreContent := `# PE ignore file
*.tmp
*.log
.env
secrets/
`
			if err := os.WriteFile(".peignore", []byte(ignoreContent), 0644); err != nil {
				return fmt.Errorf("failed to create .peignore file: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized PE repository\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - Create your first prompt file: echo 'Explain {{topic}} simply' > explain.txt\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - Run it: pe run explain.txt --var topic='quantum computing'\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  - Add evaluation: pe eval-prompt explain.txt\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force reinitialization even if .pe exists")

	return cmd
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}