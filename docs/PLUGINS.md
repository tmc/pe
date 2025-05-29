# PE Plugin Development Guide

This guide covers creating plugins to extend PE with custom providers, tools, and functionality.

## Overview

PE's plugin system allows you to:
- Add custom LLM providers (local models, proprietary APIs)
- Create new optimization methods
- Implement custom validators and metrics
- Add new commands and tools
- Integrate with external services

## Plugin Types

### 1. Provider Plugins

Provider plugins add support for new LLM backends.

```go
package main

import (
    "context"
    "github.com/tmc/pe/pkg/plugin"
    "github.com/tmc/pe/pkg/provider"
)

type MyProvider struct {
    endpoint string
    apiKey   string
}

func (p *MyProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Implement completion logic
    return &provider.CompletionResponse{
        Content: "Response from my provider",
        Tokens:  10,
        Model:   req.Model,
    }, nil
}

func (p *MyProvider) Stream(ctx context.Context, req provider.CompletionRequest) (<-chan provider.StreamChunk, error) {
    ch := make(chan provider.StreamChunk)
    go func() {
        defer close(ch)
        // Stream implementation
        ch <- provider.StreamChunk{Content: "Streaming response"}
    }()
    return ch, nil
}

// Plugin entry point
func NewPlugin() plugin.Plugin {
    return &plugin.ProviderPlugin{
        Name: "myprovider",
        Provider: &MyProvider{},
    }
}

// Required for plugin loading
var Plugin = NewPlugin()
```

### 2. Optimizer Plugins

Add new optimization methods:

```go
package main

import (
    "context"
    "github.com/tmc/pe/pkg/optimize"
    "github.com/tmc/pe/pkg/plugin"
)

type MyOptimizer struct {
    iterations int
}

func (o *MyOptimizer) Optimize(ctx context.Context, prompt string, config optimize.Config) (*optimize.Result, error) {
    improved := prompt
    score := 0.5
    
    for i := 0; i < o.iterations; i++ {
        // Your optimization logic here
        improved = enhancePrompt(improved)
        score = evaluatePrompt(improved)
        
        if score > config.TargetScore {
            break
        }
    }
    
    return &optimize.Result{
        Prompt: improved,
        Score:  score,
        Metadata: map[string]interface{}{
            "iterations": i,
            "method":     "myoptimizer",
        },
    }, nil
}

func NewPlugin() plugin.Plugin {
    return &plugin.OptimizerPlugin{
        Name: "myoptimizer",
        Optimizer: &MyOptimizer{
            iterations: 10,
        },
    }
}

var Plugin = NewPlugin()
```

### 3. Validator Plugins

Create custom style validators:

```go
package main

import (
    "github.com/tmc/pe/pkg/plugin"
    "github.com/tmc/pe/pkg/validate"
)

type ToneValidator struct {
    targetTone string
}

func (v *ToneValidator) Validate(content string) (validate.Result, error) {
    // Analyze tone of content
    detectedTone := analyzeTone(content)
    
    if detectedTone != v.targetTone {
        return validate.Result{
            Valid: false,
            Issues: []validate.Issue{
                {
                    Type:     "tone_mismatch",
                    Severity: validate.Warning,
                    Message:  fmt.Sprintf("Expected %s tone, got %s", v.targetTone, detectedTone),
                    Line:     1,
                },
            },
        }, nil
    }
    
    return validate.Result{Valid: true}, nil
}

func NewPlugin() plugin.Plugin {
    return &plugin.ValidatorPlugin{
        Name: "tone-validator",
        Validator: &ToneValidator{
            targetTone: "professional",
        },
    }
}

var Plugin = NewPlugin()
```

### 4. Command Plugins

Add new PE commands:

```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/tmc/pe/pkg/plugin"
)

func NewPlugin() plugin.Plugin {
    return &plugin.CommandPlugin{
        Name: "analyze-complexity",
        Command: &cobra.Command{
            Use:   "complexity [file]",
            Short: "Analyze prompt complexity",
            RunE: func(cmd *cobra.Command, args []string) error {
                if len(args) < 1 {
                    return fmt.Errorf("file required")
                }
                
                content, err := os.ReadFile(args[0])
                if err != nil {
                    return err
                }
                
                complexity := analyzeComplexity(string(content))
                fmt.Printf("Complexity Score: %.2f\n", complexity)
                
                return nil
            },
        },
    }
}

var Plugin = NewPlugin()
```

## Plugin Development

### Step 1: Set Up Project

