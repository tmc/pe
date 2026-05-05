package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestFusionCmdRunsConsensus(t *testing.T) {
	cmd := fusionCmd()
	cmd.SetIn(strings.NewReader(`{"votes":[{"provider":"a","output":"yes"},{"provider":"b","output":"yes"},{"provider":"c","output":"no"}]}`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\n%s", err, out.String())
	}
	for _, want := range []string{`"output": "yes"`, `"provider": "a"`} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, out.String())
		}
	}
}
