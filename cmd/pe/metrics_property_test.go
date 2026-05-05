package main

import (
	"fmt"
	"math"
	"testing"
	"testing/quick"
)

func TestCalculateOverallScorePropertyBounds(t *testing.T) {
	property := func(values []uint8) bool {
		scores := make(map[string]float64, len(values))
		minScore, maxScore := math.Inf(1), math.Inf(-1)
		for i, value := range values {
			score := float64(value) / 255
			scores[fmt.Sprint(i)] = score
			minScore = math.Min(minScore, score)
			maxScore = math.Max(maxScore, score)
		}
		got := calculateOverallScore(scores)
		if len(scores) == 0 {
			return got == 0
		}
		return got >= minScore && got <= maxScore
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}
