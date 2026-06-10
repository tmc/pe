package evaluator

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Deterministic text-metric and structural assertions. These need no judge
// provider and mirror promptfoo's semantics:
//   - levenshtein: pass when edit distance to the reference <= threshold (default 5)
//   - rouge-n:     n-gram overlap F1 (js-rouge default n=1, beta=1), pass when score >= threshold (default 0.75)
//   - bleu:        n-gram precision, pass when score >= threshold (default 0.5)
//   - word-count:  whitespace-split word count, exact value or {min,max}
//   - is-valid-openai-function-call / -tools-call: structural JSON validation

// evaluateLevenshtein passes when the Levenshtein edit distance between the
// output and the reference string (value) is within threshold (default 5).
func (ae *AssertionEvaluator) evaluateLevenshtein(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "levenshtein assertion requires a reference string value")
	}
	distance := levenshtein(output, reference)
	threshold := 5.0
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	passed := float64(distance) <= threshold
	// Score is a normalized closeness in [0,1] for reporting.
	maxLen := max(len(output), len(reference))
	score := 1.0
	if maxLen > 0 {
		score = 1.0 - float64(distance)/float64(maxLen)
	}
	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    score,
		Expected: reference,
		Actual:   distance,
		Message:  fmt.Sprintf("Levenshtein distance %d (threshold %.0f)", distance, threshold),
		Metadata: map[string]interface{}{"distance": distance, "threshold": threshold},
	}
}

// evaluateRougeN computes ROUGE-N overlap against the reference and passes when
// the score is >= threshold (default 0.75). It mirrors promptfoo, which calls
// js-rouge's rouge.n with default options: n=1 (unigram), beta=1 (so the score
// is the F1 of n-gram precision and recall), and caseSensitive=true. n can be
// overridden via the assertion's config.n.
func (ae *AssertionEvaluator) evaluateRougeN(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "rouge-n assertion requires a reference string value")
	}
	n := assertionConfigInt(assertion, "n", 1)
	score := rougeN(output, reference, n)
	threshold := 0.75
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("ROUGE-%d %.3f (threshold %.2f)", n, score, threshold), map[string]interface{}{"method": "rouge_n", "n": n})
}

// evaluateRougeL computes ROUGE-L (longest-common-subsequence) overlap against
// the reference and passes when the score is >= threshold (default 0.75). Like
// rouge-n it mirrors js-rouge: precision = LCS/len(candidate),
// recall = LCS/len(reference), score = F1 (beta=1), case-sensitive.
func (ae *AssertionEvaluator) evaluateRougeL(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "rouge-l assertion requires a reference string value")
	}
	score := rougeL(output, reference)
	threshold := 0.75
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("ROUGE-L %.3f (threshold %.2f)", score, threshold), map[string]interface{}{"method": "rouge_l"})
}

// evaluateRougeS computes ROUGE-S (skip-bigram) overlap against the reference
// and passes when the score is >= threshold (default 0.75). It mirrors
// js-rouge: skip-bigrams are all ordered token pairs (i<j) within each text,
// the score is the F1 (beta=1) of the distinct skip-bigram set intersection,
// case-sensitive.
func (ae *AssertionEvaluator) evaluateRougeS(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "rouge-s assertion requires a reference string value")
	}
	score := rougeS(output, reference)
	threshold := 0.75
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("ROUGE-S %.3f (threshold %.2f)", score, threshold), map[string]interface{}{"method": "rouge_s"})
}

// evaluateBLEU computes n-gram precision (default up to 4-grams, geometric
// mean) against the reference and passes when the score is >= threshold
// (default 0.5).
func (ae *AssertionEvaluator) evaluateBLEU(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "bleu assertion requires a reference string value")
	}
	score := bleu(output, reference, 4)
	threshold := 0.5
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("BLEU %.3f (threshold %.2f)", score, threshold), map[string]interface{}{"method": "bleu"})
}

// evaluateGLEU computes Google-BLEU (GLEU) against the reference and passes when
// the score is >= threshold (default 0.5). It mirrors promptfoo's custom gleu:
// over n-grams for n=minN..maxN (default 1..4) pooled together, GLEU is the
// minimum of precision and recall, where precision = matches / total candidate
// n-grams and recall = matches / total reference n-grams. minN/maxN can be
// overridden via the assertion's config.
func (ae *AssertionEvaluator) evaluateGLEU(assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "gleu assertion requires a reference string value")
	}
	minN := assertionConfigInt(assertion, "minN", 1)
	maxN := assertionConfigInt(assertion, "maxN", 4)
	score := gleu(output, reference, minN, maxN)
	threshold := 0.5
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, score, threshold, fmt.Sprintf("GLEU %.3f (threshold %.2f)", score, threshold), map[string]interface{}{"method": "gleu", "minN": minN, "maxN": maxN})
}

