// Package pemod provides pe.mod file parsing and manipulation functionality.
// It follows the same patterns as go.mod parsing but extends to support
// prompt engineering specific features like cryptographic signing, trust
// management, and security policies.
package pemod

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// File represents a parsed pe.mod file
type File struct {
	// Module path (optional, can be empty for local modules)
	Module *Module

	// PE version requirement
	PE *PEVersion

	// Dependencies and their constraints
	Require []Require

	// Modules to exclude
	Exclude []Exclude

	// Module replacements for development
	Replace []Replace

	// Retracted versions (for publishers)
	Retract []Retract

	// Trust settings
	Trust []Trust

	// Signing configuration
	Sign *Sign

	// Registry configuration
	Registry *Registry

	// Security policy
	Security *Security

	// Capability policy for executable text.
	Capability *Capability

	// Placement policy for executable text.
	Placement *Placement

	// Validation policy for executable text.
	Policy *Policy

	// Comments and formatting (preserved during round-trip)
	Comments []Comment

	// Syntax errors encountered during parsing
	Syntax *FileSyntax
}

// Module represents a module declaration
type Module struct {
	Mod     ModulePath
	Syntax  *ModuleSyntax
	Comment *Comment
}

// PEVersion represents the PE version requirement
type PEVersion struct {
	Version string // e.g., "1", "1.2", "1.2.3"
	Syntax  *PEVersionSyntax
	Comment *Comment
}

// Require represents a required dependency
type Require struct {
	Mod      ModulePath
	Version  string // Version constraint (e.g., "v1.0.0", "v1.2.3-beta.1")
	Indirect bool   // Added as indirect dependency
	Syntax   *RequireSyntax
	Comment  *Comment
}

// Exclude represents an excluded module version
type Exclude struct {
	Mod     ModulePath
	Version string
	Syntax  *ExcludeSyntax
	Comment *Comment
}

// Replace represents a module replacement
type Replace struct {
	Old        ModulePath
	New        ModulePath
	NewVersion string // May be empty for local replacements
	Syntax     *ReplaceSyntax
	Comment    *Comment
}

// Retract represents a retracted version range
type Retract struct {
	Low     string // Lower bound (inclusive)
	High    string // Upper bound (inclusive), may be same as Low
	Reason  string // Optional retraction reason
	Syntax  *RetractSyntax
	Comment *Comment
}

// Trust represents a trusted publisher
type Trust struct {
	Mod         ModulePath
	Fingerprint string    // Ed25519 public key fingerprint
	ValidFrom   time.Time // When trust becomes valid
	ValidUntil  time.Time // When trust expires (optional)
	Syntax      *TrustSyntax
	Comment     *Comment
}

// Sign represents signing configuration
type Sign struct {
	KeyPath   string         // Path to signing key
	Algorithm string         // Signing algorithm (ed25519, rsa-pss, etc.)
	HSM       *HSMConfig     // Hardware security module config
	Policy    *SigningPolicy // Signing policy
	Syntax    *SignSyntax
	Comment   *Comment
}

// Registry represents registry configuration
type Registry struct {
	Default string            // Default registry URL
	Mirrors []string          // Mirror URLs
	Private []string          // Private registry URLs
	Auth    map[string]string // Authentication configs per registry
	Syntax  *RegistrySyntax
	Comment *Comment
}

// Security represents security policy configuration
type Security struct {
	RequireSignatures bool             // Require all modules to be signed
	AllowUnsignedDev  bool             // Allow unsigned modules in development
	ScanContent       bool             // Enable content scanning for bias/toxicity
	Policies          []SecurityPolicy // Custom security policies
	Syntax            *SecuritySyntax
	Comment           *Comment
}

// SecurityPolicy represents a custom security policy
type SecurityPolicy struct {
	Name     string
	Type     string // "vulnerability", "license", "content", "custom"
	Config   map[string]interface{}
	Severity string // "low", "medium", "high", "critical"
	Action   string // "warn", "fail", "ignore"
}

