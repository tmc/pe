package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

// peCmd implements a script command that runs pe binary with restrictions
func peCmd(peBinary string) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "run pe command",
			Args:    "args...",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			cmd := exec.Command(peBinary, args...)
			cmd.Dir = s.Getwd()
			cmd.Env = append(os.Environ(), s.Environ()...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Start(); err != nil {
				return nil, err
			}

			return func(*script.State) (string, string, error) {
				err := cmd.Wait()
				return stdout.String(), stderr.String(), err
			}, nil
		},
	)
}

func pipeCmd(peBinary string) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "run a restricted pipeline",
			Args:    "command [args...] | command [args...] ...",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			stages, err := pipeStages(args)
			if err != nil {
				return nil, err
			}
			return func(*script.State) (string, string, error) {
				var in string
				var stderr strings.Builder
				for _, stage := range stages {
					out, errout, err := runPipeStage(s, peBinary, stage, in)
					stderr.WriteString(errout)
					if err != nil {
						return out, stderr.String(), err
					}
					in = out
				}
				return in, stderr.String(), nil
			}, nil
		},
	)
}

func pipeStages(args []string) ([][]string, error) {
	var stages [][]string
	var stage []string
	for _, arg := range args {
		if arg == "|" {
			if len(stage) == 0 {
				return nil, fmt.Errorf("empty pipeline stage")
			}
			stages = append(stages, stage)
			stage = nil
			continue
		}
		stage = append(stage, arg)
	}
	if len(stage) == 0 {
		return nil, fmt.Errorf("empty pipeline stage")
	}
	stages = append(stages, stage)
	if len(stages) < 2 {
		return nil, fmt.Errorf("pipeline needs at least two stages")
	}
	return stages, nil
}

func runPipeStage(s *script.State, peBinary string, stage []string, in string) (string, string, error) {
	switch stage[0] {
	case "echo":
		return strings.Join(stage[1:], " ") + "\n", "", nil
	case "cat":
		var out strings.Builder
		if len(stage) == 1 {
			out.WriteString(in)
		}
		for _, file := range stage[1:] {
			data, err := os.ReadFile(s.Path(file))
			if err != nil {
				return out.String(), "", err
			}
			out.Write(data)
		}
		return out.String(), "", nil
	case "pe":
		cmd := exec.Command(peBinary, stage[1:]...)
		cmd.Dir = s.Getwd()
		cmd.Env = append(os.Environ(), s.Environ()...)
		cmd.Stdin = strings.NewReader(in)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	default:
		return "", "", fmt.Errorf("unknown pipeline command %q", stage[0])
	}
}

func TestPipeStages(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    [][]string
		wantErr bool
	}{
		{
			name: "two stages",
			args: []string{"pe", "run", "prompt.txt", "|", "pe", "experimental", "extract", "--tag", "answer"},
			want: [][]string{{"pe", "run", "prompt.txt"}, {"pe", "experimental", "extract", "--tag", "answer"}},
		},
		{
			name: "three stages",
			args: []string{"cat", "input.txt", "|", "pe", "run", "-", "|", "pe", "experimental", "extract", "--tag", "answer"},
			want: [][]string{{"cat", "input.txt"}, {"pe", "run", "-"}, {"pe", "experimental", "extract", "--tag", "answer"}},
		},
		{
			name:    "one stage",
			args:    []string{"pe", "run", "prompt.txt"},
			wantErr: true,
		},
		{
			name:    "empty middle stage",
			args:    []string{"pe", "run", "prompt.txt", "|", "|", "pe", "run", "-"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pipeStages(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("pipeStages succeeded")
				}
				return
			}
			if err != nil {
				t.Fatalf("pipeStages: %v", err)
			}
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Fatalf("pipeStages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScripts(t *testing.T) {
	binDir := t.TempDir()
	peBinary := filepath.Join(binDir, "pe")
	promptfooPlugin := filepath.Join(binDir, "pe-promptfoo")

	// Build the pe binary first
	cmd := exec.Command("go", "build", "-o", peBinary, "./cmd/pe")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build pe binary: %v", err)
	}

	// Build the promptfoo plugin so `pe promptfoo` resolves during script tests.
	cmd = exec.Command("go", "build", "-o", promptfooPlugin, "./plugins/promptfoo")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build pe-promptfoo plugin: %v", err)
	}

	// Create custom commands - remove exec, add pe
	cmds := make(map[string]script.Cmd)
	for name, cmd := range scripttest.DefaultCmds() {
		// Skip exec to prevent arbitrary shell execution
		if name != "exec" {
			cmds[name] = cmd
		}
	}
	// Add our custom pe command
	cmds["pe"] = peCmd(peBinary)
	cmds["pipe"] = pipeCmd(peBinary)

	// Create engine with custom commands (no exec, only pe)
	engine := &script.Engine{
		Cmds:  cmds,
		Conds: scripttest.DefaultConds(),
	}

	// Set up test environment
	env := []string{
		"PE_TEST_MODE=true",
		"PE_MOCK_PROVIDER=true",
		"PATH=" + binDir + string(os.PathListSeparator) + os.Getenv("PATH"),
	}

	// Run tests from testdata/script directory
	scriptDir := "testdata/script"

	// Check if directory exists
	if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
		t.Skip("testdata/script directory not found")
		return
	}

	// Get all .txt files in the directory
	files, err := filepath.Glob(filepath.Join(scriptDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Skip("no test files found in testdata/script")
		return
	}

	// Run each test file
	for _, file := range files {
		file := file // capture loop variable
		name := strings.TrimSuffix(filepath.Base(file), ".txt")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			scripttest.Test(t, context.Background(), engine, env, file)
		})
	}
}
