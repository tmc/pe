package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo/execution/distributed"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage verifiable shared caching",
	Long: `Cache management for distributed prompt engineering.

Provides cryptographically signed, content-addressed caching with
witness verification for secure cache sharing across nodes.`,
}

func init() {
	cacheCmd.AddCommand(cacheStatusCmd())
	cacheCmd.AddCommand(cacheListCmd())
	cacheCmd.AddCommand(cacheInspectCmd())
	cacheCmd.AddCommand(cacheImportCmd())
	cacheCmd.AddCommand(cacheExportCmd())
	cacheCmd.AddCommand(cacheVerifyCmd())
	cacheCmd.AddCommand(cacheClearCmd())
	cacheCmd.AddCommand(cacheStatsCmd())
}

func cacheStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show cache status and statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			stats, err := cache.GetStats()
			if err != nil {
				return fmt.Errorf("failed to get cache stats: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cache Status:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Location: %s\n", ".pe/cache")
			fmt.Fprintf(cmd.OutOrStdout(), "  Entries: %d\n", stats.TotalEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Size: %s\n", formatBytes(stats.TotalSize))
			if stats.TotalEntries == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Hit rate: N/A\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  Hit Rate: %.2f%%\n", stats.HitRate*100)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Witnesses: %d\n", stats.WitnessCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Verified: %d\n", stats.VerifiedEntries)

			return nil
		},
	}
}

func cacheListCmd() *cobra.Command {
	var format string
	var filter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cache entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			entries, err := cache.List(filter)
			if err != nil {
				return fmt.Errorf("failed to list cache entries: %w", err)
			}

			switch format {
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(entries)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Cache entries:\n")
				for _, entry := range entries {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%d witnesses\n",
						entry.Hash[:12],
						entry.Type,
						entry.CreatedAt.Format("2006-01-02 15:04:05"),
						len(entry.Witnesses))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json)")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter entries by type or pattern")

	return cmd
}

func cacheInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <hash>",
		Short: "View detailed cache entry information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			entry, err := cache.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to get cache entry: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cache Entry: %s\n", entry.Hash)
			fmt.Fprintf(cmd.OutOrStdout(), "Type: %s\n", entry.Type)
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", entry.CreatedAt.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "TTL: %s\n", entry.TTL)
			fmt.Fprintf(cmd.OutOrStdout(), "Size: %s\n", formatBytes(int64(len(entry.Data))))
			fmt.Fprintf(cmd.OutOrStdout(), "Signature: %x\n", entry.Signature)
			fmt.Fprintf(cmd.OutOrStdout(), "Witnesses: %d\n", len(entry.Witnesses))
			for i, w := range entry.Witnesses {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%d] %s at %s\n", i+1, w.NodeID, w.Timestamp.Format(time.RFC3339))
			}

			return nil
		},
	}
}

func cacheImportCmd() *cobra.Command {
	var verify bool

	cmd := &cobra.Command{
		Use:   "import <bundle.tar.gz>",
		Short: "Import a cache bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read bundle: %w", err)
			}

			imported, failed, err := cache.ImportBundle(data, verify)
			if err != nil {
				return fmt.Errorf("failed to import bundle: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d entries\n", imported)
			if failed > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Failed to import %d entries\n", failed)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&verify, "verify", true, "Verify signatures before importing")

	return cmd
}

func cacheExportCmd() *cobra.Command {
	var filter string
	var sign bool

	cmd := &cobra.Command{
		Use:   "export <output.tar.gz>",
		Short: "Export cache entries to a bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			bundle, err := cache.ExportBundle(filter, sign)
			if err != nil {
				return fmt.Errorf("failed to export bundle: %w", err)
			}

			if err := os.WriteFile(args[0], bundle, 0644); err != nil {
				return fmt.Errorf("failed to write bundle: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Exported cache bundle to %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter entries to export")
	cmd.Flags().BoolVar(&sign, "sign", true, "Sign the bundle")

	return cmd
}

func cacheVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify cache integrity",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			results, err := cache.VerifyIntegrity()
			if err != nil {
				return fmt.Errorf("failed to verify cache: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cache Verification Results:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Total Entries: %d\n", results.TotalEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Valid: %d\n", results.ValidEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Invalid: %d\n", results.InvalidEntries)
			fmt.Fprintf(cmd.OutOrStdout(), "  Corrupted: %d\n", results.CorruptedEntries)

			if results.InvalidEntries > 0 || results.CorruptedEntries > 0 {
				return fmt.Errorf("cache verification failed")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nCache is valid.\n")
			return nil
		},
	}
}

func cacheClearCmd() *cobra.Command {
	var force bool
	var pattern string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cache entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Fprintf(cmd.OutOrStderr(), "This will clear cache entries. Use --force to confirm.\n")
				return nil
			}

			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			cleared, err := cache.Clear(pattern)
			if err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cache cleared\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Cleared %d cache entries\n", cleared)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force clear without confirmation")
	cmd.Flags().StringVar(&pattern, "pattern", "", "Clear only entries matching pattern")

	return cmd
}

func cacheStatsCmd() *cobra.Command {
	var detailed bool
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show detailed cache statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cache, err := distributed.NewDistributedCache(".pe/cache")
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			stats, err := cache.GetDetailedStats()
			if err != nil {
				return fmt.Errorf("failed to get cache stats: %w", err)
			}

			switch format {
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(stats)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Cache statistics:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Total entries: %d\n", stats.TotalEntries)
				fmt.Fprintf(cmd.OutOrStdout(), "  Total Size: %s\n", formatBytes(stats.TotalSize))
				fmt.Fprintf(cmd.OutOrStdout(), "  Hit Rate: %.2f%%\n", stats.HitRate*100)
				fmt.Fprintf(cmd.OutOrStdout(), "  Miss Rate: %.2f%%\n", stats.MissRate*100)
				fmt.Fprintf(cmd.OutOrStdout(), "  Avg Entry Size: %s\n", formatBytes(stats.AvgEntrySize))
				fmt.Fprintf(cmd.OutOrStdout(), "  Oldest Entry: %s\n", stats.OldestEntry.Format(time.RFC3339))
				fmt.Fprintf(cmd.OutOrStdout(), "  Newest Entry: %s\n", stats.NewestEntry.Format(time.RFC3339))

				if detailed {
					fmt.Fprintf(cmd.OutOrStdout(), "\nEntry Types:\n")
					for entryType, count := range stats.EntryTypes {
						fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", entryType, count)
					}

					fmt.Fprintf(cmd.OutOrStdout(), "\nWitness Distribution:\n")
					for witnesses, count := range stats.WitnessDistribution {
						fmt.Fprintf(cmd.OutOrStdout(), "  %d witnesses: %d entries\n", witnesses, count)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed statistics")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json)")

	return cmd
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes == 0 {
		return "0"
	}
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
