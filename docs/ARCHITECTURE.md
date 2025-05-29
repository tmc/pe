# PE Architecture

This document describes the internal architecture of PE (Go for Prompts), including design decisions, component interactions, and extension points.

## Design Principles

PE's architecture follows these core principles:

1. **Modularity**: Clear separation of concerns with well-defined interfaces
2. **Extensibility**: Plugin system for custom providers and tools
3. **Performance**: Efficient Go implementation with minimal overhead
4. **Security**: Defense in depth with sandboxing and trust verification
5. **Simplicity**: Clean APIs that are easy to understand and use

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐          │
│  │   run   │ │   test   │ │ build  │ │optimize │   ...    │
│  └────┬────┘ └────┬─────┘ └───┬────┘ └────┬────┘          │
│       └───────────┴───────────┴────────────┘               │
│                         │                                    │
├─────────────────────────┼────────────────────────────────────┤
│                    Core Engine                               │
│  ┌─────────────────────┴──────────────────────┐            │
│  │           Command Dispatcher                 │            │
│  └─────────────────────┬──────────────────────┘            │
│  ┌──────────┬──────────┼──────────┬──────────┐            │
│  │ Provider │ Version  │ Cache    │ Plugin   │            │
│  │ Manager  │ Control  │ System   │ Registry │            │
│  └──────────┴──────────┴──────────┴──────────┘            │
├──────────────────────────────────────────────────────────────┤
│                   Security Layer                             │
│  ┌──────────────┐ ┌─────────────┐ ┌──────────────┐        │
│  │   Sandbox    │ │ Trust Store │ │  Signature   │        │
│  │   Manager    │ │             │ │ Verification │        │
│  └──────────────┘ └─────────────┘ └──────────────┘        │
├──────────────────────────────────────────────────────────────┤
│                  Storage Layer                               │
│  ┌──────────────┐ ┌─────────────┐ ┌──────────────┐        │
│  │ File System  │ │   Git-like  │ │   Content    │        │
│  │   Handler    │ │    Store    │ │ Addressable  │        │
│  └──────────────┘ └─────────────┘ └──────────────┘        │
└──────────────────────────────────────────────────────────────┘
```

## Core Components

### CLI Layer (`cmd/pe/`)

The CLI layer provides the user interface:

```go
// cmd/pe/main.go
func main() {
    rootCmd := &cobra.Command{
        Use:   "pe",
        Short: "Go for Prompts - The unified toolchain for prompt engineering",
    }
    
    // Register all subcommands
    rootCmd.AddCommand(
        runCmd(),
        testCmd(),
        buildCmd(),
        optimizeCmd(),
        // ... more commands
    )
}
```

Each command follows a consistent pattern:
- Parse flags and arguments
- Validate inputs
- Call appropriate core engine method
- Format and display output

### Provider Interface (`internal/providers/`)

The provider interface abstracts LLM interactions:

```go
// internal/providers/provider.go
type Provider interface {
    // Complete generates a completion for the given prompt
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    
    // Stream generates a streaming completion
    Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    
    // Embed generates embeddings for text
    Embed(ctx context.Context, text []string) ([][]float32, error)
    
    // Models returns available models
    Models(ctx context.Context) ([]Model, error)
    
    // Validate checks if the provider is properly configured
    Validate() error
}

// Completion request structure
type CompletionRequest struct {
    Prompt      string
    Model       string
    Temperature float32
    MaxTokens   int
    Variables   map[string]string
    Stream      bool
}
```

### Version Control System (`internal/vcs/`)

PE implements a Git-like version control system:

```go
// internal/vcs/repository.go
type Repository struct {
    path     string
    objects  ObjectStore
    refs     RefStore
    index    Index
    worktree Worktree
}

// Object types
type Object interface {
    Type() ObjectType
    Hash() Hash
    Serialize() []byte
}

type Commit struct {
    Tree      Hash
    Parent    []Hash
    Author    Signature
    Message   string
    Metadata  map[string]string
}

type Tree struct {
    Entries []TreeEntry
}

