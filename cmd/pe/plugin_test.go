package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginBuildCmd_DeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "example-plugin")
	if err := os.Mkdir(pluginDir, 0755); err != nil {
		t.Fatalf("creating plugin dir: %v", err)
	}
	oldWd, _ := os.Getwd()
	os.Chdir(pluginDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny write
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}

	var out bytes.Buffer
	cmd := pluginBuildCmd()
	cmd.SetOut(&out)
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("plugin build succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("plugin build error = %v, want write policy error", err)
	}
	if _, err := os.Stat("example-plugin.so"); !os.IsNotExist(err) {
		t.Fatalf("plugin artifact stat error = %v, want not exist", err)
	}
}

func TestPluginRunCmd_DeniedByExecPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny exec
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}

	var out bytes.Buffer
	cmd := pluginRunCmd()
	cmd.SetOut(&out)
	err := cmd.RunE(cmd, []string{"example"})
	if err == nil {
		t.Fatal("plugin run succeeded, want exec policy error")
	}
	if !strings.Contains(err.Error(), "tool exec is denied by pe.mod") {
		t.Fatalf("plugin run error = %v, want exec policy error", err)
	}
}
