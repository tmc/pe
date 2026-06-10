package evaluator

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// HTML, XML, and SQL structural assertions, mirroring promptfoo's html.ts,
// xml.ts, and sql.ts. is-html/is-xml require the whole output to be a single
// well-formed document; the contains-* variants only require an embedded
// fragment to look like the target language.

// htmlIndicators are the regex signals promptfoo's containsHtml scores. The
// output is HTML when the total indicator score is >= 2. Opening and closing
// tags are scored jointly (2 if both present, 1 if only one), so they are
// handled separately from this table.
var (
	htmlOpenTag    = regexp.MustCompile(`<[a-zA-Z][a-zA-Z0-9-]*\s*(?:\s[^>]*)?>`)
	htmlCloseTag   = regexp.MustCompile(`</[a-zA-Z][a-zA-Z0-9-]*\s*>`)
	htmlSelfClose  = regexp.MustCompile(`<[a-zA-Z][a-zA-Z0-9-]*\s*(?:\s[^>]*)?/>`)
	htmlEntity     = regexp.MustCompile(`&(?:[a-zA-Z]+|#[0-9]+|#x[0-9a-fA-F]+);`)
	htmlDoctype    = regexp.MustCompile(`(?i)<!DOCTYPE\s+html`)
	htmlComment    = regexp.MustCompile(`<!--[^-]*(?:-[^-]+)*-->`)
	htmlAttribute  = regexp.MustCompile(`\s[a-zA-Z-]+=\s*["'][^"']*["']`)
	htmlCommonTags = regexp.MustCompile(`(?i)<(html|head|body|div|span|p|a|img|h[1-6]|ul|ol|li|table|tr|td|th|form|input|button|script|style|link|meta|br|hr)\b`)
	xmlDeclaration = regexp.MustCompile(`^\s*<\?xml`)
	// xmlFragment extracts a candidate XML/HTML fragment: the span from the
	// first opening tag to the last closing-or-self-closing tag.
	xmlFragment = regexp.MustCompile(`(?s)<[a-zA-Z!?][^>]*>.*<[^>]*>`)
)

// containsHTMLScore returns promptfoo's htmlIndicators score for output.
func containsHTMLScore(output string) int {
	if len(output) > 10*1024*1024 { // 10MB DoS cap, matching promptfoo.
		return 0
	}
	score := 0
	open := htmlOpenTag.MatchString(output)
	closing := htmlCloseTag.MatchString(output)
	switch {
	case open && closing:
		score += 2
	case open || closing:
		score++
	}
	if htmlSelfClose.MatchString(output) {
		score++
	}
	if htmlEntity.MatchString(output) {
		score++
	}
	if htmlDoctype.MatchString(output) {
		score += 2
	}
	if htmlComment.MatchString(output) {
		score++
	}
	if htmlAttribute.MatchString(output) {
		score++
	}
	if htmlCommonTags.MatchString(output) {
		score += 2
	}
	return score
}

// evaluateContainsHTML passes when output contains an HTML fragment, scored by
// promptfoo's indicator heuristic with a >= 2 pass threshold.
func (ae *AssertionEvaluator) evaluateContainsHTML(assertion Assertion, output string) *AssertionResult {
	score := containsHTMLScore(output)
	passed := score >= 2
	return boolResult(assertion, passed, fmt.Sprintf("HTML indicator score %d (threshold 2)", score))
}

// evaluateIsHTML passes when the whole output is well-formed HTML. promptfoo
// uses parse5; pe has no HTML parser dependency, so it requires the output to
// be a single well-formed XML-style tag tree (so any valid XHTML passes) that
// is not an XML document and carries strong HTML indicators. This is stricter
// than parse5 on malformed-but-recoverable HTML; the divergence is documented.
func (ae *AssertionEvaluator) evaluateIsHTML(assertion Assertion, output string) *AssertionResult {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return boolResult(assertion, false, "output is empty")
	}
	if xmlDeclaration.MatchString(trimmed) {
		return boolResult(assertion, false, "output is an XML document, not HTML")
	}
	if containsHTMLScore(trimmed) < 2 {
		return boolResult(assertion, false, "output does not look like HTML")
	}
	if !isWellFormedMarkup(trimmed) {
		return boolResult(assertion, false, "output is not well-formed HTML")
	}
	return boolResult(assertion, true, "output is well-formed HTML")
}

// evaluateIsXML passes when the whole output parses as well-formed XML, using
// Go's encoding/xml in the role promptfoo gives fast-xml-parser.
func (ae *AssertionEvaluator) evaluateIsXML(assertion Assertion, output string) *AssertionResult {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return boolResult(assertion, false, "output is empty")
	}
	if !isWellFormedMarkup(trimmed) {
		return boolResult(assertion, false, "output is not well-formed XML")
	}
	return boolResult(assertion, true, "output is well-formed XML")
}

// evaluateContainsXML passes when output embeds a well-formed XML fragment,
// mirroring promptfoo's extract-then-parse approach.
func (ae *AssertionEvaluator) evaluateContainsXML(assertion Assertion, output string) *AssertionResult {
	fragment := xmlFragment.FindString(output)
	if fragment == "" {
		return boolResult(assertion, false, "no XML fragment found")
	}
	if !isWellFormedMarkup(fragment) {
		return boolResult(assertion, false, "embedded fragment is not well-formed XML")
	}
	return boolResult(assertion, true, "output contains well-formed XML")
}

// isWellFormedMarkup reports whether s parses cleanly as a stream of XML tokens
// with balanced elements, used by is-xml/is-html/contains-xml.
func isWellFormedMarkup(s string) bool {
	dec := xml.NewDecoder(strings.NewReader(s))
	dec.Strict = true
	dec.AutoClose = nil
	depth := 0
	sawElement := false
	for {
		tok, err := dec.Token()
		if err != nil {
			// io.EOF is the only clean stop; anything else is malformed.
			return err.Error() == "EOF" && depth == 0 && sawElement
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
			sawElement = true
		case xml.EndElement:
			depth--
			if depth < 0 {
				return false
			}
		}
	}
}

// sqlFence extracts a SQL code block from markdown backticks (```sql ... ``` or
// `...`), mirroring promptfoo's /(?:sql)?([^`]+)/ extraction.
var sqlFence = regexp.MustCompile("(?s)```(?:sql)?\\s*(.*?)```|`([^`]+)`")

// evaluateContainsSQL passes when output contains a SQL statement, extracting a
// fenced code block when present (promptfoo's behavior) and falling back to the
// whole string, then validating shape with pe's local SQL checker.
func (ae *AssertionEvaluator) evaluateContainsSQL(assertion Assertion, output string) *AssertionResult {
	candidate := strings.TrimSpace(output)
	if m := sqlFence.FindStringSubmatch(output); m != nil {
		if m[1] != "" {
			candidate = strings.TrimSpace(m[1])
		} else {
			candidate = strings.TrimSpace(m[2])
		}
	}
	if candidate == "" {
		return boolResult(assertion, false, "no SQL found in output")
	}
	if err := validateSQLShape(candidate); err != nil {
		return boolResult(assertion, false, fmt.Sprintf("output does not contain valid SQL: %v", err))
	}
	return boolResult(assertion, true, "output contains valid SQL")
}