type Blob struct {
    Content []byte
}
```

### Cache System (`internal/cache/`)

The cache system provides verifiable response caching:

```go
// internal/cache/cache.go
type Cache interface {
    // Get retrieves a cached response
    Get(key CacheKey) (*CacheEntry, error)
    
    // Put stores a response with signature
    Put(key CacheKey, entry *CacheEntry) error
    
    // Verify checks entry signatures
    Verify(entry *CacheEntry) error
    
    // Export creates a portable cache bundle
    Export(filter Filter) (*Bundle, error)
    
    // Import merges an external cache
    Import(bundle *Bundle, verify bool) error
}

type CacheEntry struct {
    Request   CompletionRequest
    Response  CompletionResponse
    Timestamp time.Time
    Signature Signature
    Witnesses []Witness
}
```

### Plugin System (`internal/plugins/`)

Plugins extend PE functionality:

```go
// internal/plugins/plugin.go
type Plugin interface {
    // Metadata returns plugin information
    Metadata() Metadata
    
    // Initialize prepares the plugin
    Initialize(config map[string]interface{}) error
    
    // Execute runs the plugin
    Execute(ctx context.Context, args []string) error
    
    // Shutdown cleans up resources
    Shutdown() error
}

// Provider plugin interface
type ProviderPlugin interface {
    Plugin
    Provider
}

// Tool plugin interface  
type ToolPlugin interface {
    Plugin
    Run(input io.Reader, output io.Writer) error
}
```

### Optimization Engine (`internal/optimize/`)

The optimization engine implements various methods:

```go
// internal/optimize/optimizer.go
type Optimizer interface {
    // Optimize improves a prompt
    Optimize(ctx context.Context, prompt string, config Config) (*Result, error)
    
    // SupportsObjective checks if an objective is supported
    SupportsObjective(objective string) bool
}

// Optimization methods
type PE2Optimizer struct{}
type TextGradOptimizer struct{}
type APEXOptimizer struct{}
type EvolutionaryOptimizer struct{}

// Factory pattern for optimizer creation
func NewOptimizer(method string) (Optimizer, error) {
    switch method {
    case "pe2":
        return &PE2Optimizer{}, nil
    case "textgrad":
        return &TextGradOptimizer{}, nil
    // ... more methods
    }
}
```

### Security Layer (`internal/security/`)

Security components protect the system:

```go
// internal/security/sandbox.go
type Sandbox interface {
    // Execute runs a command in the sandbox
    Execute(cmd *exec.Cmd) error
    
    // GrantPath allows access to a path
    GrantPath(path string, mode AccessMode) error
    
    // RevokePath removes path access
    RevokePath(path string) error
}

// internal/security/trust.go
type TrustStore interface {
    // IsTrusted checks if a binary is trusted
    IsTrusted(path string) bool
    
    // AddTrust marks a binary as trusted
    AddTrust(path string, signature []byte) error
    
    // VerifySignature checks binary signature
    VerifySignature(path string) error
}
```

## Data Flow

### 1. Command Execution Flow

```
User Input → CLI Parser → Command Handler → Core Engine → Provider → Response
                                    ↓
                            Security Check
                                    ↓
                            Cache Lookup
                                    ↓
                              Execution
```

### 2. Optimization Flow

```
Input Prompt → Parser → Optimizer Selection → Method Execution
                                                    ↓
                                            Iteration Loop
                                                    ↓
                                         Evaluation & Scoring
                                                    ↓
                                         Convergence Check
                                                    ↓
                                           Result Output
```

### 3. Version Control Flow

```
Working Directory → Index → Object Store → References
         ↓                        ↓              ↓
     File Changes            Content Hash    Branch/Tag
         ↓                        ↓              ↓
     Stage Changes          Commit Object    Update HEAD
```

## File Formats

### PE Module Format (`pe.mod`)

```
module github.com/user/my-prompts

go 1.21

require (
    github.com/org/prompt-utils v1.2.3
    gist:user/templates v0.1.0
)

replace gist:user/templates => ./vendor/templates
```

### txtar Format

PE uses Go's txtar format for multi-file archives:

```
-- prompt.txt --
Main prompt content

-- config.yaml --
provider: gpt-4
temperature: 0.7

-- tests.yaml --
tests:
  - input: "test"
    expect: "response"