// Capability declares module-wide allowed and denied capability classes.
type Capability struct {
	Data      AllowDeny
	Prompts   AllowDeny
	Providers AllowDeny
	Tools     AllowDeny
}

// AllowDeny is a policy dimension using simple allow and deny lists.
type AllowDeny struct {
	Allow []string
	Deny  []string
}

// Placement declares where executable text may run.
type Placement struct {
	Run            string
	Workspace      string
	Network        *bool
	DataClassRules []DataClassRule
}

// DataClassRule maps a data class to allowed provider classes.
type DataClassRule struct {
	Class     string
	Providers []string
}

// Policy declares static validation behavior.
type Policy struct {
	Composition            string
	RequireTypedIO         bool
	RequireReviewedImports bool
}

// HSMConfig represents hardware security module configuration
type HSMConfig struct {
	Provider string            // HSM provider (pkcs11, etc.)
	Config   map[string]string // Provider-specific configuration
}

// SigningPolicy represents automated signing policies
type SigningPolicy struct {
	RequireTimestamp bool     // Require timestamp authority
	AllowedAlgos     []string // Allowed signing algorithms
	MinKeySize       int      // Minimum key size
}

// ModulePath represents a module path
type ModulePath string

// Comment represents a comment in the pe.mod file
type Comment struct {
	Token  string
	Suffix bool // true if comment is on same line as syntax
}

// Syntax nodes for preserving formatting
type FileSyntax struct {
	Name  string
	Lines []Line
}

type Line interface {
	Comment() *Comment
}

type ModuleSyntax struct {
	Tok  string
	Path string
	Comm *Comment
}

func (m *ModuleSyntax) Comment() *Comment { return m.Comm }

type PEVersionSyntax struct {
	Tok     string
	Version string
	Comm    *Comment
}

func (p *PEVersionSyntax) Comment() *Comment { return p.Comm }

type RequireSyntax struct {
	Tok      string
	Path     string
	Version  string
	Indirect string
	Comm     *Comment
}

func (r *RequireSyntax) Comment() *Comment { return r.Comm }

type ExcludeSyntax struct {
	Tok     string
	Path    string
	Version string
	Comm    *Comment
}

func (e *ExcludeSyntax) Comment() *Comment { return e.Comm }

type ReplaceSyntax struct {
	Tok        string
	OldPath    string
	Arrow      string
	NewPath    string
	NewVersion string
	Comm       *Comment
}

func (r *ReplaceSyntax) Comment() *Comment { return r.Comm }

type RetractSyntax struct {
	Tok    string
	Low    string
	High   string
	Reason string
	Comm   *Comment
}

func (r *RetractSyntax) Comment() *Comment { return r.Comm }

type TrustSyntax struct {
	Tok         string
	Path        string
	Fingerprint string
	ValidFrom   string
	ValidUntil  string
	Comm        *Comment
}

func (t *TrustSyntax) Comment() *Comment { return t.Comm }

type SignSyntax struct {
	Tok   string
	Block string
	Comm  *Comment
}

func (s *SignSyntax) Comment() *Comment { return s.Comm }

type RegistrySyntax struct {
	Tok   string
	Block string
	Comm  *Comment
}

func (r *RegistrySyntax) Comment() *Comment { return r.Comm }

type SecuritySyntax struct {
	Tok   string
	Block string
	Comm  *Comment
}

func (s *SecuritySyntax) Comment() *Comment { return s.Comm }

// Parse parses a pe.mod file from the given input
func Parse(r io.Reader) (*File, error) {
	scanner := bufio.NewScanner(r)
	file := &File{}
	var lineNum int

	// State for block parsing
	var inBlock string
	var blockLines []string

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle block endings
		if inBlock != "" && (line == "}" || line == ")") {
			if err := parseBlock(file, inBlock, blockLines); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			inBlock = ""
			blockLines = nil
			continue
		}

		// If we're in a block, collect lines
		if inBlock != "" {
			blockLines = append(blockLines, line)
			continue
		}

		// Check for block start first
		if strings.HasSuffix(line, " {") || strings.HasSuffix(line, " (") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				inBlock = parts[0]
			}
			continue
		}

		// Parse top-level directives
		if err := parseDirective(file, line, lineNum); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading pe.mod: %w", err)
	}

	// Validate the parsed file
	if err := validateFile(file); err != nil {
		return nil, err
	}

	return file, nil
}