// evaluateWordCount checks the output's whitespace-delimited word count against
// min/max bounds or an exact value, mirroring promptfoo's word-count assertion.
func (ae *AssertionEvaluator) evaluateWordCount(assertion Assertion, output string) *AssertionResult {
	count := len(strings.Fields(output))
	var passed bool
	var message string
	switch {
	case assertion.Min != nil && assertion.Max != nil:
		passed = count >= int(*assertion.Min) && count <= int(*assertion.Max)
		message = fmt.Sprintf("Word count %d (expected between %d and %d)", count, int(*assertion.Min), int(*assertion.Max))
	case assertion.Min != nil:
		passed = count >= int(*assertion.Min)
		message = fmt.Sprintf("Word count %d (expected at least %d)", count, int(*assertion.Min))
	case assertion.Max != nil:
		passed = count <= int(*assertion.Max)
		message = fmt.Sprintf("Word count %d (expected at most %d)", count, int(*assertion.Max))
	default:
		exact, ok := assertionInt(assertion.Value)
		if !ok {
			return failResult(assertion, "word-count assertion requires min/max bounds or an exact integer value")
		}
		passed = count == exact
		message = fmt.Sprintf("Word count %d (expected %d)", count, exact)
	}
	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    boolScore(passed),
		Actual:   count,
		Message:  message,
		Metadata: map[string]interface{}{"count": count},
	}
}

// evaluateStartsWith passes when the output begins with the reference string.
// It is case-sensitive, mirroring promptfoo's String.prototype.startsWith check.
func (ae *AssertionEvaluator) evaluateStartsWith(assertion Assertion, output string) *AssertionResult {
	prefix, ok := assertion.Value.(string)
	if !ok {
		return failResult(assertion, "starts-with assertion requires a string value")
	}
	passed := strings.HasPrefix(output, prefix)
	return boolResult(assertion, passed, fmt.Sprintf("output %s with %q", passOrNot(passed, "starts", "does not start"), prefix))
}

// passOrNot returns yes when b is true, otherwise no; a tiny helper for building
// human-readable assertion messages.
func passOrNot(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

// evaluateIsValidFunctionCall validates that the output (or its function_call
// field) is shaped {name: string, arguments: <stringified JSON>}.
func (ae *AssertionEvaluator) evaluateIsValidFunctionCall(assertion Assertion, output string) *AssertionResult {
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(output), &root); err != nil {
		return boolResult(assertion, false, "output is not valid JSON")
	}
	call := root
	if fc, ok := root["function_call"].(map[string]interface{}); ok {
		call = fc
	}
	if msg, ok := validateFunctionCall(call); !ok {
		return boolResult(assertion, false, msg)
	}
	return boolResult(assertion, true, "valid OpenAI function call")
}

// evaluateIsValidToolsCall validates that the output is an array (or has a
// tool_calls array) of {type:"function", function:{name, arguments}} objects.
func (ae *AssertionEvaluator) evaluateIsValidToolsCall(assertion Assertion, output string) *AssertionResult {
	var raw interface{}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return boolResult(assertion, false, "output is not valid JSON")
	}
	var calls []interface{}
	switch v := raw.(type) {
	case []interface{}:
		calls = v
	case map[string]interface{}:
		tc, ok := v["tool_calls"].([]interface{})
		if !ok {
			return boolResult(assertion, false, "object has no tool_calls array")
		}
		calls = tc
	default:
		return boolResult(assertion, false, "output is not a tool_calls array or object")
	}
	if len(calls) == 0 {
		return boolResult(assertion, false, "tool_calls is empty")
	}
	for i, c := range calls {
		entry, ok := c.(map[string]interface{})
		if !ok {
			return boolResult(assertion, false, fmt.Sprintf("tool_calls[%d] is not an object", i))
		}
		if entry["type"] != "function" {
			return boolResult(assertion, false, fmt.Sprintf("tool_calls[%d].type is not \"function\"", i))
		}
		fn, ok := entry["function"].(map[string]interface{})
		if !ok {
			return boolResult(assertion, false, fmt.Sprintf("tool_calls[%d].function missing", i))
		}
		if msg, ok := validateFunctionCall(fn); !ok {
			return boolResult(assertion, false, fmt.Sprintf("tool_calls[%d].function %s", i, msg))
		}
	}
	return boolResult(assertion, true, fmt.Sprintf("valid OpenAI tools call (%d)", len(calls)))
}

