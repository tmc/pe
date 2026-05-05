package distributed

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunLocalPreservesOrder(t *testing.T) {
	tasks := []Task[int]{
		{ID: "slow", Run: func(context.Context) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 1, nil
		}},
		{ID: "fast", Run: func(context.Context) (int, error) {
			return 2, nil
		}},
		{ID: "last", Run: func(context.Context) (int, error) {
			return 3, nil
		}},
	}

	results, err := RunLocal(context.Background(), 3, tasks)
	if err != nil {
		t.Fatalf("RunLocal: %v", err)
	}

	gotIDs := []string{results[0].ID, results[1].ID, results[2].ID}
	wantIDs := []string{"slow", "fast", "last"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("ids = %v, want %v", gotIDs, wantIDs)
	}
	for i, result := range results {
		if result.Value != i+1 {
			t.Fatalf("result %d value = %d, want %d", i, result.Value, i+1)
		}
	}
}

func TestRunLocalStopsOnError(t *testing.T) {
	wantErr := errors.New("boom")
	tasks := []Task[int]{
		{ID: "ok", Run: func(context.Context) (int, error) {
			return 1, nil
		}},
		{ID: "bad", Run: func(context.Context) (int, error) {
			return 0, wantErr
		}},
		{ID: "skipped", Run: func(context.Context) (int, error) {
			t.Fatal("task should not be scheduled after first error")
			return 0, nil
		}},
	}

	results, err := RunLocal(context.Background(), 1, tasks)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunLocal error = %v, want %v", err, wantErr)
	}
	if results[0].Err != nil {
		t.Fatalf("first task error = %v", results[0].Err)
	}
	if !errors.Is(results[1].Err, wantErr) {
		t.Fatalf("second task error = %v, want %v", results[1].Err, wantErr)
	}
	if results[2].ID != "" {
		t.Fatalf("third task was scheduled: %+v", results[2])
	}
}

func TestRunLocalLimitsWorkers(t *testing.T) {
	var running int32
	var maxRunning int32

	tasks := make([]Task[int], 20)
	for i := range tasks {
		i := i
		tasks[i] = Task[int]{
			ID: string(rune('a' + i)),
			Run: func(context.Context) (int, error) {
				n := atomic.AddInt32(&running, 1)
				defer atomic.AddInt32(&running, -1)
				for {
					old := atomic.LoadInt32(&maxRunning)
					if n <= old || atomic.CompareAndSwapInt32(&maxRunning, old, n) {
						break
					}
				}
				time.Sleep(time.Millisecond)
				return i, nil
			},
		}
	}

	if _, err := RunLocal(context.Background(), 3, tasks); err != nil {
		t.Fatalf("RunLocal: %v", err)
	}
	if maxRunning > 3 {
		t.Fatalf("max running = %d, want at most 3", maxRunning)
	}
}

func TestRunLocalNilTaskReturnsIndexedError(t *testing.T) {
	results, err := RunLocal(context.Background(), 1, []Task[int]{
		{ID: "nil"},
		{ID: "after", Run: func(context.Context) (int, error) {
			t.Fatal("task after nil task should not be scheduled")
			return 0, nil
		}},
	})
	if err == nil || !strings.Contains(err.Error(), `task "nil" has no run function`) {
		t.Fatalf("RunLocal error = %v, want nil task error", err)
	}
	if results[0].ID != "nil" || results[0].Err == nil {
		t.Fatalf("first result = %+v, want indexed nil task error", results[0])
	}
	if results[1].ID != "" {
		t.Fatalf("second task was scheduled: %+v", results[1])
	}
}

func TestRunLocalHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tasks := []Task[int]{
		{ID: "never", Run: func(context.Context) (int, error) {
			t.Fatal("task should not run after cancellation")
			return 0, nil
		}},
	}

	results, err := RunLocal(ctx, 2, tasks)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunLocal error = %v, want %v", err, context.Canceled)
	}
	if results[0].ID != "" {
		t.Fatalf("canceled task was scheduled: %+v", results[0])
	}
}

func TestMajorityAggregatesDeterministically(t *testing.T) {
	votes := []Vote{
		{Provider: "b", Output: "Paris", Weight: 1},
		{Provider: "c", Output: " Lyon ", Weight: 2},
		{Provider: "a", Output: "Paris", Weight: 2},
		{Provider: "d", Output: "Paris\n", Weight: 1},
	}

	got, err := Majority(votes)
	if err != nil {
		t.Fatalf("Majority: %v", err)
	}

	want := Consensus{
		Output:    "Paris",
		Weight:    4,
		Count:     3,
		Total:     6,
		Providers: []string{"a", "b", "d"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Majority() = %+v, want %+v", got, want)
	}
}

func TestMajorityBreaksTiesByFirstVote(t *testing.T) {
	got, err := Majority([]Vote{
		{Provider: "b", Output: "first", Weight: 2},
		{Provider: "a", Output: "second", Weight: 2},
	})
	if err != nil {
		t.Fatalf("Majority: %v", err)
	}
	if got.Output != "first" {
		t.Fatalf("output = %q, want first", got.Output)
	}
	if !reflect.DeepEqual(got.Providers, []string{"b"}) {
		t.Fatalf("providers = %v, want [b]", got.Providers)
	}
}

func TestMajorityIgnoresEmptyVotesAndCountsTotalWeight(t *testing.T) {
	got, err := Majority([]Vote{
		{Provider: "empty", Output: " \n\t", Weight: 100},
		{Provider: "zero", Output: "yes", Weight: 0},
		{Provider: "weighted", Output: "yes", Weight: 3},
	})
	if err != nil {
		t.Fatalf("Majority: %v", err)
	}
	if got.Output != "yes" {
		t.Fatalf("output = %q, want yes", got.Output)
	}
	if got.Weight != 4 {
		t.Fatalf("weight = %d, want 4", got.Weight)
	}
	if got.Total != 4 {
		t.Fatalf("total = %d, want 4", got.Total)
	}
	if !reflect.DeepEqual(got.Providers, []string{"weighted", "zero"}) {
		t.Fatalf("providers = %v, want [weighted zero]", got.Providers)
	}
}
