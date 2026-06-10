package evaluator

import "strings"

// assertionAlias maps a promptfoo assertion "type" id to pe's canonical
// AssertionType. promptfoo and pe chose different ids for the same checks
// (promptfoo "regex" vs pe "matches", "llm-rubric" vs "llm-judge", "similar"
// vs "similarity", "is-json" vs "json"). The map lets an unmodified promptfoo
// config validate and run against pe's evaluator.
//
// Only ids that resolve to a real pe evaluator are listed; ids with no pe
// implementation are deliberately absent so normalizeAssertionType reports
// ok=false and the caller can fall back or warn.
var assertionAlias = map[string]AssertionType{
	// Canonical ids map to themselves so normalize is the single entry point.
	"contains":             AssertionContains,
	"not-contains":         AssertionNotContains,
	"contains-any":         AssertionContainsAny,
	"contains-all":         AssertionContainsAll,
	"icontains":            AssertionIContains,
	"icontains-any":        AssertionIContainsAny,
	"icontains-all":        AssertionIContainsAll,
	"equals":               AssertionEquals,
	"matches":              AssertionMatches,
	"length":               AssertionLength,
	"readability":          AssertionReadability,
	"sentiment":            AssertionSentiment,
	"toxicity":             AssertionToxicity,
	"coherence":            AssertionCoherence,
	"factuality":           AssertionFactuality,
	"llm-judge":            AssertionLLMJudge,
	"classify":             AssertionClassify,
	"similarity":           AssertionSimilarity,
	"g-eval":               AssertionGEval,
	"answer-relevance":     AssertionAnswerRelevance,
	"context-faithfulness": AssertionContextFaithfulness,
	"context-recall":       AssertionContextRecall,
	"context-relevance":    AssertionContextRelevance,

	// Rubric-based model-graded assertions (agent-rubric/search-rubric run as a
	// plain rubric in pe; the agentic/web-search grader requirement is not
	// enforced).
	"agent-rubric":           AssertionAgentRubric,
	"search-rubric":          AssertionSearchRubric,
	"conversation-relevance": AssertionConversationRelevance,
	"latency":                AssertionLatency,
	"cost":                   AssertionCost,
	"tokens":                 AssertionTokens,
	"json":                   AssertionJSON,
	"sql":                    AssertionSQL,
	"code":                   AssertionCode,
	"structure":              AssertionStructure,
	"pass-at-n":              AssertionPassAtN,
	"structured-output":      AssertionStructuredOutput,

	// deterministic text-metric and structural assertions.
	"levenshtein":                   AssertionLevenshtein,
	"rouge-n":                       AssertionRougeN,
	"bleu":                          AssertionBLEU,
	"gleu":                          AssertionGLEU,
	"word-count":                    AssertionWordCount,
	"starts-with":                   AssertionStartsWith,
	"is-refusal":                    AssertionIsRefusal,
	"is-html":                       AssertionIsHTML,
	"contains-html":                 AssertionContainsHTML,
	"is-xml":                        AssertionIsXML,
	"contains-xml":                  AssertionContainsXML,
	"contains-sql":                  AssertionContainsSQL,
	"is-valid-openai-function-call": AssertionIsValidFunctionCall,
	"is-valid-openai-tools-call":    AssertionIsValidToolsCall,
	// promptfoo also accepts the non-"openai" spelling of the function-call check.
	"is-valid-function-call": AssertionIsValidFunctionCall,

	// promptfoo ids that pe spells differently.
	"regex":         AssertionMatches,
	"llm-rubric":    AssertionLLMJudge,
	"similar":       AssertionSimilarity,
	"classifier":    AssertionClassify,
	"is-json":       AssertionJSON,
	"is-sql":        AssertionSQL,
	"contains-json": AssertionJSON,

	// promptfoo similar:<metric> variants share pe's similarity evaluator;
	// the metric refinement is not yet honored but the assertion still runs.
	"similar:cosine":    AssertionSimilarity,
	"similar:dot":       AssertionSimilarity,
	"similar:euclidean": AssertionSimilarity,

	// model-graded factuality aliases.
	"model-graded-factuality": AssertionFactuality,
	"model-graded-closedqa":   AssertionLLMJudge,

	// Deliberately absent (no pe evaluator yet), so normalizeAssertionType
	// reports ok=false rather than silently passing:
	//   - "perplexity" / "perplexity-score": need token logprobs, which pe's
	//     ProviderResponse does not surface, so they cannot be computed
	//     faithfully and are deferred.
	//   - "finish-reason": pe's ProviderResponse carries no finish_reason; this
	//     must be plumbed through the provider layer first.
	//   - "meteor": promptfoo relies on WordNet synonym matching (the natural
	//     npm package); a faithful port needs the WordNet data set, which would
	//     violate pe's no-large-data-dependency policy, so it is deferred.
	//   - "rouge-l" / "rouge-s": LCS- and skip-bigram-based ROUGE variants that
	//     pe does not implement (only rouge-n is supported).
	//   - code execution ("javascript", "python", "ruby", "webhook"): blocked by
	//     pe's no-external-execution policy.
	//   - external scoring services ("moderation", "guardrails", "pi"): call
	//     hosted APIs (OpenAI/Azure moderation, AWS/Azure guardrail metadata,
	//     the Pi Labs scorer); pe has no client for these yet.
	//   - multi-output comparison ("select-best", "max-score"): rank several
	//     candidate outputs across a test row, which the single-output
	//     EvaluateAssertion signature cannot express; they need a comparison
	//     layer (assert-set) that pe does not have.
}

// normalizeAssertionType resolves a config assertion id (promptfoo or pe
// spelling, optionally "not-" prefixed) to pe's canonical AssertionType.
//
// negate is true when the id carried a leading "not-" (promptfoo's inverse
// family, e.g. "not-regex", "not-llm-rubric"); the caller inverts pass/score.
// ok is false when no pe evaluator backs the id, so the caller can fall back
// to the string-match path or surface an unsupported-assertion warning.
// isModelGraded reports whether an assertion type is evaluated by a model
// judge and therefore requires a judge provider. Such assertions error when no
// provider is available rather than degrading to a string-match fallback.
func isModelGraded(t AssertionType) bool {
	switch t {
	case AssertionLLMJudge, AssertionGEval, AssertionAnswerRelevance,
		AssertionContextFaithfulness, AssertionContextRecall, AssertionContextRelevance,
		AssertionAgentRubric, AssertionSearchRubric, AssertionConversationRelevance:
		return true
	default:
		return false
	}
}

func normalizeAssertionType(id string) (canonical AssertionType, negate bool, ok bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	// promptfoo accepts both "llm-rubric" and "llm_rubric"; normalize the
	// separator so either spelling resolves.
	id = strings.ReplaceAll(id, "_", "-")
	if rest, found := strings.CutPrefix(id, "not-"); found {
		negate = true
		id = rest
	}
	canonical, ok = assertionAlias[id]
	return canonical, negate, ok
}
