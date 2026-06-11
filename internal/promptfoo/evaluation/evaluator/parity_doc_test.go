package evaluator

import (
	"os"
	"strings"
	"testing"
)

// docPath is the in-repo parity reference. The Assertion Parity section there
// is meant to mirror assertionAlias + the comparative/policy-blocked sets, so
// this test fails if code and doc drift apart.
const docPath = "../../../../docs/PROMPTFOO_INTEGRATION.md"

// comparativeIDs are handled by the post-pass (comparative.go), not the alias
// map; they are documented in their own subsection.
var comparativeIDs = []string{"select-best", "max-score"}

// policyBlockedIDs are deliberately unimplemented; the doc must explain each.
var policyBlockedIDs = []string{
	"javascript", "python", "ruby", "webhook",
	"moderation", "guardrails", "pi",
	"meteor", "human",
}

// TestParityDocCoversAllAcceptedIDs asserts every promptfoo id PE accepts is
// mentioned in the parity doc, so the doc cannot silently fall behind the code.
func TestParityDocCoversAllAcceptedIDs(t *testing.T) {
	doc := readDoc(t)
	for id := range assertionAlias {
		if !strings.Contains(doc, id) {
			t.Errorf("accepted assertion id %q is not documented in %s", id, docPath)
		}
	}
	for _, id := range comparativeIDs {
		if !strings.Contains(doc, id) {
			t.Errorf("comparative assertion id %q is not documented in %s", id, docPath)
		}
	}
	for _, id := range policyBlockedIDs {
		if !strings.Contains(doc, id) {
			t.Errorf("policy-blocked assertion id %q is not documented in %s", id, docPath)
		}
	}
}

// TestPolicyBlockedIDsStayUnimplemented guards the inverse: an id the doc lists
// as deliberately blocked must not have quietly gained an alias entry. If one
// does, the doc's policy table is now wrong and must be updated.
func TestPolicyBlockedIDsStayUnimplemented(t *testing.T) {
	for _, id := range policyBlockedIDs {
		if _, _, ok := normalizeAssertionType(id); ok {
			t.Errorf("assertion id %q is documented as policy-blocked but now resolves; update the parity doc", id)
		}
	}
}

func readDoc(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read parity doc: %v", err)
	}
	return string(b)
}
