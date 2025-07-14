package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo/security/attestation"
)

var attestCmd = &cobra.Command{
	Use:   "attest",
	Short: "Manage cryptographic attestations of prompt runs",
	Long: `Manage cryptographic attestations that prove prompt executions.

Every prompt run can be cryptographically signed and chained, creating
an immutable audit trail of all executions.`,
}

func init() {
	attestCmd.AddCommand(attestInitCmd)
	attestCmd.AddCommand(attestListCmd)
	attestCmd.AddCommand(attestVerifyCmd)
	attestCmd.AddCommand(attestShowCmd)
	attestCmd.AddCommand(attestExportCmd)
	attestCmd.AddCommand(attestKeyCmd)
}

var attestInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize attestation store",
	Long:  `Initialize the attestation store with signing keys and configuration.`,
	RunE:  runAttestInit,
}

var attestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List attestations",
	Long:  `List all attestations in the local chain.`,
	RunE:  runAttestList,
}

var attestVerifyCmd = &cobra.Command{
	Use:   "verify [attestation-id]",
	Short: "Verify attestations",
	Long: `Verify the cryptographic integrity of attestations.

Without an ID, verifies the entire chain.
With an ID, verifies a specific attestation.`,
	RunE: runAttestVerify,
}

var attestShowCmd = &cobra.Command{
	Use:   "show <attestation-id>",
	Short: "Show attestation details",
	Args:  cobra.ExactArgs(1),
	RunE:  runAttestShow,
}

var attestExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export attestation chain",
	Long: `Export the attestation chain in various formats.

Formats:
  json   - Full JSON (default)
  jsonl  - JSON Lines
  csv    - Summary CSV
  proof  - Cryptographic proof bundle`,
	RunE: runAttestExport,
}

var attestKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage attestation keys",
	Long:  `Show or rotate attestation signing keys.`,
	RunE:  runAttestKey,
}

// Flags
var (
	attestFormat      string
	attestOutput      string
	attestLimit       int
	attestSince       string
	attestProvider    string
	attestVerbose     bool
	attestKeyGenerate bool
)

func init() {
	attestListCmd.Flags().IntVar(&attestLimit, "limit", 10, "Maximum number of attestations to show")
	attestListCmd.Flags().StringVar(&attestSince, "since", "", "Show attestations since date (RFC3339)")
	attestListCmd.Flags().StringVar(&attestProvider, "provider", "", "Filter by provider")

	attestExportCmd.Flags().StringVar(&attestFormat, "format", "json", "Export format (json, jsonl, csv, proof)")
	attestExportCmd.Flags().StringVar(&attestOutput, "output", "", "Output file (default: stdout)")

	attestVerifyCmd.Flags().BoolVar(&attestVerbose, "verbose", false, "Show detailed verification info")

	attestKeyCmd.Flags().BoolVar(&attestKeyGenerate, "generate", false, "Generate new key pair")
}