// validateFunctionCall checks the {name: string, arguments: stringified JSON}
// shape shared by function_call and tool_calls[].function.
func validateFunctionCall(call map[string]interface{}) (string, bool) {
	name, ok := call["name"].(string)
	if !ok || name == "" {
		return "missing string name", false
	}
	args, ok := call["arguments"].(string)
	if !ok {
		return "arguments must be a JSON string", false
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(args), &parsed); err != nil {
		return "arguments is not valid JSON", false
	}
	return "", true
}

// --- shared helpers ---

func failResult(assertion Assertion, msg string) *AssertionResult {
	return &AssertionResult{Type: assertion.Type, Passed: false, Score: 0, Message: msg}
}

func boolResult(assertion Assertion, passed bool, msg string) *AssertionResult {
	return &AssertionResult{Type: assertion.Type, Passed: passed, Score: boolScore(passed), Message: msg}
}

func scoreResult(assertion Assertion, score, threshold float64, msg string, meta map[string]interface{}) *AssertionResult {
	meta["threshold"] = threshold
	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   score >= threshold,
		Score:    score,
		Message:  msg,
		Metadata: meta,
	}
}

func boolScore(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// assertionInt coerces an assertion value (from JSON/YAML, so typically
// float64) into an int, reporting whether it held a usable integer.
func assertionInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

// assertionConfigInt reads an integer option from assertion.Config[key],
// falling back to def when absent or not numeric.
func assertionConfigInt(assertion Assertion, key string, def int) int {
	if assertion.Config == nil {
		return def
	}
	if n, ok := assertionInt(assertion.Config[key]); ok {
		return n
	}
	return def
}

// levenshtein returns the edit distance between a and b using the standard
// two-row dynamic-programming algorithm over runes.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(min(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

// ngrams returns the list of n-grams (as space-joined token strings) over the
// whitespace tokens of s. When caseFold is true the tokens are lowercased
// first. js-rouge tokenizes case-sensitively by default, so ROUGE passes false;
// our BLEU lowercases, so it passes true.
func ngrams(s string, n int, caseFold bool) []string {
	if caseFold {
		s = strings.ToLower(s)
	}
	tokens := strings.Fields(s)
	if n < 1 || len(tokens) < n {
		return nil
	}
	out := make([]string, 0, len(tokens)-n+1)
	for i := 0; i+n <= len(tokens); i++ {
		out = append(out, strings.Join(tokens[i:i+n], " "))
	}
	return out
}

// rougeN reproduces js-rouge's rouge.n (the routine promptfoo calls): it forms
// the n-grams of candidate and reference, takes the set intersection of the
// distinct n-grams, then returns the F1 (beta=1) of precision and recall, where
// precision = |intersection| / (total candidate n-grams) and recall =
// |intersection| / (total reference n-grams). Tokenization is case-sensitive.
func rougeN(output, reference string, n int) float64 {
	cand := ngrams(output, n, false)
	ref := ngrams(reference, n, false)
	if len(cand) == 0 || len(ref) == 0 {
		return 0
	}
	candSet := make(map[string]bool, len(cand))
	for _, g := range cand {
		candSet[g] = true
	}
	refSet := make(map[string]bool, len(ref))
	for _, g := range ref {
		refSet[g] = true
	}
	overlap := 0
	for g := range candSet {
		if refSet[g] {
			overlap++
		}
	}
	if overlap == 0 {
		return 0
	}
	precision := float64(overlap) / float64(len(cand))
	recall := float64(overlap) / float64(len(ref))
	return fMeasure(precision, recall, 1.0)
}

// rougeL computes ROUGE-L: the F1 (beta=1) of LCS-based precision and recall,
// where precision = LCS / len(candidate tokens) and recall = LCS / len(reference
// tokens). This matches js-rouge and the reference implementation in
// github.com/tmc/misc/rouge. Tokenization is case-sensitive.
func rougeL(output, reference string) float64 {
	cand := strings.Fields(output)
	ref := strings.Fields(reference)
	if len(cand) == 0 || len(ref) == 0 {
		return 0
	}
	lcs := lcsLength(cand, ref)
	if lcs == 0 {
		return 0
	}
	precision := float64(lcs) / float64(len(cand))
	recall := float64(lcs) / float64(len(ref))
	return fMeasure(precision, recall, 1.0)
}

// lcsLength returns the length of the longest common subsequence of two token
// slices using the standard two-row dynamic-programming algorithm.
func lcsLength(a, b []string) int {
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else if prev[j] >= curr[j-1] {
				curr[j] = prev[j]
			} else {
				curr[j] = curr[j-1]
			}
		}
		prev, curr = curr, prev
		for j := range curr {
			curr[j] = 0
		}
	}
	return prev[len(b)]
}