// parseDirective parses a single directive line
func parseDirective(file *File, line string, lineNum int) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "module":
		if len(parts) < 2 {
			return fmt.Errorf("module directive requires path")
		}
		file.Module = &Module{
			Mod: ModulePath(parts[1]),
		}

	case "pe":
		if len(parts) < 2 {
			return fmt.Errorf("pe directive requires version")
		}
		file.PE = &PEVersion{
			Version: parts[1],
		}

	case "require":
		return parseInlineRequire(file, parts[1:])

	case "exclude":
		return parseInlineExclude(file, parts[1:])

	case "replace":
		return parseInlineReplace(file, parts[1:])

	case "retract":
		return parseInlineRetract(file, parts[1:])

	case "trust":
		return parseInlineTrust(file, parts[1:])

	default:
		return fmt.Errorf("unknown directive: %s", parts[0])
	}

	return nil
}

// parseInlineRequire parses a single-line require directive
func parseInlineRequire(file *File, parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("require directive needs module and version")
	}

	req := Require{
		Mod:     ModulePath(parts[0]),
		Version: parts[1],
	}

	// Check for indirect comment
	if len(parts) >= 4 && parts[2] == "//" && parts[3] == "indirect" {
		req.Indirect = true
	}

	file.Require = append(file.Require, req)
	return nil
}

// parseInlineExclude parses a single-line exclude directive
func parseInlineExclude(file *File, parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("exclude directive needs module and version")
	}

	exc := Exclude{
		Mod:     ModulePath(parts[0]),
		Version: parts[1],
	}

	file.Exclude = append(file.Exclude, exc)
	return nil
}

// parseInlineReplace parses a single-line replace directive
func parseInlineReplace(file *File, parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("replace directive needs old => new")
	}

	if parts[1] != "=>" {
		return fmt.Errorf("replace directive must use '=>' separator")
	}

	repl := Replace{
		Old: ModulePath(parts[0]),
		New: ModulePath(parts[2]),
	}

	// Check for version in replacement
	if len(parts) >= 4 {
		repl.NewVersion = parts[3]
	}

	file.Replace = append(file.Replace, repl)
	return nil
}

// parseInlineRetract parses a single-line retract directive
func parseInlineRetract(file *File, parts []string) error {
	if len(parts) < 1 {
		return fmt.Errorf("retract directive needs version or range")
	}

	version := parts[0]
	var reason string

	// Check for reason comment
	if len(parts) >= 3 && parts[1] == "//" {
		reason = strings.Join(parts[2:], " ")
	}

	ret := Retract{
		Low:    version,
		High:   version,
		Reason: reason,
	}

	// Handle version ranges like [v1.0.0,v1.0.9]
	if strings.HasPrefix(version, "[") && strings.HasSuffix(version, "]") {
		rangeStr := strings.Trim(version, "[]")
		rangeParts := strings.Split(rangeStr, ",")
		if len(rangeParts) == 2 {
			ret.Low = strings.TrimSpace(rangeParts[0])
			ret.High = strings.TrimSpace(rangeParts[1])
		}
	}

	file.Retract = append(file.Retract, ret)
	return nil
}

// parseInlineTrust parses a single-line trust directive
func parseInlineTrust(file *File, parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("trust directive needs module and fingerprint")
	}

	trust := Trust{
		Mod:         ModulePath(parts[0]),
		Fingerprint: parts[1],
	}

	// Parse optional validity period
	if len(parts) >= 4 && parts[2] == "valid" {
		// Format: trust module fingerprint valid 2024-01-01 2024-12-31
		if len(parts) >= 4 {
			if t, err := time.Parse("2006-01-02", parts[3]); err == nil {
				trust.ValidFrom = t
			}
		}
		if len(parts) >= 5 {
			if t, err := time.Parse("2006-01-02", parts[4]); err == nil {
				trust.ValidUntil = t
			}
		}
	}

	file.Trust = append(file.Trust, trust)
	return nil
}

