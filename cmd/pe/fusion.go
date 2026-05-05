package main

import "github.com/spf13/cobra"

func fusionCmd() *cobra.Command {
	cmd := expConsensusCmd()
	cmd.Use = "fusion [--input votes.json]"
	cmd.Short = "Aggregate provider outputs with deterministic consensus"
	cmd.Long = `Aggregate provider outputs with deterministic consensus.

Input JSON contains a votes array. Rows with an error are reported and excluded;
successful rows are passed to the local weighted majority aggregator. This is
the stable top-level form of pe exp consensus.`
	return cmd
}
