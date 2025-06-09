package metrics

import (
	"fmt"
	"testing"
	"time"
)

func TestPassNStoreBasic(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	store, err := NewPassNStore(tmpDir)
	if err != nil {
		t.Fatalf("NewPassNStore() error = %v", err)
	}
	if store == nil {
		t.Fatal("NewPassNStore returned nil")
	}

	// Test storing an evaluation
	evaluation := PassNEvaluation{
		ID:         "eval-test",
		Timestamp:  time.Now(),
		N:          1,
		NumSamples: 10,
		NumPassed:  8,
		PassRate:   0.8,
	}

	err = store.StoreEvaluation("prompt-1", "Test prompt", "provider-1", evaluation)
	if err != nil {
		t.Errorf("StoreEvaluation() error = %v", err)
	}

	// Test retrieving evaluations
	metadata, err := store.GetEvaluations("prompt-1")
	if err != nil {
		t.Errorf("GetEvaluations() error = %v", err)
	}

	if metadata == nil {
		t.Fatal("GetEvaluations() returned nil")
	}

	if len(metadata.Evaluations) == 0 {
		t.Error("No evaluations found")
	}
}

func TestPassNReport(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewPassNStore(tmpDir)
	if err != nil {
		t.Fatalf("NewPassNStore() error = %v", err)
	}

	// Store some evaluations
	for i := 0; i < 3; i++ {
		evaluation := PassNEvaluation{
			ID:         fmt.Sprintf("eval-%d", i),
			Timestamp:  time.Now(),
			N:          1,
			NumSamples: 20,
			NumPassed:  15 + i,
			PassRate:   float64(15+i) / 20.0,
		}

		err = store.StoreEvaluation("prompt-report", "Test prompt", "provider-1", evaluation)
		if err != nil {
			t.Errorf("StoreEvaluation() error = %v", err)
		}
	}

	// Generate report
	report, err := store.GeneratePassNReport("prompt-report")
	if err != nil {
		t.Errorf("GeneratePassNReport() error = %v", err)
	}

	if report == nil {
		t.Fatal("GeneratePassNReport() returned nil")
	}

	if report.TotalEvals != 3 {
		t.Errorf("TotalEvals = %d, want 3", report.TotalEvals)
	}
}