// parseBlock parses a block directive (sign, registry, security, require, etc.)
func parseBlock(file *File, blockType string, lines []string) error {
	switch blockType {
	case "require":
		return parseRequireBlock(file, lines)
	case "exclude":
		return parseExcludeBlock(file, lines)
	case "replace":
		return parseReplaceBlock(file, lines)
	case "retract":
		return parseRetractBlock(file, lines)
	case "trust":
		return parseTrustBlock(file, lines)
	case "sign":
		return parseSignBlock(file, lines)
	case "registry":
		return parseRegistryBlock(file, lines)
	case "security":
		return parseSecurityBlock(file, lines)
	case "capability":
		return parseCapabilityBlock(file, lines)
	case "placement":
		return parsePlacementBlock(file, lines)
	case "policy":
		return parsePolicyBlock(file, lines)
	default:
		return fmt.Errorf("unknown block type: %s", blockType)
	}
}

// parseRequireBlock parses a require block
func parseRequireBlock(file *File, lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if err := parseInlineRequire(file, parts); err != nil {
				return err
			}
		} else if len(parts) > 0 {
			return fmt.Errorf("require directive needs module and version")
		}
	}
	return nil
}

// parseExcludeBlock parses an exclude block
func parseExcludeBlock(file *File, lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if err := parseInlineExclude(file, parts); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseReplaceBlock parses a replace block
func parseReplaceBlock(file *File, lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			if err := parseInlineReplace(file, parts); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseRetractBlock parses a retract block
func parseRetractBlock(file *File, lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			if err := parseInlineRetract(file, parts); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseTrustBlock parses a trust block
func parseTrustBlock(file *File, lines []string) error {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if err := parseInlineTrust(file, parts); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseSignBlock parses a sign configuration block
func parseSignBlock(file *File, lines []string) error {
	sign := &Sign{}

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "key":
			sign.KeyPath = parts[1]
		case "algorithm":
			sign.Algorithm = parts[1]
		case "hsm":
			// Parse HSM configuration
			if len(parts) >= 3 {
				sign.HSM = &HSMConfig{
					Provider: parts[1],
					Config:   make(map[string]string),
				}
				// Parse additional HSM config as key=value pairs
				for i := 2; i < len(parts); i++ {
					if kv := strings.Split(parts[i], "="); len(kv) == 2 {
						sign.HSM.Config[kv[0]] = kv[1]
					}
				}
			}
		}
	}

	file.Sign = sign
	return nil
}

// parseRegistryBlock parses a registry configuration block
func parseRegistryBlock(file *File, lines []string) error {
	registry := &Registry{
		Auth: make(map[string]string),
	}

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "default":
			registry.Default = parts[1]
		case "mirror":
			registry.Mirrors = append(registry.Mirrors, parts[1])
		case "private":
			registry.Private = append(registry.Private, parts[1])
		case "auth":
			if len(parts) >= 3 {
				registry.Auth[parts[1]] = parts[2]
			}
		}
	}

	file.Registry = registry
	return nil
}

// parseSecurityBlock parses a security policy block
func parseSecurityBlock(file *File, lines []string) error {
	security := &Security{}

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "require-signatures":
			security.RequireSignatures = parseBool(parts[1])
		case "allow-unsigned-dev":
			security.AllowUnsignedDev = parseBool(parts[1])
		case "scan-content":
			security.ScanContent = parseBool(parts[1])
		case "policy":
			// Parse security policy definition
			if len(parts) >= 4 {
				policy := SecurityPolicy{
					Name: parts[1],
					Type: parts[2],
				}
				// Parse additional attributes
				for i := 3; i < len(parts); i++ {
					if kv := strings.Split(parts[i], "="); len(kv) == 2 {
						switch kv[0] {
						case "severity":
							policy.Severity = kv[1]
						case "action":
							policy.Action = kv[1]
						}
					}
				}
				security.Policies = append(security.Policies, policy)
			}
		}
	}

	file.Security = security
	return nil
}

