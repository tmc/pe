# How Russ Cox Would Build PE

## Core Principles

### 1. **Simplicity Over Structure**
```go
// BAD: Java-style over-organization
cmd/pe/commands/run/run.go
cmd/pe/commands/prompt/prompt.go
cmd/pe/commands/module/module.go

// GOOD: Everything in one package
cmd/pe/run.go
cmd/pe/prompt.go
cmd/pe/mod.go
```

### 2. **Delete Code Aggressively**
- ExecutionLog with 100+ lines? DELETE
- SHA256 hashing everything? DELETE  
- Complex logging system? DELETE - use shell redirection
- 46 commands when 5 would do? DELETE

### 3. **Unix Philosophy**
```bash
# Don't build logging into the tool
pe run prompt.txt > output.log

# Don't build complex filtering
pe run prompt.txt | grep "error"

# Don't build retry logic
while ! pe run prompt.txt; do sleep 1; done
```

### 4. **Focus on Core Use Cases**
```go
// What PE actually needs:
pe run "prompt"           // Run a prompt
pe run file.prompt        // Run from file
cat data | pe run -       // Unix pipes
pe mod init               // Module management
pe test                   // Test prompts

// What PE doesn't need:
pe optimize               // Academic exercise
pe semantic               // Research project
pe evolve                 // Nobody uses this
pe consensus-demo         // Demo code
pe playground             // Not core
```

### 5. **No Premature Abstraction**
```go
// BAD: Factory patterns everywhere
func NewCommand() CommandInterface {
    return &commandImpl{
        factory: NewFactory(),
        logger: NewLogger(),
    }
}

// GOOD: Just a function
func runCmd() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // Direct implementation
        },
    }
}
```

### 6. **Clear is Better Than Clever**
```go
// BAD: Clever pipe syntax parsing
{{.input | run math-solver | run formatter}}

// GOOD: Simple and clear
result := solveMath(input)
formatted := format(result)
```

## What Russ Would Actually Do

### Step 1: Delete 80% of the code
```bash
# Delete rarely-used commands
rm cmd/pe/evolve.go
rm cmd/pe/semantic.go
rm cmd/pe/optimize.go
rm cmd/pe/playground.go
rm cmd/pe/consensus_demo.go
# ... etc
```

### Step 2: Consolidate shared code in internal/
```go
// internal/prompt/metadata.go - GOOD, you already did this
func ExtractDefaults(content string) map[string]string
func ExtractVariables(content string) []string
```

### Step 3: Make pe run the primary interface
```bash
# Most users just want to run prompts
pe "What is 2+2?"                    # Direct
pe run file.prompt                   # From file
pe run file.prompt --var task="foo"  # With variables
```

### Step 4: Keep module system dead simple
```bash
# No JSON files, no complex registry
echo "github.com/user/prompts" > pe.mod
pe mod get
```

### Step 5: Make it fast
- No unnecessary file I/O
- No complex initialization
- No provider detection logic (just use cgpt)
- Binary under 10MB

## Real Example: How Go's toolchain works

Look at how `go run` is implemented:
```go
// It's not in cmd/go/commands/run/run.go
// It's just in cmd/go/internal/run/run.go
// One package, clear purpose, no abstractions
```

## The Test

Ask yourself:
1. Can I explain what every line does?
2. Could a new programmer understand this in 10 minutes?
3. Is the code smaller than the problem it solves?
4. Does it compose well with Unix tools?

If any answer is "no", you're over-engineering.

## Quote from Russ Cox

> "The most important property of a program is whether it accomplishes the intention of its user."

Not whether it has perfect architecture, not whether it follows patterns, but whether it WORKS and is SIMPLE.