```bash
# Create plugin directory
mkdir pe-plugin-myprovider
cd pe-plugin-myprovider

# Initialize module
go mod init github.com/myuser/pe-plugin-myprovider

# Add PE plugin SDK
go get github.com/tmc/pe/pkg/plugin
```

### Step 2: Implement Plugin

Create `main.go`:

```go
package main

import (
    "github.com/tmc/pe/pkg/plugin"
    // Your imports
)

// Implement your plugin type

func NewPlugin() plugin.Plugin {
    // Return your plugin implementation
}

// Required export
var Plugin = NewPlugin()
```

### Step 3: Build Plugin

```bash
# Build as shared library
go build -buildmode=plugin -o myprovider.so

# Or use PE tool
pe plugin build .
```

### Step 4: Test Plugin

Create `plugin_test.go`:

```go
package main

import (
    "testing"
    "github.com/tmc/pe/pkg/plugin/test"
)

func TestPlugin(t *testing.T) {
    p := NewPlugin()
    
    // Test metadata
    meta := p.Metadata()
    if meta.Name != "myprovider" {
        t.Errorf("wrong name: %s", meta.Name)
    }
    
    // Test functionality
    test.TestProvider(t, p.(*plugin.ProviderPlugin).Provider)
}
```

Run tests:

```bash
go test -v
```

### Step 5: Package Plugin

Create `plugin.yaml`:

```yaml
name: myprovider
version: 1.0.0
description: Custom provider for MyLLM
author: Your Name
license: MIT
homepage: https://github.com/myuser/pe-plugin-myprovider

requirements:
  pe: ">=1.0.0"
  go: ">=1.21"

configuration:
  - name: endpoint
    type: string
    required: true
    description: API endpoint URL
  - name: api_key
    type: string
    required: true
    description: API key for authentication
    secret: true

files:
  - myprovider.so
  - README.md
  - LICENSE
```

Package it:

```bash
pe plugin package .
# Creates: myprovider-1.0.0.tar.gz
```

## Plugin Configuration

### User Configuration

Users configure plugins in `~/.pe/config.yaml`:

```yaml
plugins:
  myprovider:
    endpoint: https://api.myllm.com/v1
    api_key: ${MYLLM_API_KEY}
    
  tone-validator:
    default_tone: professional
    strict_mode: true
```

### Accessing Configuration

```go
func (p *MyProvider) Initialize(config map[string]interface{}) error {
    if endpoint, ok := config["endpoint"].(string); ok {
        p.endpoint = endpoint
    }
    
    if apiKey, ok := config["api_key"].(string); ok {
        p.apiKey = apiKey
    }
    
    return p.validate()
}
```

## Plugin SDK Reference

### Core Interfaces

```go
// Base plugin interface
type Plugin interface {
    Metadata() Metadata
    Initialize(config map[string]interface{}) error
    Shutdown() error
}

// Plugin metadata
type Metadata struct {
    Name        string
    Version     string
    Description string
    Author      string
    License     string
    Homepage    string
}

// Provider plugin
type ProviderPlugin struct {
    Plugin
    provider.Provider
}

// Optimizer plugin
type OptimizerPlugin struct {
    Plugin
    optimize.Optimizer
}

// Validator plugin
type ValidatorPlugin struct {
    Plugin
    validate.Validator
}

// Command plugin
type CommandPlugin struct {
    Plugin
    *cobra.Command
}
```

### Helper Functions

```go
// Load configuration value
func GetConfig[T any](config map[string]interface{}, key string, defaultValue T) T

// Log messages
func Log(level, message string, args ...interface{})

// Register metrics
func RegisterMetric(name string, metric Metric)

// Access PE services
func GetCache() cache.Cache
func GetVCS() vcs.Repository
```

## Best Practices

### 1. Error Handling

```go
func (p *MyProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Validate request
    if err := p.validateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Handle API errors
    resp, err := p.callAPI(req)
    if err != nil {
        // Wrap with context
        return nil, fmt.Errorf("API call failed: %w", err)
    }
    
    return resp, nil
}
```

### 2. Resource Management

```go
type MyProvider struct {
    client *http.Client
    pool   *connectionPool
}

func (p *MyProvider) Initialize(config map[string]interface{}) error {
    // Create resources
    p.client = &http.Client{
        Timeout: 30 * time.Second,
    }
    
    p.pool = newConnectionPool(10)
    
    return nil
}

func (p *MyProvider) Shutdown() error {
    // Clean up resources
    if p.pool != nil {
        return p.pool.Close()
    }
    return nil
}
```

