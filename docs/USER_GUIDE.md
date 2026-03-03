# PE User Guide

Welcome to the **Prompt Engineering (PE)** toolkit, a unified toolchain for LLM development that brings Go's philosophy of simplicity, composability, and performance to prompt engineering.

### LLM Providers
PE supports various local and remote LLM providers:

#### Built-in CLI Presets
Run local models easily with built-in presets:

- **Ollama**: `pe run --provider ollama:llama3 "Hello"`
- **MLX-LM**: `pe run --provider mlx-lm:mistral "Hello"` (alias: `mlx`)
- **MLX-Go**: `pe run --provider mlx-go:mistral "Hello"`
- **Llama.cpp**: `pe run --provider llama-cpp:./models/7b.gguf "Hello"`
- **Generic CLI**: Use any CLI tool via config.

#### Cloud Providers
- **OpenAI**: `pe run --provider openai:gpt-4 "Hello"`
- **Anthropic**: `pe run --provider anthropic:claude-3-opus "Hello"`

### Configuration
You can define custom CLI providers in `pe-config.yaml`:

```yaml
providers:
  - id: my-local-model
    type: cli
    config:
      command: "python my_script.py --prompt {{.Prompt}}"
```

## 1. Philosophy: "Go, for Prompts"

PE is designed to be for prompts what the `go` toolchain is for Go code. It unifies scattered utilities into a single, cohesive binary with standard commands:

- `pe init` ≈ `go mod init`
- `pe run` ≈ `go run`
- `pe test` ≈ `go test`
- `pe mod` ≈ `go mod`
- `pe fmt` ≈ `go fmt`

The toolkit emphasizes **Unix composability** (pipes/streams) and **production readiness** (testing, security, optimization).

## 2. Getting Started

### Installation
```bash
go install github.com/tmc/pe/cmd/pe@latest
```

### Initializing a Project
Initialize a new PE repository with a standard structure:
```bash
mkdir my-prompts
cd my-prompts
pe init
```
This creates a `.pe` directory for configuration and modules.

## 3. Core Workflows

### Evaluating Prompts (`pe eval`)
The core loop of PE is evaluation. Define your test cases and prompts in a config file (compatible with promptfoo).

**Example `config.yaml`:**
```yaml
prompts: [prompts/*.txt]
providers: [openai:gpt-4]
tests:
  - vars:
      topic: "Go programming"
    assert:
      - type: contains
        value: "language"
```

Run the evaluation:
```bash
pe eval config.yaml --save-db
```

### Viewing Results (`pe view`)
Visualize your evaluation results in a browser-based UI:
```bash
pe view           # View most recent evaluation
pe view <eval-id> # View specific evaluation
```

### Interactive Development (`pe interactive`)
Start a REPL to iterate on prompts with real-time feedback:
```bash
pe interactive --provider openai:gpt-4
```

## 4. Pipeline & Streaming

PE shines in Unix pipelines, allowing you to compose complex LLM workflows.

**Ask a question:**
```bash
echo "Explain quantum computing" | pe ask --provider openai:gpt-4
```

**Analyze results stream:**
```bash
pe eval config.yaml | pe filter --success | pe analyze --metric latency
```

**Aggregate data:**
```bash
pe eval config.yaml | pe reduce --operation mean
```

## 5. Module Management

Manage prompt dependencies just like Go modules.

- **Initialize**: `pe mod init`
- **Add Dependency**: `pe mod get github.com/user/repo`
- **Tidy Dependencies**: `pe mod tidy`
- **Vendor Dependencies**: `pe mod vendor`

## 6. Security & Attestation

Ensure your prompts are secure and trusted.

**Security Scan (OWASP LLM Top 10):**
```bash
pe security scan --prompt task.txt
```

**Cryptographic Attestation:**
Prototype command group (inspect current interface):
```bash
pe exp attest --help
```

## 7. Optimization

Automatically optimize your prompts using advanced techniques (like PE2, APEX):

```bash
pe optimize --prompt prompt.txt --method apex
```

## 8. Command Reference Summary

| Command | Description |
|---------|-------------|
| `eval` | Evaluate prompt configurations |
| `view` | View results in browser |
| `run` | Execute a prompt immediately |
| `test` | Run advanced testing (property-based, regression) |
| `fmt` | Format prompt files |
| `ask` | Execute prompts (pipeline-friendly) |
| `stream` | Process results as a stream |
| `mod` | Manage modules |
| `security` | Run security scans |
| `profile` | Performance profiling |

For a complete reference, run `pe help` or see `CLI_REFERENCE.md`.
