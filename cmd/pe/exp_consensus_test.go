package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestExpConsensusTieUsesFirstVote(t *testing.T) {
	out := runExpConsensusCommand(t, `{
  "votes": [
    {"provider":"a", "output":"alpha", "weight":2},
    {"provider":"b", "output":"beta", "weight":2}
  ]
}`)
	if out.Consensus.Output != "alpha" {
		t.Fatalf("output = %q, want alpha", out.Consensus.Output)
	}
	if out.Consensus.Weight != 2 {
		t.Fatalf("weight = %d, want 2", out.Consensus.Weight)
	}
	if len(out.Errors) != 0 {
		t.Fatalf("errors = %+v, want none", out.Errors)
	}
}

func TestExpConsensusRejectsEmptyVotes(t *testing.T) {
	cmd := expConsensusCmd()
	cmd.SetIn(strings.NewReader(`{"votes":[]}`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute succeeded, want empty votes error")
	}
	if !strings.Contains(err.Error(), "consensus input requires votes") {
		t.Fatalf("error = %v, want empty votes error", err)
	}
}

func TestExpConsensusWeightedVotes(t *testing.T) {
	out := runExpConsensusCommand(t, `{
  "votes": [
    {"provider":"a", "output":"north", "weight":1},
    {"provider":"b", "output":"south", "weight":3},
    {"provider":"c", "output":"north", "weight":1}
  ]
}`)
	if out.Consensus.Output != "south" {
		t.Fatalf("output = %q, want south", out.Consensus.Output)
	}
	if out.Consensus.Weight != 3 {
		t.Fatalf("weight = %d, want 3", out.Consensus.Weight)
	}
	if out.Consensus.Total != 5 {
		t.Fatalf("total = %d, want 5", out.Consensus.Total)
	}
}

func TestExpConsensusProviderErrorRows(t *testing.T) {
	out := runExpConsensusCommand(t, `{
  "votes": [
    {"provider":"ok-a", "output":"yes", "weight":1},
    {"provider":"bad", "error":"provider timeout", "weight":100},
    {"provider":"ok-b", "output":"yes", "weight":2}
  ]
}`)
	if out.Consensus.Output != "yes" {
		t.Fatalf("output = %q, want yes", out.Consensus.Output)
	}
	if out.Consensus.Weight != 3 {
		t.Fatalf("weight = %d, want 3", out.Consensus.Weight)
	}
	if out.Consensus.Total != 3 {
		t.Fatalf("total = %d, want 3", out.Consensus.Total)
	}
	if len(out.Errors) != 1 || out.Errors[0].Provider != "bad" || out.Errors[0].Error != "provider timeout" {
		t.Fatalf("errors = %+v, want bad provider timeout", out.Errors)
	}
	for _, row := range out.Rows {
		if row.Provider == "bad" && row.Used {
			t.Fatalf("provider error row marked used: %+v", row)
		}
	}
}

func TestExpConsensusRejectsOnlyProviderErrors(t *testing.T) {
	cmd := expConsensusCmd()
	cmd.SetIn(strings.NewReader(`{
  "votes": [
    {"provider":"bad", "error":"provider timeout"}
  ]
}`))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute succeeded, want no votes error")
	}
	if !strings.Contains(err.Error(), "no votes") {
		t.Fatalf("error = %v, want no votes", err)
	}
}

func runExpConsensusCommand(t *testing.T, input string) expConsensusOutput {
	t.Helper()

	cmd := expConsensusCmd()
	cmd.SetIn(strings.NewReader(input))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute consensus: %v\n%s", err, out.String())
	}

	var got expConsensusOutput
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\n%s", err, out.String())
	}
	return got
}