func parseCapabilityBlock(file *File, lines []string) error {
	c := &Capability{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		var dim *AllowDeny
		switch parts[0] {
		case "data":
			dim = &c.Data
		case "prompts":
			dim = &c.Prompts
		case "providers":
			dim = &c.Providers
		case "tools":
			dim = &c.Tools
		default:
			return fmt.Errorf("unknown capability dimension: %s", parts[0])
		}
		switch parts[1] {
		case "allow":
			dim.Allow = append(dim.Allow, parts[2:]...)
		case "deny":
			dim.Deny = append(dim.Deny, parts[2:]...)
		default:
			return fmt.Errorf("unknown capability operation: %s", parts[1])
		}
	}
	file.Capability = c
	return nil
}

func parsePlacementBlock(file *File, lines []string) error {
	p := &Placement{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "run":
			p.Run = parts[1]
		case "workspace":
			p.Workspace = parts[1]
		case "network":
			v := parseBool(parts[1])
			p.Network = &v
		case "data-class":
			if len(parts) < 5 || parts[2] != "=>" || parts[3] != "providers" {
				return fmt.Errorf("data-class rule must be: data-class <class> => providers <provider>...")
			}
			p.DataClassRules = append(p.DataClassRules, DataClassRule{
				Class:     parts[1],
				Providers: append([]string(nil), parts[4:]...),
			})
		default:
			return fmt.Errorf("unknown placement directive: %s", parts[0])
		}
	}
	file.Placement = p
	return nil
}

func parsePolicyBlock(file *File, lines []string) error {
	p := &Policy{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		switch parts[0] {
		case "composition":
			p.Composition = parts[1]
		case "require-typed-io":
			p.RequireTypedIO = parseBool(parts[1])
		case "require-reviewed-imports":
			p.RequireReviewedImports = parseBool(parts[1])
		default:
			return fmt.Errorf("unknown policy directive: %s", parts[0])
		}
	}
	file.Policy = p
	return nil
}

// parseBool parses a boolean value from string
func parseBool(s string) bool {
	val, _ := strconv.ParseBool(s)
	return val
}

// validateFile validates the parsed pe.mod file
func validateFile(file *File) error {
	// PE version is required
	if file.PE == nil {
		return fmt.Errorf("pe.mod must specify PE version")
	}

	// Validate version format
	if !isValidPEVersion(file.PE.Version) {
		return fmt.Errorf("invalid PE version: %s", file.PE.Version)
	}

	// Check for duplicate requirements
	seen := make(map[ModulePath]bool)
	for _, req := range file.Require {
		if seen[req.Mod] {
			return fmt.Errorf("duplicate requirement for module: %s", req.Mod)
		}
		seen[req.Mod] = true
	}

	return nil
}

// isValidPEVersion checks if a PE version string is valid
func isValidPEVersion(version string) bool {
	// For now, accept simple numeric versions like "1", "1.2", "1.2.3"
	parts := strings.Split(version, ".")
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return len(parts) >= 1 && len(parts) <= 3
}