// rougeS computes ROUGE-S: the F1 (beta=1) of the distinct skip-bigram set
// intersection, where a skip-bigram is any ordered pair of tokens (i<j) within
// a text. precision divides by the candidate's distinct skip-bigram count,
// recall by the reference's. Tokenization is case-sensitive, matching js-rouge.
func rougeS(output, reference string) float64 {
	candSet := skipBigrams(strings.Fields(output))
	refSet := skipBigrams(strings.Fields(reference))
	if len(candSet) == 0 || len(refSet) == 0 {
		return 0
	}
	overlap := 0
	for sb := range candSet {
		if refSet[sb] {
			overlap++
		}
	}
	if overlap == 0 {
		return 0
	}
	precision := float64(overlap) / float64(len(candSet))
	recall := float64(overlap) / float64(len(refSet))
	return fMeasure(precision, recall, 1.0)
}

// skipBigrams returns the set of distinct ordered token pairs (i<j) of tokens.
func skipBigrams(tokens []string) map[string]bool {
	set := make(map[string]bool)
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			set[tokens[i]+"\x00"+tokens[j]] = true
		}
	}
	return set
}

// fMeasure mirrors js-rouge's f-measure: with beta=1 it is the harmonic mean of
// precision and recall; beta weights recall over precision.
func fMeasure(precision, recall, beta float64) float64 {
	if precision == 0 && recall == 0 {
		return 0
	}
	b2 := beta * beta
	denom := b2*precision + recall
	if denom == 0 {
		return 0
	}
	return (1 + b2) * precision * recall / denom
}

// bleu returns a simplified BLEU score: the geometric mean of clipped n-gram
// precisions for n=1..maxN, times a brevity penalty. References promptfoo's
// pass-when->=threshold usage; not a full sacreBLEU but order-faithful.
func bleu(output, reference string, maxN int) float64 {
	outTokens := strings.Fields(strings.ToLower(output))
	refTokens := strings.Fields(strings.ToLower(reference))
	if len(outTokens) == 0 || len(refTokens) == 0 {
		return 0
	}
	logSum := 0.0
	for n := 1; n <= maxN; n++ {
		outGrams := countNgrams(output, n)
		refGrams := countNgrams(reference, n)
		total, clipped := 0, 0
		for g, c := range outGrams {
			total += c
			if refGrams[g] < c {
				clipped += refGrams[g]
			} else {
				clipped += c
			}
		}
		if total == 0 {
			return 0
		}
		precision := float64(clipped) / float64(total)
		if precision == 0 {
			return 0
		}
		logSum += math.Log(precision)
	}
	geoMean := math.Exp(logSum / float64(maxN))
	// Brevity penalty.
	bp := 1.0
	if len(outTokens) < len(refTokens) {
		bp = math.Exp(1.0 - float64(len(refTokens))/float64(len(outTokens)))
	}
	return bp * geoMean
}

// countNgrams counts lowercased n-grams; used by bleu's clipped precision.
func countNgrams(s string, n int) map[string]int {
	counts := make(map[string]int)
	for _, g := range ngrams(s, n, true) {
		counts[g]++
	}
	return counts
}

// gleu reproduces promptfoo's gleu (Google-BLEU): it pools the n-grams for
// n=minN..maxN of candidate and reference, counts the matches (clipped by
// reference availability), then returns min(precision, recall) where
// precision = matches / total candidate n-grams and recall = matches / total
// reference n-grams. Tokenization is case-sensitive, matching promptfoo.
func gleu(output, reference string, minN, maxN int) float64 {
	if minN < 1 {
		minN = 1
	}
	if maxN < minN {
		maxN = minN
	}
	candCounts := make(map[string]int)
	refCounts := make(map[string]int)
	candTotal, refTotal := 0, 0
	for n := minN; n <= maxN; n++ {
		for _, g := range ngrams(output, n, false) {
			candCounts[g]++
			candTotal++
		}
		for _, g := range ngrams(reference, n, false) {
			refCounts[g]++
			refTotal++
		}
	}
	if candTotal == 0 || refTotal == 0 {
		return 0
	}
	matches := 0
	for g, c := range candCounts {
		if r := refCounts[g]; r < c {
			matches += r
		} else {
			matches += c
		}
	}
	precision := float64(matches) / float64(candTotal)
	recall := float64(matches) / float64(refTotal)
	return math.Min(precision, recall)
}