```

### Cache Entry Format

```json
{
  "version": "1.0",
  "entries": [
    {
      "key": "sha256:abc123...",
      "request": {
        "prompt": "...",
        "model": "gpt-4",
        "temperature": 0.7
      },
      "response": {
        "content": "...",
        "tokens": 100,
        "latency_ms": 1234
      },
      "metadata": {
        "timestamp": "2024-01-20T10:00:00Z",
        "signature": "sig:xyz789...",
        "signer": "alice@example.com",
        "witnesses": ["bob@example.com"]
      }
    }
  ]
}
```

## Extension Points

### 1. Custom Providers

Implement the Provider interface:

```go
type MyProvider struct {
    endpoint string
    apiKey   string
}

func (p *MyProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
    // Custom implementation
}

// Register the provider
func init() {
    providers.Register("myprovider", func(config map[string]string) Provider {
        return &MyProvider{
            endpoint: config["endpoint"],
            apiKey:   config["api_key"],
        }
    })
}
```

### 2. Custom Optimizers

Implement the Optimizer interface:

```go
type MyOptimizer struct{}

func (o *MyOptimizer) Optimize(ctx context.Context, prompt string, config Config) (*Result, error) {
    // Custom optimization logic
}

// Register the optimizer
func init() {
    optimizers.Register("mymethod", func() Optimizer {
        return &MyOptimizer{}
    })
}
```

### 3. Style Validators

Create custom validators:

```go
type StyleValidator interface {
    Validate(content string, rules []Rule) ([]Violation, error)
}

type MyValidator struct{}

func (v *MyValidator) Validate(content string, rules []Rule) ([]Violation, error) {
    // Custom validation logic
}
```

## Performance Considerations

### 1. Caching Strategy

- Content-addressed storage for deduplication
- LRU eviction for memory-bounded caches
- Bloom filters for quick existence checks
- Merkle trees for efficient verification

### 2. Concurrency

- Provider calls use connection pooling
- Parallel test execution with worker pools
- Lock-free cache reads
- Copy-on-write for version control

### 3. Memory Management

- Streaming for large responses
- Chunked file processing
- Incremental optimization updates
- Garbage collection tuning for large datasets

## Security Model

### 1. Threat Model

PE protects against:
- Malicious prompt injection
- Untrusted binary execution
- Cache poisoning
- Version control tampering
- Plugin vulnerabilities

### 2. Security Measures

- **Sandboxing**: All operations run in restricted environments
- **Code Signing**: Binaries must be signed and trusted
- **Input Validation**: All inputs are sanitized
- **Cryptographic Verification**: Signatures on all shared data
- **Least Privilege**: Minimal permissions requested

### 3. Trust Boundaries

```
Trusted Zone:
├── PE Core Binary (signed)
├── Trusted Plugins (verified)
└── User Configuration

Untrusted Zone:
├── External Prompts
├── API Responses
├── Imported Caches
└── Third-party Plugins
```

## Future Architecture Considerations

### 1. Distributed Operations

- Distributed cache protocol
- Federated version control
- Peer-to-peer optimization
- Consensus mechanisms

### 2. Advanced Features

- Real-time collaboration
- Federated learning for optimization
- Homomorphic encryption for private prompts
- Zero-knowledge proofs for verification

### 3. Scalability

- Horizontal scaling for API calls
- Sharded cache storage
- Distributed optimization algorithms
- Cloud-native deployment options

## Development Guidelines

### 1. Adding a New Command

1. Create command file in `cmd/pe/`
2. Implement command logic
3. Add to root command
4. Write tests
5. Update documentation

### 2. Adding a New Provider

1. Implement Provider interface
2. Add to provider registry
3. Write integration tests
4. Document configuration
5. Submit PR with examples

### 3. Code Organization

```
internal/
├── commands/     # Command implementations
├── core/         # Core business logic
├── providers/    # LLM providers
├── optimize/     # Optimization methods
├── vcs/          # Version control
├── cache/        # Caching system
├── security/     # Security components
├── plugins/      # Plugin system
└── utils/        # Shared utilities
```

## Testing Architecture

### 1. Test Levels

- **Unit Tests**: Individual component testing
- **Integration Tests**: Component interaction testing
- **E2E Tests**: Full workflow testing
- **Benchmark Tests**: Performance testing

### 2. Test Infrastructure

- Mock providers for offline testing
- Scripttest for CLI testing
- Fuzzing for security testing
- Property-based testing for optimization

## Conclusion

PE's architecture provides a solid foundation for prompt engineering at scale. The modular design allows for easy extension while maintaining security and performance. The Go implementation ensures efficiency and reliability for production use.