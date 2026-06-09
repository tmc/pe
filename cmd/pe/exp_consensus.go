package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/distributed"
)

type expConsensusInput struct {
	Votes []expConsensusVote `json:"votes"`
}

type expConsensusVote struct {
	Provider string `json:"provider,omitempty"`
	Output   string `json:"output,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Error    string `json:"error,omitempty"`
}

type expConsensusOutput struct {
	Consensus distributed.Consensus   `json:"consensus"`
	Errors    []expConsensusError     `json:"errors,omitempty"`
	Votes     []distributed.Vote      `json:"votes,omitempty"`
	Rows      []expConsensusOutputRow `json:"rows,omitempty"`
}

type expConsensusError struct {
	Provider string `json:"provider,omitempty"`
	Error    string `json:"error"`
}

type expConsensusOutputRow struct {
	Provider string `json:"provider,omitempty"`
	Output   string `json:"output,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Error    string `json:"error,omitempty"`
	Used     bool   `json:"used"`
}

func init() {
	expCmd.AddCommand(expConsensusCmd())
}

func expConsensusCmd() *cobra.Command {
	var inputPath string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "consensus [--input votes.json]",
		Short: "Aggregate local provider votes deterministically",
		Long: `Aggregate local provider votes deterministically.

Input JSON contains a votes array. Rows with an error are reported and excluded;
successful rows are passed to the local weighted majority aggregator.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readExpConsensusInput(cmd.InOrStdin(), inputPath)
			if err != nil {
				return err
			}
			out, err := runExpConsensus(in)
			if err != nil {
				return err
			}
			return writeExpConsensusOutput(cmd.OutOrStdout(), outputPath, out)
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "input JSON file, or - for stdin")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "-", "output JSON file, or - for stdout")
	return cmd
}

func readExpConsensusInput(stdin io.Reader, path string) (*expConsensusInput, error) {
	var r io.Reader = stdin
	if path != "" && path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open consensus input: %w", err)
		}
		defer f.Close()
		r = f
	}

	var in expConsensusInput
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return nil, fmt.Errorf("decode consensus input: %w", err)
	}
	if len(in.Votes) == 0 {
		return nil, fmt.Errorf("consensus input requires votes")
	}
	return &in, nil
}

func runExpConsensus(in *expConsensusInput) (*expConsensusOutput, error) {
	if in == nil {
		return nil, fmt.Errorf("consensus input is nil")
	}
	votes := make([]distributed.Vote, 0, len(in.Votes))
	errors := make([]expConsensusError, 0)
	rows := make([]expConsensusOutputRow, 0, len(in.Votes))
	for _, row := range in.Votes {
		used := row.Error == ""
		if row.Error != "" {
			errors = append(errors, expConsensusError{
				Provider: row.Provider,
				Error:    row.Error,
			})
		} else {
			votes = append(votes, distributed.Vote{
				Provider: row.Provider,
				Output:   row.Output,
				Weight:   row.Weight,
			})
		}
		rows = append(rows, expConsensusOutputRow{
			Provider: row.Provider,
			Output:   row.Output,
			Weight:   row.Weight,
			Error:    row.Error,
			Used:     used,
		})
	}

	consensus, err := distributed.Majority(votes)
	if err != nil {
		return nil, err
	}
	return &expConsensusOutput{
		Consensus: consensus,
		Errors:    errors,
		Votes:     votes,
		Rows:      rows,
	}, nil
}

func writeExpConsensusOutput(stdout io.Writer, path string, out *expConsensusOutput) error {
	var w io.Writer = stdout
	if path != "" && path != "-" {
		if err := enforceRuntimeToolPolicy("write"); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create consensus output: %w", err)
		}
		defer f.Close()
		w = f
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode consensus output: %w", err)
	}
	return nil
}
