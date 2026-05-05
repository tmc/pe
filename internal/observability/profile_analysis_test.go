package observability

import (
	"context"
	"testing"
	"time"
)

func TestAnalyzeProfileStats(t *testing.T) {
	analysis := AnalyzeProfileStats(ProfileStats{
		HeapObjects:   2_000_000,
		GCCPUFraction: 0.2,
		NumGoroutines: 2000,
	})
	if len(analysis.Findings) != 3 {
		t.Fatalf("findings = %v, want 3 findings", analysis.Findings)
	}
}

func TestContinuousProfiler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	profiler := NewContinuousProfiler(NewProfiler(t.TempDir()), time.Millisecond)
	profiler.Start(ctx)
	time.Sleep(5 * time.Millisecond)
	cancel()
	time.Sleep(time.Millisecond)

	if len(profiler.Samples()) == 0 {
		t.Fatal("continuous profiler collected no samples")
	}
}