func runAttestInit(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()

	// Create attestations directory
	attestDir := filepath.Join(dataDir, "attestations")
	if err := os.MkdirAll(attestDir, 0755); err != nil {
		return fmt.Errorf("creating attestations directory: %w", err)
	}

	// Create config.json
	config := map[string]interface{}{
		"version":    "1.0",
		"algorithm":  "ed25519",
		"chain_type": "linear",
		"created":    time.Now().UTC().Format(time.RFC3339),
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	configPath := filepath.Join(attestDir, "config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Initialize attestation service to generate keys
	_, err = attestation.NewAttestationService(dataDir)
	if err != nil {
		return fmt.Errorf("initializing attestation service: %w", err)
	}

	fmt.Println("Attestation store initialized")
	return nil
}

func runAttestList(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()
	_, err := attestation.NewAttestationService(dataDir)
	if err != nil {
		return fmt.Errorf("initializing attestation service: %w", err)
	}

	chainFile := filepath.Join(dataDir, "attestations", "chain.jsonl")
	file, err := os.Open(chainFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No attestations found.")
			return nil
		}
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()

	var attestations []attestation.RunAttestation
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var att attestation.RunAttestation
		if err := decoder.Decode(&att); err != nil {
			return fmt.Errorf("decoding attestation: %w", err)
		}

		// Apply filters
		if attestProvider != "" && att.Provider != attestProvider {
			continue
		}

		if attestSince != "" {
			sinceTime, err := time.Parse(time.RFC3339, attestSince)
			if err != nil {
				return fmt.Errorf("parsing since time: %w", err)
			}
			if att.Timestamp.Before(sinceTime) {
				continue
			}
		}

		attestations = append(attestations, att)
	}

	// Show latest first
	for i := len(attestations) - 1; i >= 0 && i >= len(attestations)-attestLimit; i-- {
		att := attestations[i]
		fmt.Printf("ID: %s\n", att.ID)
		fmt.Printf("Time: %s\n", att.Timestamp.Format(time.RFC3339))
		fmt.Printf("Provider: %s (Model: %s)\n", att.Provider, att.Model)
		fmt.Printf("Prompt: %s\n", truncate(att.Prompt, 60))
		fmt.Printf("Response: %s\n", truncate(att.Response, 60))
		fmt.Printf("Tokens: %d (latency: %s)\n", att.TokensUsed.Total, att.Latency)
		fmt.Printf("Hash: %s\n", att.InputHash[:16]+"...")
		fmt.Println()
	}

	fmt.Printf("Total attestations: %d\n", len(attestations))

	return nil
}

func runAttestVerify(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()
	service, err := attestation.NewAttestationService(dataDir)
	if err != nil {
		return fmt.Errorf("initializing attestation service: %w", err)
	}

	if len(args) == 0 {
		// Verify entire chain
		fmt.Println("Verifying attestation chain...")
		if err := service.VerifyChain(); err != nil {
			return fmt.Errorf("chain verification failed: %w", err)
		}
		fmt.Println("✓ Chain verified successfully")
		return nil
	}

	// Verify specific attestation
	attID := args[0]
	attFile := filepath.Join(dataDir, "attestations", attID+".json")

	data, err := os.ReadFile(attFile)
	if err != nil {
		return fmt.Errorf("reading attestation: %w", err)
	}

	var att attestation.RunAttestation
	if err := json.Unmarshal(data, &att); err != nil {
		return fmt.Errorf("parsing attestation: %w", err)
	}

	if attestVerbose {
		fmt.Printf("Verifying attestation %s...\n", attID)
		fmt.Printf("Timestamp: %s\n", att.Timestamp.Format(time.RFC3339))
		fmt.Printf("Input hash: %s\n", att.InputHash)
		fmt.Printf("Output hash: %s\n", att.OutputHash)
		fmt.Printf("Previous hash: %s\n", att.PreviousHash)
		fmt.Printf("Signature: %s...\n", att.Signature[:32])
	}

	if err := service.Verify(&att); err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	fmt.Printf("✓ Attestation %s verified successfully\n", attID)

	return nil
}

func runAttestShow(cmd *cobra.Command, args []string) error {
	attID := args[0]
	dataDir := getDataDir()
	attFile := filepath.Join(dataDir, "attestations", attID+".json")

	data, err := os.ReadFile(attFile)
	if err != nil {
		return fmt.Errorf("reading attestation: %w", err)
	}

	// Pretty print the JSON
	var att attestation.RunAttestation
	if err := json.Unmarshal(data, &att); err != nil {
		return fmt.Errorf("parsing attestation: %w", err)
	}

	output, err := json.MarshalIndent(att, "", "  ")
	if err != nil {
		return fmt.Errorf("formatting attestation: %w", err)
	}

	fmt.Println(string(output))

	return nil
}

func runAttestExport(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()

	var output *os.File
	if attestOutput != "" {
		f, err := os.Create(attestOutput)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}

	switch attestFormat {
	case "json":
		return exportJSON(dataDir, output)
	case "jsonl":
		return exportJSONL(dataDir, output)
	case "csv":
		return exportCSV(dataDir, output)
	case "proof":
		return exportProof(dataDir, output)
	default:
		return fmt.Errorf("unknown format: %s", attestFormat)
	}
}

func runAttestKey(cmd *cobra.Command, args []string) error {
	dataDir := getDataDir()

	// Check if --generate flag is set
	if attestKeyGenerate {
		service, err := attestation.NewAttestationService(dataDir)
		if err != nil {
			return fmt.Errorf("initializing attestation service: %w", err)
		}

		return service.GenerateNewKeyPair()
	}

	// Otherwise show current public key
	service, err := attestation.NewAttestationService(dataDir)
	if err != nil {
		return fmt.Errorf("initializing attestation service: %w", err)
	}

	pubKeyStr := service.GetPublicKeyString()

	fmt.Printf("Public Key: %s\n", pubKeyStr)
	fmt.Printf("Key Storage: Secure (macOS Keychain or encrypted file)\n")

	// TODO: Add key rotation functionality

	return nil
}

// Helper functions

func getDataDir() string {
	if dir := os.Getenv("PE_DATA_DIR"); dir != "" {
		return dir
	}
	return ".pe"
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func exportJSON(dataDir string, output *os.File) error {
	chainFile := filepath.Join(dataDir, "attestations", "chain.jsonl")
	file, err := os.Open(chainFile)
	if err != nil {
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()

	var attestations []attestation.RunAttestation
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var att attestation.RunAttestation
		if err := decoder.Decode(&att); err != nil {
			return fmt.Errorf("decoding attestation: %w", err)
		}
		attestations = append(attestations, att)
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(attestations)
}

func exportJSONL(dataDir string, output *os.File) error {
	chainFile := filepath.Join(dataDir, "attestations", "chain.jsonl")
	data, err := os.ReadFile(chainFile)
	if err != nil {
		return fmt.Errorf("reading chain file: %w", err)
	}

	_, err = output.Write(data)
	return err
}

func exportCSV(dataDir string, output *os.File) error {
	// CSV header
	fmt.Fprintln(output, "ID,Timestamp,Provider,Model,PromptLength,ResponseLength,TotalTokens,Latency,InputHash")

	chainFile := filepath.Join(dataDir, "attestations", "chain.jsonl")
	file, err := os.Open(chainFile)
	if err != nil {
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	for decoder.More() {
		var att attestation.RunAttestation
		if err := decoder.Decode(&att); err != nil {
			return fmt.Errorf("decoding attestation: %w", err)
		}

		fmt.Fprintf(output, "%s,%s,%s,%s,%d,%d,%d,%s,%s\n",
			att.ID,
			att.Timestamp.Format(time.RFC3339),
			att.Provider,
			att.Model,
			len(att.Prompt),
			len(att.Response),
			att.TokensUsed.Total,
			att.Latency,
			att.InputHash[:16],
		)
	}

	return nil
}

func exportProof(dataDir string, output *os.File) error {
	// Export a cryptographic proof bundle
	// This would include:
	// - The full chain
	// - Public key
	// - Merkle tree root
	// - Instructions for verification

	bundle := map[string]interface{}{
		"version": "1.0",
		"type":    "pe-attestation-proof",
		"created": time.Now().UTC(),
	}

	// Add public key
	pubKey, err := os.ReadFile(filepath.Join(dataDir, "attestations", "signing.pub"))
	if err != nil {
		return fmt.Errorf("reading public key: %w", err)
	}
	bundle["public_key"] = fmt.Sprintf("%x", pubKey)

	// Add chain
	chainData, err := exportChainData(dataDir)
	if err != nil {
		return fmt.Errorf("exporting chain data: %w", err)
	}
	bundle["chain"] = chainData

	// Add verification instructions
	bundle["verification"] = map[string]string{
		"method":         "ed25519",
		"hash_algorithm": "sha256",
		"instructions":   "Verify each attestation's signature and check chain integrity",
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bundle)
}

func exportChainData(dataDir string) ([]map[string]interface{}, error) {
	chainFile := filepath.Join(dataDir, "attestations", "chain.jsonl")
	file, err := os.Open(chainFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chain []map[string]interface{}
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var att attestation.RunAttestation
		if err := decoder.Decode(&att); err != nil {
			return nil, err
		}

		// Include only essential fields for proof
		chain = append(chain, map[string]interface{}{
			"id":            att.ID,
			"timestamp":     att.Timestamp.Unix(),
			"input_hash":    att.InputHash,
			"output_hash":   att.OutputHash,
			"previous_hash": att.PreviousHash,
			"signature":     att.Signature,
		})
	}

	return chain, nil
}
