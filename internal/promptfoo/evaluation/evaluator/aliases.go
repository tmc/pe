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
	"contains":          AssertionContains,
	"not-contains":      AssertionNotContains,
	"equals":            AssertionEquals,
	"matches":           AssertionMatches,
	"length":            AssertionLength,
	"readability":       AssertionReadability,
	"sentiment":         AssertionSentiment,
	"toxicity":          AssertionToxicity,
	"coherence":         AssertionCoherence,
	"factuality":        AssertionFactuality,
	"llm-judge":         AssertionLLMJudge,
	"classify":          AssertionClassify,
	"similarity":        AssertionSimilarity,
	"latency":           AssertionLatency,
	"cost":              AssertionCost,
	"tokens":            AssertionTokens,
	"json":              AssertionJSON,
	"sql":               AssertionSQL,
	"code":              AssertionCode,
	"structure":         AssertionStructure,
	"pass-at-n":         AssertionPassAtN,
	"structured-output": AssertionStructuredOutput,

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
}

// normalizeAssertionType resolves a config assertion id (promptfoo or pe
// spelling, optionally "not-" prefixed) to pe's canonical AssertionType.
//
// negate is true when the id carried a leading "not-" (promptfoo's inverse
// family, e.g. "not-regex", "not-llm-rubric"); the caller inverts pass/score.
// ok is false when no pe evaluator backs the id, so the caller can fall back
// to the string-match path or surface an unsupported-assertion warning.
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