### 3. Testing

```go
// Mock provider for testing
type MockProvider struct {
    responses map[string]string
}

func (m *MockProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
    if resp, ok := m.responses[req.Prompt]; ok {
        return &provider.CompletionResponse{
            Content: resp,
        }, nil
    }
    return nil, fmt.Errorf("no mock response")
}

// Test with mock
func TestOptimizer(t *testing.T) {
    mock := &MockProvider{
        responses: map[string]string{
            "test prompt": "test response",
        },
    }
    
    opt := &MyOptimizer{provider: mock}
    result, err := opt.Optimize(context.Background(), "test prompt", optimize.Config{})
    
    if err != nil {
        t.Fatal(err)
    }
    
    if result.Score < 0.8 {
        t.Errorf("low score: %f", result.Score)
    }
}
```

### 4. Documentation

Include comprehensive documentation:

```markdown
# PE Plugin: MyProvider

## Installation

```bash
pe plugin install github.com/myuser/pe-plugin-myprovider
```

## Configuration

Add to ~/.pe/config.yaml:

```yaml
plugins:
  myprovider:
    endpoint: https://api.example.com
    api_key: your-key-here
```

## Usage

```bash
pe run "Hello" --provider myprovider
```

## Features

- Streaming support
- Custom models
- Token counting
- Error retry

## API

See [API Documentation](docs/api.md)
```

## Publishing Plugins

### 1. GitHub Repository

```bash
# Create repository
git init
git add .
git commit -m "Initial plugin"
git remote add origin https://github.com/myuser/pe-plugin-myprovider
git push -u origin main

# Tag release
git tag v1.0.0
git push origin v1.0.0
```

### 2. Plugin Registry

Submit to PE plugin registry:

```bash
# Create registry entry
cat > registry.yaml << EOF
name: myprovider
repository: github.com/myuser/pe-plugin-myprovider
description: Custom provider for MyLLM
categories:
  - provider
  - llm
tags:
  - custom
  - api
EOF

# Submit PR to pe-plugins repository
```

### 3. Distribution

Users can install via:

```bash
# From GitHub
pe plugin install github.com/myuser/pe-plugin-myprovider

# From registry
pe plugin search myprovider
pe plugin install myprovider

# From local file
pe plugin install ./myprovider-1.0.0.tar.gz
```

## Advanced Topics

### 1. Plugin Communication

Plugins can communicate via PE's event system:

```go
// Emit event
plugin.EmitEvent("optimization.complete", map[string]interface{}{
    "prompt": improved,
    "score":  score,
})

// Subscribe to events
plugin.Subscribe("cache.hit", func(data map[string]interface{}) {
    // Handle cache hit event
})
```

### 2. Shared State

Access shared state safely:

```go
// Get shared state
state := plugin.GetState()

// Update atomically
state.Update("counter", func(val interface{}) interface{} {
    if v, ok := val.(int); ok {
        return v + 1
    }
    return 1
})
```

### 3. Custom Metrics

Export metrics for monitoring:

```go
// Counter
completeCount := plugin.NewCounter("myprovider_completions_total")
completeCount.Inc()

// Histogram
latency := plugin.NewHistogram("myprovider_latency_seconds")
latency.Observe(time.Since(start).Seconds())

// Gauge
queueSize := plugin.NewGauge("myprovider_queue_size")
queueSize.Set(float64(len(queue)))
```

## Troubleshooting

### Common Issues

1. **Plugin won't load**
   - Check Go version compatibility
   - Verify plugin was built with same Go version as PE
   - Check for missing dependencies

2. **Initialization fails**
   - Verify configuration format
   - Check required fields
   - Look at PE logs with `PE_DEBUG=true`

3. **Performance issues**
   - Profile with `go tool pprof`
   - Check for blocking operations
   - Use connection pooling

### Debug Mode

Enable debug logging:

```bash
PE_DEBUG=true pe run "test" --provider myprovider
```

## Examples

See the [examples/plugins](../examples/plugins/) directory for complete plugin examples:

- `ollama-provider/` - Local model provider
- `grammarly-validator/` - Grammar checking validator  
- `cost-optimizer/` - Cost-aware optimizer
- `slack-notifier/` - Slack notification command

## Conclusion

PE's plugin system provides a powerful way to extend functionality while maintaining security and performance. Follow the patterns in this guide to create robust, reusable plugins that integrate seamlessly with the PE ecosystem.