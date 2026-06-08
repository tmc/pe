// Package distributed provides local deterministic execution helpers.
//
// The package is intentionally local-only. It does not start daemons, discover
// peers, open sockets, or coordinate work outside the current process.
package distributed

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Task is a unit of local work.
type Task[T any] struct {
	ID  string
	Run func(context.Context) (T, error)
}

// DAGTask is a local task with parent dependencies.
//
// A task is runnable after every task named in After has completed
// successfully.
type DAGTask[T any] struct {
	Task[T]
	After []string
}

// Result is the outcome for one task.
type Result[T any] struct {
	ID    string
	Value T
	Err   error
}

// RunLocal runs tasks with at most workers concurrent task functions.
//
// Results are returned in the same order as tasks. RunLocal stops scheduling
// new work after the first task error or context cancellation, but it waits for
// already-started tasks to finish before returning.
func RunLocal[T any](ctx context.Context, workers int, tasks []Task[T]) ([]Result[T], error) {
	if workers < 1 {
		workers = 1
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type item struct {
		index int
		task  Task[T]
	}

	jobs := make(chan item)
	results := make([]Result[T], len(tasks))
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	recordErr := func(err error) {
		if err == nil {
			return
		}
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				result := Result[T]{ID: job.task.ID}
				if job.task.Run == nil {
					result.Err = fmt.Errorf("task %q has no run function", job.task.ID)
				} else {
					result.Value, result.Err = job.task.Run(ctx)
				}
				results[job.index] = result
				recordErr(result.Err)
			}
		}()
	}

	for i, task := range tasks {
		if err := ctx.Err(); err != nil {
			recordErr(err)
			break
		}
		select {
		case jobs <- item{index: i, task: task}:
		case <-ctx.Done():
			recordErr(ctx.Err())
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}
	return results, nil
}

