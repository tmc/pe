package module

import (
	"fmt"
	"testing"
	"testing/quick"
)

func TestVersionConstraintPropertyExactMatchesSelf(t *testing.T) {
	property := func(major, minor, patch uint8) bool {
		version := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
		constraint, err := ParseVersionConstraint(version)
		if err != nil {
			return false
		}
		return constraint.Matches(version)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}