// Format formats a pe.mod file back to text
func (f *File) Format() string {
	var buf strings.Builder

	// Module declaration
	if f.Module != nil {
		fmt.Fprintf(&buf, "module %s\n\n", f.Module.Mod)
	}

	// PE version
	if f.PE != nil {
		fmt.Fprintf(&buf, "pe %s\n\n", f.PE.Version)
	}

	// Requirements
	if len(f.Require) > 0 {
		if len(f.Require) == 1 && !f.Require[0].Indirect {
			// Single require on one line
			req := f.Require[0]
			fmt.Fprintf(&buf, "require %s %s\n\n", req.Mod, req.Version)
		} else {
			// Multi-line require block
			buf.WriteString("require (\n")
			for _, req := range f.Require {
				if req.Indirect {
					fmt.Fprintf(&buf, "\t%s %s // indirect\n", req.Mod, req.Version)
				} else {
					fmt.Fprintf(&buf, "\t%s %s\n", req.Mod, req.Version)
				}
			}
			buf.WriteString(")\n\n")
		}
	}

	// Replacements
	for _, repl := range f.Replace {
		if repl.NewVersion != "" {
			fmt.Fprintf(&buf, "replace %s => %s %s\n", repl.Old, repl.New, repl.NewVersion)
		} else {
			fmt.Fprintf(&buf, "replace %s => %s\n", repl.Old, repl.New)
		}
	}
	if len(f.Replace) > 0 {
		buf.WriteString("\n")
	}

	// Exclusions
	for _, exc := range f.Exclude {
		fmt.Fprintf(&buf, "exclude %s %s\n", exc.Mod, exc.Version)
	}
	if len(f.Exclude) > 0 {
		buf.WriteString("\n")
	}

	// Retractions
	for _, ret := range f.Retract {
		if ret.Low == ret.High {
			if ret.Reason != "" {
				fmt.Fprintf(&buf, "retract %s // %s\n", ret.Low, ret.Reason)
			} else {
				fmt.Fprintf(&buf, "retract %s\n", ret.Low)
			}
		} else {
			if ret.Reason != "" {
				fmt.Fprintf(&buf, "retract [%s,%s] // %s\n", ret.Low, ret.High, ret.Reason)
			} else {
				fmt.Fprintf(&buf, "retract [%s,%s]\n", ret.Low, ret.High)
			}
		}
	}
	if len(f.Retract) > 0 {
		buf.WriteString("\n")
	}

	// Trust settings
	for _, trust := range f.Trust {
		if !trust.ValidFrom.IsZero() && !trust.ValidUntil.IsZero() {
			fmt.Fprintf(&buf, "trust %s %s valid %s %s\n",
				trust.Mod, trust.Fingerprint,
				trust.ValidFrom.Format("2006-01-02"),
				trust.ValidUntil.Format("2006-01-02"))
		} else {
			fmt.Fprintf(&buf, "trust %s %s\n", trust.Mod, trust.Fingerprint)
		}
	}
	if len(f.Trust) > 0 {
		buf.WriteString("\n")
	}

	// Signing configuration
	if f.Sign != nil {
		buf.WriteString("sign {\n")
		if f.Sign.KeyPath != "" {
			fmt.Fprintf(&buf, "\tkey %s\n", f.Sign.KeyPath)
		}
		if f.Sign.Algorithm != "" {
			fmt.Fprintf(&buf, "\talgorithm %s\n", f.Sign.Algorithm)
		}
		if f.Sign.HSM != nil {
			fmt.Fprintf(&buf, "\thsm %s", f.Sign.HSM.Provider)
			for k, v := range f.Sign.HSM.Config {
				fmt.Fprintf(&buf, " %s=%s", k, v)
			}
			buf.WriteString("\n")
		}
		buf.WriteString("}\n\n")
	}

	// Registry configuration
	if f.Registry != nil {
		buf.WriteString("registry {\n")
		if f.Registry.Default != "" {
			fmt.Fprintf(&buf, "\tdefault %s\n", f.Registry.Default)
		}
		for _, mirror := range f.Registry.Mirrors {
			fmt.Fprintf(&buf, "\tmirror %s\n", mirror)
		}
		for _, private := range f.Registry.Private {
			fmt.Fprintf(&buf, "\tprivate %s\n", private)
		}
		for registry, auth := range f.Registry.Auth {
			fmt.Fprintf(&buf, "\tauth %s %s\n", registry, auth)
		}
		buf.WriteString("}\n\n")
	}

	// Security configuration
	if f.Security != nil {
		buf.WriteString("security {\n")
		fmt.Fprintf(&buf, "\trequire-signatures %t\n", f.Security.RequireSignatures)
		fmt.Fprintf(&buf, "\tallow-unsigned-dev %t\n", f.Security.AllowUnsignedDev)
		fmt.Fprintf(&buf, "\tscan-content %t\n", f.Security.ScanContent)
		for _, policy := range f.Security.Policies {
			fmt.Fprintf(&buf, "\tpolicy %s %s severity=%s action=%s\n",
				policy.Name, policy.Type, policy.Severity, policy.Action)
		}
		buf.WriteString("}\n\n")
	}

	if f.Capability != nil {
		buf.WriteString("capability {\n")
		formatAllowDeny(&buf, "data", f.Capability.Data)
		formatAllowDeny(&buf, "prompts", f.Capability.Prompts)
		formatAllowDeny(&buf, "providers", f.Capability.Providers)
		formatAllowDeny(&buf, "tools", f.Capability.Tools)
		buf.WriteString("}\n\n")
	}

	if f.Placement != nil {
		buf.WriteString("placement {\n")
		if f.Placement.Run != "" {
			fmt.Fprintf(&buf, "\trun %s\n", f.Placement.Run)
		}
		if f.Placement.Workspace != "" {
			fmt.Fprintf(&buf, "\tworkspace %s\n", f.Placement.Workspace)
		}
		if f.Placement.Network != nil {
			fmt.Fprintf(&buf, "\tnetwork %t\n", *f.Placement.Network)
		}
		for _, rule := range f.Placement.DataClassRules {
			fmt.Fprintf(&buf, "\tdata-class %s => providers %s\n", rule.Class, strings.Join(rule.Providers, " "))
		}
		buf.WriteString("}\n\n")
	}

	if f.Policy != nil {
		buf.WriteString("policy {\n")
		if f.Policy.Composition != "" {
			fmt.Fprintf(&buf, "\tcomposition %s\n", f.Policy.Composition)
		}
		fmt.Fprintf(&buf, "\trequire-typed-io %t\n", f.Policy.RequireTypedIO)
		fmt.Fprintf(&buf, "\trequire-reviewed-imports %t\n", f.Policy.RequireReviewedImports)
		buf.WriteString("}\n\n")
	}

	return strings.TrimSpace(buf.String()) + "\n"
}

