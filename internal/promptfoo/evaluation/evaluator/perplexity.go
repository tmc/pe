package evaluator

import (
	"fmt"
	"math"
)

// perplexity and perplexity-score read the response's per-token log
// probabilities from metadata["logprobs"] — a list of natural-log token
// probabilities the provider returned. pe has no provider that surfaces these
// yet, so in practice these fail with a clear message; when a provider does
// populate metadata["logprobs"], the assertions compute faithfully.
//
//	perplexity       = exp(-mean(logprob)); pass when <= threshold (lower is
//	                   more confident). Default threshold: no upper bound (always
//	                   reports, used for visibility) unless max/threshold set.
//	perplexity-score = 1 / (1 + perplexity), a 0..1 score where higher is more
//	                   confident; pass when >= threshold (default 0.5).

// logprobsFromMeta extracts the per-token log probabilities from metadata.
func logprobsFromMeta(metadata map[string]interface{}) ([]float64, bool) {
	if metadata == nil {
		return nil, false
	}
	raw, ok := metadata["logprobs"]
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case []float64:
		return v, len(v) > 0
	case []interface{}:
		out := make([]float64, 0, len(v))
		for _, e := range v {
			if f, ok := asFloat(e); ok {
				out = append(out, f)
			}
		}
		return out, len(out) > 0
	}
	return nil, false
}

// perplexity returns exp(-mean(logprobs)).
func perplexity(logprobs []float64) float64 {
	if len(logprobs) == 0 {
		return math.Inf(1)
	}
	sum := 0.0
	for _, lp := range logprobs {
		sum += lp
	}
	mean := sum / float64(len(logprobs))
	return math.Exp(-mean)
}

// evaluatePerplexity computes perplexity from the response logprobs and passes
// when it is within the max bound (assertion.Max or assertion.Threshold). With
// no bound it reports the value and passes (visibility-only, matching how
// latency behaves without a max).
func (ae *AssertionEvaluator) evaluatePerplexity(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	logprobs, ok := logprobsFromMeta(metadata)
	if !ok {
		return failResult(assertion, "perplexity requires per-token logprobs in the response (metadata.logprobs); the provider did not supply them")
	}
	pp := perplexity(logprobs)
	var bound *float64
	if assertion.Max != nil {
		bound = assertion.Max
	} else if assertion.Threshold != nil {
		bound = assertion.Threshold
	}
	passed := true
	message := fmt.Sprintf("perplexity %.3f", pp)
	if bound != nil {
		passed = pp <= *bound
		message = fmt.Sprintf("perplexity %.3f (expected <= %.3f)", pp, *bound)
	}
	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    boolScore(passed),
		Actual:   pp,
		Message:  message,
		Metadata: map[string]interface{}{"method": "perplexity", "perplexity": pp, "tokens": len(logprobs)},
	}
}

// evaluatePerplexityScore computes the normalized confidence score
// 1/(1+perplexity) and passes when it is >= threshold (default 0.5).
func (ae *AssertionEvaluator) evaluatePerplexityScore(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	logprobs, ok := logprobsFromMeta(metadata)
	if !ok {
		return failResult(assertion, "perplexity-score requires per-token logprobs in the response (metadata.logprobs); the provider did not supply them")
	}
	pp := perplexity(logprobs)
	score := 1.0 / (1.0 + pp)
	threshold := 0.5
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("perplexity-score %.3f (perplexity %.3f, threshold %.2f)", score, pp, threshold), map[string]interface{}{"method": "perplexity_score", "perplexity": pp})
}
