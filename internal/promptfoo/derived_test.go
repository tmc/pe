package promptfoo

import (
	"math"
	"testing"
)

func TestEvalExpr(t *testing.T) {
	env := map[string]float64{
		"Consistency": 0.5,
		"TotalScore":  9,
		"__count":     3,
	}
	tests := []struct {
		expr    string
		want    float64
		wantErr bool
	}{
		{"Consistency * 2", 1.0, false},
		{"TotalScore / __count", 3.0, false},
		{"1 + 2 * 3", 7.0, false},
		{"(1 + 2) * 3", 9.0, false},
		{"10 - 4 - 3", 3.0, false},
		{"-Consistency + 1", 0.5, false},
		{"2.5 * 4", 10.0, false},
		{"TotalScore / 0", 0, true},
		{"Missing + 1", 0, true},
		{"", 0, true},
		{"1 +", 0, true},
		{"(1 + 2", 0, true},
		{"1 2", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := evalExpr(tt.expr, env)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("evalExpr(%q) = %v, want error", tt.expr, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("evalExpr(%q) error: %v", tt.expr, err)
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("evalExpr(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestComputeDerivedMetrics(t *testing.T) {
	metrics := []DerivedMetric{
		{Name: "DoubleConsistency", Value: "Consistency * 2"},
		{Name: "Average", Value: "TotalScore / __count"},
		{Name: "Chained", Value: "DoubleConsistency + 1"}, // depends on an earlier derived metric
	}
	named := map[string]float64{"Consistency": 0.4, "TotalScore": 9}
	out, err := ComputeDerivedMetrics(metrics, named, 3)
	if err != nil {
		t.Fatalf("ComputeDerivedMetrics: %v", err)
	}
	if math.Abs(out["DoubleConsistency"]-0.8) > 1e-9 {
		t.Fatalf("DoubleConsistency = %v, want 0.8", out["DoubleConsistency"])
	}
	if math.Abs(out["Average"]-3.0) > 1e-9 {
		t.Fatalf("Average = %v, want 3.0", out["Average"])
	}
	if math.Abs(out["Chained"]-1.8) > 1e-9 {
		t.Fatalf("Chained = %v, want 1.8", out["Chained"])
	}
}

func TestComputeDerivedMetricsReportsBadExpr(t *testing.T) {
	metrics := []DerivedMetric{
		{Name: "Good", Value: "1 + 1"},
		{Name: "Bad", Value: "Nonexistent * 2"},
	}
	out, err := ComputeDerivedMetrics(metrics, nil, 1)
	if err == nil {
		t.Fatal("expected an error for the bad metric")
	}
	// The good metric is still computed.
	if out["Good"] != 2 {
		t.Fatalf("Good = %v, want 2", out["Good"])
	}
	if _, ok := out["Bad"]; ok {
		t.Fatal("Bad metric should be skipped")
	}
}