func formatAllowDeny(buf *strings.Builder, name string, v AllowDeny) {
	if len(v.Allow) > 0 {
		fmt.Fprintf(buf, "\t%s allow %s\n", name, strings.Join(v.Allow, " "))
	}
	if len(v.Deny) > 0 {
		fmt.Fprintf(buf, "\t%s deny %s\n", name, strings.Join(v.Deny, " "))
	}
}

// AddRequire adds a requirement to the file
func (f *File) AddRequire(path string, version string) {
	// Check if already exists
	for i, req := range f.Require {
		if req.Mod == ModulePath(path) {
			f.Require[i].Version = version
			return
		}
	}

	// Add new requirement
	f.Require = append(f.Require, Require{
		Mod:     ModulePath(path),
		Version: version,
	})

	// Sort requirements
	sort.Slice(f.Require, func(i, j int) bool {
		return f.Require[i].Mod < f.Require[j].Mod
	})
}

// RemoveRequire removes a requirement from the file
func (f *File) RemoveRequire(path string) {
	for i, req := range f.Require {
		if req.Mod == ModulePath(path) {
			f.Require = append(f.Require[:i], f.Require[i+1:]...)
			return
		}
	}
}

// GetRequire returns the requirement for a module path
func (f *File) GetRequire(path string) *Require {
	for _, req := range f.Require {
		if req.Mod == ModulePath(path) {
			return &req
		}
	}
	return nil
}

// SetModule sets the module path
func (f *File) SetModule(path string) {
	if f.Module == nil {
		f.Module = &Module{}
	}
	f.Module.Mod = ModulePath(path)
}

// SetPEVersion sets the PE version
func (f *File) SetPEVersion(version string) {
	if f.PE == nil {
		f.PE = &PEVersion{}
	}
	f.PE.Version = version
}