// RunDAGLocal runs tasks after their dependencies with at most workers
// concurrent task functions.
//
// Results are returned in the same order as tasks. Cycles, duplicate task IDs,
// missing dependencies, and empty task IDs are rejected before any task runs.
// If a task fails, dependent tasks are skipped with an error.
func RunDAGLocal[T any](ctx context.Context, workers int, tasks []DAGTask[T]) ([]Result[T], error) {
	if workers < 1 {
		workers = 1
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	indexByID := make(map[string]int, len(tasks))
	for i, task := range tasks {
		if task.ID == "" {
			return nil, fmt.Errorf("task %d has empty id", i)
		}
		if _, ok := indexByID[task.ID]; ok {
			return nil, fmt.Errorf("duplicate task id %q", task.ID)
		}
		indexByID[task.ID] = i
	}

	children := make([][]int, len(tasks))
	remaining := make([]int, len(tasks))
	for i, task := range tasks {
		seen := make(map[string]bool, len(task.After))
		for _, parentID := range task.After {
			parent, ok := indexByID[parentID]
			if !ok {
				return nil, fmt.Errorf("task %q depends on missing task %q", task.ID, parentID)
			}
			if seen[parentID] {
				continue
			}
			seen[parentID] = true
			children[parent] = append(children[parent], i)
			remaining[i]++
		}
	}
	if err := checkDAG(tasks, children, remaining); err != nil {
		return nil, err
	}

	results := make([]Result[T], len(tasks))
	for i, task := range tasks {
		results[i].ID = task.ID
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var firstErr error
	var zero T
	blockedBy := make([]error, len(tasks))
	ready := make([]int, 0, len(tasks))
	for i := range tasks {
		if remaining[i] == 0 {
			ready = append(ready, i)
		}
	}

	for len(ready) > 0 {
		sort.Ints(ready)
		if err := ctx.Err(); err != nil {
			for _, i := range ready {
				results[i].Err = err
				results[i].Value = zero
			}
			if firstErr == nil {
				firstErr = err
			}
			break
		}

		run, skip := splitReady(ready, blockedBy)
		for _, i := range skip {
			results[i].Err = blockedBy[i]
			if firstErr == nil {
				firstErr = results[i].Err
			}
		}
		if len(run) > 0 {
			runDAGReady(ctx, workers, tasks, run, results)
			for _, i := range run {
				if results[i].Err != nil && firstErr == nil {
					firstErr = results[i].Err
				}
			}
		}

		next := make([]int, 0, len(ready))
		for _, parent := range ready {
			parentErr := results[parent].Err
			for _, child := range children[parent] {
				if parentErr != nil && blockedBy[child] == nil {
					blockedBy[child] = fmt.Errorf("task %q skipped after dependency %q failed: %w", tasks[child].ID, tasks[parent].ID, parentErr)
				}
				remaining[child]--
				if remaining[child] == 0 {
					next = append(next, child)
				}
			}
		}
		ready = next
	}

	if firstErr != nil {
		return results, firstErr
	}
	return results, nil
}

func checkDAG[T any](tasks []DAGTask[T], children [][]int, remaining []int) error {
	degrees := append([]int(nil), remaining...)
	queue := make([]int, 0, len(tasks))
	for i, degree := range degrees {
		if degree == 0 {
			queue = append(queue, i)
		}
	}

	visited := 0
	for len(queue) > 0 {
		sort.Ints(queue)
		node := queue[0]
		queue = queue[1:]
		visited++
		for _, child := range children[node] {
			degrees[child]--
			if degrees[child] == 0 {
				queue = append(queue, child)
			}
		}
	}
	if visited != len(tasks) {
		return fmt.Errorf("task graph contains a cycle")
	}
	return nil
}

func splitReady(ready []int, blockedBy []error) (run []int, skip []int) {
	for _, i := range ready {
		if blockedBy[i] != nil {
			skip = append(skip, i)
		} else {
			run = append(run, i)
		}
	}
	return run, skip
}

func runDAGReady[T any](ctx context.Context, workers int, tasks []DAGTask[T], ready []int, results []Result[T]) {
	type item struct {
		index int
		task  DAGTask[T]
	}

	jobs := make(chan item)
	var wg sync.WaitGroup
	if workers > len(ready) {
		workers = len(ready)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				result := Result[T]{ID: job.task.ID}
				if job.task.Run == nil {
					result.Err = fmt.Errorf("task %q has no run function", job.task.ID)
				} else {
					result.Value, result.Err = job.task.Run(ctx)
				}
				results[job.index] = result
			}
		}()
	}

	for _, index := range ready {
		if err := ctx.Err(); err != nil {
			results[index].Err = err
			continue
		}
		select {
		case jobs <- item{index: index, task: tasks[index]}:
		case <-ctx.Done():
			results[index].Err = ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
}

// Vote is one provider's local response for a consensus decision.
type Vote struct {
	Provider string
	Output   string
	Weight   int
}

// Consensus is a deterministic aggregation of provider votes.
type Consensus struct {
	Output    string
	Weight    int
	Count     int
	Total     int
	Providers []string
}

// Majority returns the highest-weight response after normalizing output
// whitespace. Ties are resolved by first appearance in votes.
func Majority(votes []Vote) (Consensus, error) {
	if len(votes) == 0 {
		return Consensus{}, errors.New("no votes")
	}

	type bucket struct {
		output    string
		weight    int
		count     int
		first     int
		providers []string
	}

	buckets := make(map[string]*bucket)
	total := 0
	for i, vote := range votes {
		output := normalize(vote.Output)
		if output == "" {
			continue
		}
		weight := vote.Weight
		if weight <= 0 {
			weight = 1
		}
		total += weight

		b := buckets[output]
		if b == nil {
			b = &bucket{output: output, first: i}
			buckets[output] = b
		}
		b.weight += weight
		b.count++
		if vote.Provider != "" {
			b.providers = append(b.providers, vote.Provider)
		}
	}
	if len(buckets) == 0 {
		return Consensus{}, errors.New("no non-empty votes")
	}

	var best *bucket
	for _, b := range buckets {
		if best == nil || b.weight > best.weight || (b.weight == best.weight && b.first < best.first) {
			best = b
		}
	}
	sort.Strings(best.providers)

	return Consensus{
		Output:    best.output,
		Weight:    best.weight,
		Count:     best.count,
		Total:     total,
		Providers: append([]string(nil), best.providers...),
	}, nil
}

func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
