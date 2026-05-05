package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var expCacheCmd = newExpCacheCmd()

func newExpCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Inspect a local content-addressed cache",
		Long: `Inspect a local content-addressed cache for files and unsigned manifests.

Cache keys are SHA-256 digests. The cache is local only and unsigned; it does
not prove identity, origin, or freshness.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Command 'cache' is an experimental prototype.")
			cmd.Println("Use 'key', 'put', 'get', or 'verify' for local content-addressed cache work.")
		},
	}

	var keyManifest bool
	keyCmd := &cobra.Command{
		Use:   "key <file>",
		Short: "Print the deterministic SHA-256 key for a file or manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string
			var err error
			if keyManifest {
				key, err = unsignedManifestDigestFile(args[0])
			} else {
				key, err = contentKeyFile(args[0])
			}
			if err != nil {
				return err
			}
			cmd.Println(key)
			return nil
		},
	}
	keyCmd.Flags().BoolVar(&keyManifest, "manifest", false, "hash canonical unsigned manifest JSON instead of raw file content")

	var putCacheDir string
	putCmd := &cobra.Command{
		Use:   "put <file>",
		Short: "Store a file in the local content-addressed cache",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := putCacheFile(putCacheDir, args[0])
			if err != nil {
				return err
			}
			cmd.Println(key)
			return nil
		},
	}
	putCmd.Flags().StringVar(&putCacheDir, "cache-dir", ".pe/cache", "cache directory")

	var getCacheDir string
	getCmd := &cobra.Command{
		Use:   "get <sha256>",
		Short: "Write cached content to stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := getCacheBytes(getCacheDir, args[0])
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
	getCmd.Flags().StringVar(&getCacheDir, "cache-dir", ".pe/cache", "cache directory")

	var verifyCacheDir string
	verifyCmd := &cobra.Command{
		Use:   "verify <sha256>",
		Short: "Verify a cached entry against its SHA-256 key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := verifyCacheKey(verifyCacheDir, args[0]); err != nil {
				return err
			}
			cmd.Println("cache entry verified")
			return nil
		},
	}
	verifyCmd.Flags().StringVar(&verifyCacheDir, "cache-dir", ".pe/cache", "cache directory")

	cmd.AddCommand(keyCmd, putCmd, getCmd, verifyCmd)
	return cmd
}

func contentKeyFile(path string) (string, error) {
	data, err := readRegularFile(path)
	if err != nil {
		return "", err
	}
	return contentKey(data), nil
}

func contentKey(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func unsignedManifestDigestFile(path string) (string, error) {
	data, err := readRegularFile(path)
	if err != nil {
		return "", err
	}
	var manifest unsignedFileManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("parse manifest: %w", err)
	}
	if manifest.Type != unsignedManifestType {
		return "", fmt.Errorf("unsupported manifest type %q", manifest.Type)
	}
	if manifest.Algorithm != "sha256" {
		return "", fmt.Errorf("unsupported manifest algorithm %q", manifest.Algorithm)
	}
	canonical, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("canonicalize manifest: %w", err)
	}
	return contentKey(canonical), nil
}

func putCacheFile(cacheDir, path string) (string, error) {
	data, err := readRegularFile(path)
	if err != nil {
		return "", err
	}
	key := contentKey(data)
	objectPath, err := cacheObjectPath(cacheDir, key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		return "", fmt.Errorf("create cache directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(objectPath), ".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create cache temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", fmt.Errorf("write cache temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("close cache temp file: %w", err)
	}
	if err := os.Rename(tmpName, objectPath); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("store cache object: %w", err)
	}
	return key, nil
}

func getCacheBytes(cacheDir, key string) ([]byte, error) {
	objectPath, err := cacheObjectPath(cacheDir, key)
	if err != nil {
		return nil, err
	}
	if err := rejectSymlinkFile(objectPath); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(objectPath) // #nosec G304 -- objectPath is derived from a validated SHA-256 key under cacheDir.
	if err != nil {
		return nil, fmt.Errorf("read cache object: %w", err)
	}
	return data, nil
}

func verifyCacheKey(cacheDir, key string) error {
	data, err := getCacheBytes(cacheDir, key)
	if err != nil {
		return err
	}
	if got := contentKey(data); got != strings.ToLower(key) {
		return fmt.Errorf("cache object hash mismatch: got %s", got)
	}
	return nil
}

func cacheObjectPath(cacheDir, key string) (string, error) {
	key = strings.ToLower(key)
	if !isSHA256Hex(key) {
		return "", fmt.Errorf("invalid sha256 key: %q", key)
	}
	cacheAbs, err := filepath.Abs(cacheDir)
	if err != nil {
		return "", fmt.Errorf("resolve cache dir: %w", err)
	}
	base := filepath.Join(cacheAbs, "objects", "sha256")
	objectPath := filepath.Join(base, key[:2], key[2:])
	rel, err := filepath.Rel(cacheAbs, objectPath)
	if err != nil {
		return "", fmt.Errorf("resolve cache object: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("cache object escapes cache directory")
	}
	return objectPath, nil
}

func readRegularFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	if err := rejectSymlinkFile(path); err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file: %s", path)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- caller explicitly selects the file and symlink components are rejected.
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

func rejectSymlinkFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not supported in cache paths: %s", path)
	}
	return nil
}

func isSHA256Hex(key string) bool {
	if len(key) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(key)
	return err == nil
}
