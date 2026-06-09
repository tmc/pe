package main

import (
	"crypto/ed25519"
	"crypto/rand"
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
const signedManifestType = "pe.signed_file_manifest.v1"

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

type signedManifestEnvelope struct {
	Type        string               `json:"type"`
	Algorithm   string               `json:"algorithm"`
	PublicKey   string               `json:"public_key"`
	Fingerprint string               `json:"fingerprint"`
	Signature   string               `json:"signature"`
	Payload     unsignedFileManifest `json:"payload"`
}

type attestKeyPair struct {
	Algorithm   string `json:"algorithm"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	Fingerprint string `json:"fingerprint"`
}

func newExpAttestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest",
		Short: "Create, sign, and verify local hash manifests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Command 'attest' is an experimental prototype.")
			cmd.Println("Use 'manifest' to create an unsigned hash manifest, 'verify' to check one, or 'sign' for an Ed25519 envelope.")
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

	keygenCmd := &cobra.Command{
		Use:   "keygen",
		Short: "Write an Ed25519 attestation key pair to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := generateAttestKeyPair()
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(key)
		},
	}

	var signKey string
	signCmd := &cobra.Command{
		Use:   "sign <manifest.json>",
		Short: "Sign an unsigned manifest with an Ed25519 private key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envelope, err := signUnsignedManifestFile(args[0], signKey)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(envelope)
		},
	}
	signCmd.Flags().StringVar(&signKey, "private-key", "", "hex Ed25519 private key from keygen")

	var signedRoot string
	var signedPublicKey string
	verifySignedCmd := &cobra.Command{
		Use:   "verify-signed <signed-manifest.json>",
		Short: "Verify a signed manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := verifySignedManifestFile(signedRoot, args[0], signedPublicKey); err != nil {
				return err
			}
			cmd.Println("signed manifest verified")
			return nil
		},
	}
	verifySignedCmd.Flags().StringVar(&signedRoot, "root", ".", "root directory for relative paths")
	verifySignedCmd.Flags().StringVar(&signedPublicKey, "public-key", "", "optional expected hex Ed25519 public key")

	cmd.AddCommand(manifestCmd, verifyCmd, keygenCmd, signCmd, verifySignedCmd)
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
	manifest, err := readUnsignedManifestFile(manifestPath)
	if err != nil {
		return err
	}
	return verifyUnsignedManifest(root, manifest)
}

func generateAttestKeyPair() (*attestKeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return &attestKeyPair{
		Algorithm:   "ed25519",
		PublicKey:   hex.EncodeToString(publicKey),
		PrivateKey:  hex.EncodeToString(privateKey),
		Fingerprint: publicKeyFingerprint(publicKey),
	}, nil
}

func signUnsignedManifestFile(manifestPath, privateKeyHex string) (*signedManifestEnvelope, error) {
	if privateKeyHex == "" {
		return nil, fmt.Errorf("private key is required")
	}
	privateKey, err := parseEd25519PrivateKey(privateKeyHex)
	if err != nil {
		return nil, err
	}
	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid Ed25519 private key")
	}
	manifest, err := readUnsignedManifestFile(manifestPath)
	if err != nil {
		return nil, err
	}
	payload, err := canonicalManifestPayload(*manifest)
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(privateKey, signedManifestMessage(payload))
	return &signedManifestEnvelope{
		Type:        signedManifestType,
		Algorithm:   "ed25519",
		PublicKey:   hex.EncodeToString(publicKey),
		Fingerprint: publicKeyFingerprint(publicKey),
		Signature:   hex.EncodeToString(signature),
		Payload:     *manifest,
	}, nil
}

func verifySignedManifestFile(root, envelopePath, expectedPublicKeyHex string) error {
	data, err := os.ReadFile(envelopePath) // #nosec G304 -- envelope path is the explicit file being verified.
	if err != nil {
		return fmt.Errorf("read signed manifest: %w", err)
	}
	var envelope signedManifestEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("parse signed manifest: %w", err)
	}
	if err := verifySignedManifestEnvelope(&envelope, expectedPublicKeyHex); err != nil {
		return err
	}
	return verifyUnsignedManifest(root, &envelope.Payload)
}

func verifySignedManifestEnvelope(envelope *signedManifestEnvelope, expectedPublicKeyHex string) error {
	if envelope == nil {
		return fmt.Errorf("signed manifest is nil")
	}
	if envelope.Type != signedManifestType {
		return fmt.Errorf("unsupported signed manifest type %q", envelope.Type)
	}
	if envelope.Algorithm != "ed25519" {
		return fmt.Errorf("unsupported signed manifest algorithm %q", envelope.Algorithm)
	}
	if envelope.Signature == "" {
		return fmt.Errorf("missing signature")
	}
	publicKey, err := parseEd25519PublicKey(envelope.PublicKey)
	if err != nil {
		return err
	}
	if expectedPublicKeyHex != "" && !strings.EqualFold(expectedPublicKeyHex, envelope.PublicKey) {
		return fmt.Errorf("signed manifest public key does not match expected key")
	}
	fp := publicKeyFingerprint(publicKey)
	if envelope.Fingerprint != "" && envelope.Fingerprint != fp {
		return fmt.Errorf("signed manifest fingerprint does not match public key")
	}
	signature, err := hex.DecodeString(envelope.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	payload, err := canonicalManifestPayload(envelope.Payload)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, signedManifestMessage(payload), signature) {
		return fmt.Errorf("invalid signed manifest signature")
	}
	return nil
}

func readUnsignedManifestFile(manifestPath string) (*unsignedFileManifest, error) {
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- manifest path is the explicit file being signed.
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var manifest unsignedFileManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := validateUnsignedManifest(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func validateUnsignedManifest(manifest *unsignedFileManifest) error {
	if manifest.Type != unsignedManifestType {
		return fmt.Errorf("unsupported manifest type %q", manifest.Type)
	}
	if manifest.Algorithm != "sha256" {
		return fmt.Errorf("unsupported manifest algorithm %q", manifest.Algorithm)
	}
	return nil
}

func verifyUnsignedManifest(root string, manifest *unsignedFileManifest) error {
	if err := validateUnsignedManifest(manifest); err != nil {
		return err
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

func canonicalManifestPayload(manifest unsignedFileManifest) ([]byte, error) {
	if err := validateUnsignedManifest(&manifest); err != nil {
		return nil, err
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("canonicalize manifest: %w", err)
	}
	return data, nil
}

func signedManifestMessage(payload []byte) []byte {
	msg := make([]byte, 0, len(payload)+len("pe signed manifest v1\n"))
	msg = append(msg, "pe signed manifest v1\n"...)
	msg = append(msg, payload...)
	return msg
}

func parseEd25519PrivateKey(key string) (ed25519.PrivateKey, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(key))
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("private key must be %d bytes", ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(raw), nil
}

func parseEd25519PublicKey(key string) (ed25519.PublicKey, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(key))
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key must be %d bytes", ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(raw), nil
}

func publicKeyFingerprint(publicKey ed25519.PublicKey) string {
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:])
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
