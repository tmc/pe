package evaluator

import (
	"strings"
	"testing"
)

// policyBlockedIDs are promptfoo assertion ids pe deliberately does not
// implement (external code execution, hosted scoring APIs, the WordNet-backed
// meteor metric, and interactive human grading). They must resolve ok=false so
// an unmodified config surfaces an unsupported-assertion warning rather than
// silently passing. See docs/PROMPTFOO_INTEGRATION.md for the rationale per id.
var policyBlockedIDs = []string{
	"javascript", "python", "ruby", "webhook",
	"moderation", "guardrails", "pi",
	"meteor", "human",
}

// TestPolicyBlockedIDsUnresolved guards a real safety boundary: none of the
// deliberately-unimplemented ids may quietly gain an alias entry, which would
// un-block a check pe's policy forbids. normalizeAssertionType must report
// ok=false for each.
func TestPolicyBlockedIDsUnresolved(t *testing.T) {
	for _, id := range policyBlockedIDs {
		if _, _, ok := normalizeAssertionType(id); ok {
			t.Errorf("policy-blocked assertion id %q now resolves; it must stay unimplemented (update docs/PROMPTFOO_INTEGRATION.md if this is intentional)", id)
		}
	}
}

// TestAliasMapResolves round-trips the alias table against its own resolver:
// every key in assertionAlias must normalize to a non-empty canonical type with
// ok=true. This catches a key whose value drifts to the zero AssertionType or a
// normalization rule (lowercasing, "_"→"-") that stops a listed id from
// resolving.
//
// "not-"-prefixed keys are a deliberate exception to the value check:
// normalizeAssertionType strips the prefix before the lookup and returns
// negate=true, so e.g. "not-contains" resolves to (contains, negate=true) — not
// to the stored AssertionNotContains. Those keys must still resolve, but their
// canonical is the un-prefixed type.
func TestAliasMapResolves(t *testing.T) {
	for id, want := range assertionAlias {
		canonical, negate, ok := normalizeAssertionType(id)
		if !ok {
			t.Errorf("assertionAlias key %q does not resolve via normalizeAssertionType", id)
			continue
		}
		if strings.HasPrefix(id, "not-") {
			if !negate {
				t.Errorf("assertionAlias key %q should resolve with negate=true", id)
			}
			continue
		}
		if canonical != want {
			t.Errorf("assertionAlias[%q] = %q, but normalizeAssertionType resolved %q", id, want, canonical)
		}
	}
}
