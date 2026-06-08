package distributed

import (
	"context"
	"errors"
	"reflect"
	"strconv"
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

func TestRunDAGLocalTopologicalOrder(t *testing.T) {
	var order []string
	tasks := []DAGTask[int]{
		{Task: Task[int]{ID: "render", Run: func(context.Context) (int, error) {
			order = append(order, "render")
			return 3, nil
		}}, After: []string{"parse", "validate"}},
		{Task: Task[int]{ID: "parse", Run: func(context.Context) (int, error) {
			order = append(order, "parse")
			return 1, nil
		}}},
		{Task: Task[int]{ID: "validate", Run: func(context.Context) (int, error) {
			order = append(order, "validate")
			return 2, nil
		}}, After: []string{"parse"}},
	}

	results, err := RunDAGLocal(context.Background(), 1, tasks)
	if err != nil {
		t.Fatalf("RunDAGLocal: %v", err)
	}
	if !reflect.DeepEqual(order, []string{"parse", "validate", "render"}) {
		t.Fatalf("order = %v, want parse validate render", order)
	}
	if got := []int{results[0].Value, results[1].Value, results[2].Value}; !reflect.DeepEqual(got, []int{3, 1, 2}) {
		t.Fatalf("values = %v, want [3 1 2]", got)
	}
}

func TestRunDAGLocalCycleDetection(t *testing.T) {
	var ran atomic.Bool
	tasks := []DAGTask[int]{
		{Task: Task[int]{ID: "a", Run: func(context.Context) (int, error) {
			ran.Store(true)
			return 0, nil
		}}, After: []string{"b"}},
		{Task: Task[int]{ID: "b", Run: func(context.Context) (int, error) {
			ran.Store(true)
			return 0, nil
		}}, After: []string{"a"}},
	}

	results, err := RunDAGLocal(context.Background(), 2, tasks)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("RunDAGLocal error = %v, want cycle error", err)
	}
	if results != nil {
		t.Fatalf("results = %+v, want nil", results)
	}
	if ran.Load() {
		t.Fatal("cycle should be rejected before running tasks")
	}
}

func TestRunDAGLocalParentFailureSkipsChildren(t *testing.T) {
	wantErr := errors.New("policy denied")
	tasks := []DAGTask[int]{
		{Task: Task[int]{ID: "policy", Run: func(context.Context) (int, error) {
			return 0, wantErr
		}}},
		{Task: Task[int]{ID: "child", Run: func(context.Context) (int, error) {
			t.Fatal("child should be skipped after parent failure")
			return 0, nil
		}}, After: []string{"policy"}},
		{Task: Task[int]{ID: "grandchild", Run: func(context.Context) (int, error) {
			t.Fatal("grandchild should be skipped after ancestor failure")
			return 0, nil
		}}, After: []string{"child"}},
	}

	results, err := RunDAGLocal(context.Background(), 2, tasks)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunDAGLocal error = %v, want %v", err, wantErr)
	}
	if !errors.Is(results[0].Err, wantErr) {
		t.Fatalf("parent error = %v, want %v", results[0].Err, wantErr)
	}
	if results[1].Err == nil || !strings.Contains(results[1].Err.Error(), `task "child" skipped after dependency "policy" failed`) {
		t.Fatalf("child error = %v, want dependency skip", results[1].Err)
	}
	if results[2].Err == nil || !strings.Contains(results[2].Err.Error(), `task "grandchild" skipped after dependency "child" failed`) {
		t.Fatalf("grandchild error = %v, want dependency skip", results[2].Err)
	}
}

func TestRunDAGLocalHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var ranSecond atomic.Bool

	tasks := []DAGTask[int]{
		{Task: Task[int]{ID: "cancel", Run: func(context.Context) (int, error) {
			cancel()
			return 1, nil
		}}},
		{Task: Task[int]{ID: "after", Run: func(context.Context) (int, error) {
			ranSecond.Store(true)
			return 2, nil
		}}},
	}

	results, err := RunDAGLocal(ctx, 1, tasks)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunDAGLocal error = %v, want %v", err, context.Canceled)
	}
	if results[0].Err != nil || results[0].Value != 1 {
		t.Fatalf("first result = %+v, want successful cancellation trigger", results[0])
	}
	if !errors.Is(results[1].Err, context.Canceled) {
		t.Fatalf("second result error = %v, want %v", results[1].Err, context.Canceled)
	}
	if ranSecond.Load() {
		t.Fatal("second task ran after context cancellation")
	}
}

func TestRunDAGLocalLimitsWorkers(t *testing.T) {
	var running int32
	var maxRunning int32

	tasks := make([]DAGTask[int], 20)
	for i := range tasks {
		i := i
		tasks[i] = DAGTask[int]{
			Task: Task[int]{
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
			},
		}
	}

	if _, err := RunDAGLocal(context.Background(), 3, tasks); err != nil {
		t.Fatalf("RunDAGLocal: %v", err)
	}
	if maxRunning > 3 {
		t.Fatalf("max running = %d, want at most 3", maxRunning)
	}
}

func TestRunDAGLocalRejectsInvalidGraph(t *testing.T) {
	tests := []struct {
		name  string
		tasks []DAGTask[int]
		want  string
	}{
		{
			name:  "empty id",
			tasks: []DAGTask[int]{{Task: Task[int]{Run: func(context.Context) (int, error) { return 0, nil }}}},
			want:  "empty id",
		},
		{
			name: "duplicate",
			tasks: []DAGTask[int]{
				{Task: Task[int]{ID: "same", Run: func(context.Context) (int, error) { return 0, nil }}},
				{Task: Task[int]{ID: "same", Run: func(context.Context) (int, error) { return 0, nil }}},
			},
			want: "duplicate task id",
		},
		{
			name: "missing dependency",
			tasks: []DAGTask[int]{
				{Task: Task[int]{ID: "child", Run: func(context.Context) (int, error) { return 0, nil }}, After: []string{"missing"}},
			},
			want: "depends on missing task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := RunDAGLocal(context.Background(), 1, tt.tasks)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RunDAGLocal error = %v, want %q", err, tt.want)
			}
			if results != nil {
				t.Fatalf("results = %+v, want nil", results)
			}
		})
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

func BenchmarkRunLocal(b *testing.B) {
	for _, workers := range []int{1, 4} {
		b.Run("workers="+strconv.Itoa(workers), func(b *testing.B) {
			tasks := make([]Task[int], 32)
			for i := range tasks {
				i := i
				tasks[i] = Task[int]{
					ID: string(rune('a' + i%26)),
					Run: func(context.Context) (int, error) {
						return i, nil
					},
				}
			}

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				results, err := RunLocal(context.Background(), workers, tasks)
				if err != nil {
					b.Fatal(err)
				}
				if len(results) != len(tasks) || results[len(results)-1].Value != len(tasks)-1 {
					b.Fatalf("RunLocal returned %d results, last=%d", len(results), results[len(results)-1].Value)
				}
			}
		})
	}
}
