package rlm

import (
	"github.com/tmc/pe/internal/distributed"
)

// Aggregation rule names. Majority is the initial deterministic default.
const (
	RuleMajority = "majority"
	RuleConcat   = "concat"
)

// aggregate merges child results into a parent answer using the named rule.
//
// majority resolves competing child answers with deterministic weighted voting
// (see distributed.Majority). concat joins child outputs in order. The returned
// votes record every non-error child answer considered.
func aggregate(rule string, results []mapResult) (Aggregation, error) {
	votes := make([]Vote, 0, len(results))
	dvotes := make([]distributed.Vote, 0, len(results))
	for _, r := range results {
		if r.err != nil || r.output == "" {
			continue
		}
		votes = append(votes, Vote{
			Output: r.output,
			Weight: 1,
		})
		dvotes = append(dvotes, distributed.Vote{
			Provider: "",
			Output:   r.output,
			Weight:   1,
		})
	}

	switch rule {
	case RuleConcat:
		outputs := make([]string, 0, len(results))
		for _, r := range results {
			if r.err == nil {
				outputs = append(outputs, r.output)
			}
		}
		return Aggregation{
			Rule:      RuleConcat,
			Consensus: reduce(outputs, nil),
			Votes:     votes,
		}, nil
	case RuleMajority, "":
		consensus, err := distributed.Majority(dvotes)
		if err != nil {
			return Aggregation{}, err
		}
		return Aggregation{
			Rule:      RuleMajority,
			Consensus: consensus.Output,
			Votes:     votes,
		}, nil
	default:
		return Aggregation{}, &unsupportedRuleError{rule: rule}
	}
}

type unsupportedRuleError struct{ rule string }

func (e *unsupportedRuleError) Error() string {
	return "unsupported aggregation rule: " + e.rule
}
