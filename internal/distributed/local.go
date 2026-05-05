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
