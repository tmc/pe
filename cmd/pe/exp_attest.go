package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const unsignedManifestType = "pe.unsigned_file_manifest.v1"

var expAttestCmd = newExpAttestCmd()

type unsignedFileManifest struct {
	Type      string                  `json:"type"`
	Algorithm string                  `json:"algorithm"`
	Entries   []unsignedManifestEntry `json:"entries"`
}

type unsignedManifestEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

func newExpAttestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest",
		Short: "Create and verify unsigned local hash manifests",
		Long: `Create and verify deterministic SHA-256 manifests for local files.

These manifests are unsigned and local-only. They detect file content changes,
missing files, and manifest tampering, but they do not prove identity, origin,
or freshness.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Command 'attest' is an experimental prototype.")
			cmd.Println("Use 'manifest' to create an unsigned hash manifest or 'verify' to check one.")
		},
	}

	var root string
	manifestCmd := &cobra.Command{
		Use:   "manifest [file-or-dir...]",
		Short: "Write an unsigned deterministic file manifest to stdout",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifest, err := buildUnsignedManifest(root, args)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(manifest)
		},
	}
	manifestCmd.Flags().StringVar(&root, "root", ".", "root directory for relative paths")

	var verifyRoot string
	verifyCmd := &cobra.Command{
		Use:   "verify <manifest.json>",
		Short: "Verify an unsigned file manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := verifyUnsignedManifestFile(verifyRoot, args[0]); err != nil {
				return err
			}
			cmd.Println("unsigned manifest verified")
			return nil
		},
	}
	verifyCmd.Flags().StringVar(&verifyRoot, "root", ".", "root directory for relative paths")

	cmd.AddCommand(manifestCmd, verifyCmd)
	return cmd
}

func buildUnsignedManifest(root string, paths []string) (*unsignedFileManifest, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	seen := make(map[string]unsignedManifestEntry)
	for _, path := range paths {
		full, rel, err := resolveManifestPath(rootAbs, path)
		if err != nil {
			return nil, err
		}
		if err := collectManifestEntries(rootAbs, full, rel, seen); err != nil {
			return nil, err
		}
	}

	manifest := &unsignedFileManifest{
		Type:      unsignedManifestType,
		Algorithm: "sha256",
		Entries:   make([]unsignedManifestEntry, 0, len(seen)),
	}
	for _, entry := range seen {
		manifest.Entries = append(manifest.Entries, entry)
	}
	sort.Slice(manifest.Entries, func(i, j int) bool {
		return manifest.Entries[i].Path < manifest.Entries[j].Path
	})
	return manifest, nil
}

func verifyUnsignedManifestFile(root, manifestPath string) error {
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- manifest path is the explicit file being verified.
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	var manifest unsignedFileManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}
	if manifest.Type != unsignedManifestType {
		return fmt.Errorf("unsupported manifest type %q", manifest.Type)
	}
	if manifest.Algorithm != "sha256" {
		return fmt.Errorf("unsupported manifest algorithm %q", manifest.Algorithm)
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	for _, entry := range manifest.Entries {
		full, rel, err := resolveManifestPath(rootAbs, filepath.FromSlash(entry.Path))
		if err != nil {
			return err
		}
		if filepath.ToSlash(rel) != entry.Path {
			return fmt.Errorf("manifest path is not canonical: %s", entry.Path)
		}
		got, err := hashManifestFile(full, rel)
		if err != nil {
			return err
		}
		if got.Size != entry.Size {
			return fmt.Errorf("size mismatch for %s", entry.Path)
		}
		if got.SHA256 != entry.SHA256 {
			return fmt.Errorf("sha256 mismatch for %s", entry.Path)
		}
	}
	return nil
}

func collectManifestEntries(rootAbs, full, rel string, entries map[string]unsignedManifestEntry) error {
	info, err := os.Lstat(full)
	if err != nil {
		return fmt.Errorf("stat %s: %w", rel, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not supported in unsigned manifests: %s", rel)
	}
	if !info.IsDir() {
		entry, err := hashManifestFile(full, rel)
		if err != nil {
			return err
		}
		entries[entry.Path] = entry
		return nil
	}

	return filepath.WalkDir(full, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			walkRel, relErr := filepath.Rel(rootAbs, path)
			if relErr != nil {
				return relErr
			}
			return fmt.Errorf("symlinks are not supported in unsigned manifests: %s", walkRel)
		}
		if d.IsDir() {
			return nil
		}
		walkRel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		entry, err := hashManifestFile(path, walkRel)
		if err != nil {
			return err
		}
		entries[entry.Path] = entry
		return nil
	})
}

func resolveManifestPath(rootAbs, path string) (string, string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", "", fmt.Errorf("manifest paths must be relative: %q", path)
	}
	full, err := filepath.Abs(filepath.Join(rootAbs, path))
	if err != nil {
		return "", "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	rel, err := filepath.Rel(rootAbs, full)
	if err != nil {
		return "", "", fmt.Errorf("resolve relative path %q: %w", path, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("path escapes manifest root: %s", path)
	}
	return full, rel, nil
}

func hashManifestFile(path, rel string) (unsignedManifestEntry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return unsignedManifestEntry{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return unsignedManifestEntry{}, fmt.Errorf("symlinks are not supported in unsigned manifests: %s", rel)
	}
	file, err := os.Open(path) // #nosec G304 -- caller resolves path under the manifest root.
	if err != nil {
		return unsignedManifestEntry{}, fmt.Errorf("open %s: %w", rel, err)
	}
	defer file.Close()

	h := sha256.New()
	n, err := io.Copy(h, file)
	if err != nil {
		return unsignedManifestEntry{}, fmt.Errorf("hash %s: %w", rel, err)
	}
	return unsignedManifestEntry{
		Path:   filepath.ToSlash(rel),
		Size:   n,
		SHA256: hex.EncodeToString(h.Sum(nil)),
	}, nil
}
