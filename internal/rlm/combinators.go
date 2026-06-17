package rlm

import (
	"context"
	"fmt"
	"strings"
)

// Chunk is a bounded byte range of a target artifact, addressed out-of-prompt by
// its content key.
type Chunk struct {
	Index    int
	Start    int
	End      int
	CacheKey string
	Data     []byte
}

// chunk splits data into contiguous ranges of at most size bytes. The final
// chunk may be shorter. A non-positive size is rejected.
func chunk(data []byte, size int) ([]Chunk, error) {
	if size <= 0 {
		return nil, fmt.Errorf("chunk size must be positive, got %d", size)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var chunks []Chunk
	for start := 0; start < len(data); start += size {
		end := min(start+size, len(data))
		payload := data[start:end]
		chunks = append(chunks, Chunk{
			Index:    len(chunks),
			Start:    start,
			End:      end,
			CacheKey: ContentKey(payload),
			Data:     payload,
		})
	}
	return chunks, nil
}

// Worker runs one bounded prompt over one chunk and reports its token cost.
//
// Implementations must not invent new operations or execute model-generated
// code. The runner enforces all budgets before calling Run.
type Worker interface {
	// Name identifies the worker for trace records, e.g. "provider:model".
	Name() string

	// Run answers prompt for the given chunk. The returned cost reports token
	// usage; a zero cost is valid for deterministic local workers.
	Run(ctx context.Context, prompt string, c Chunk) (output string, cost Cost, err error)
}

// mapResult pairs a chunk with the worker output for that chunk.
type mapResult struct {
	chunk  Chunk
	output string
	cost   Cost
	err    error
}

// reduce combines child outputs with a deterministic rule. The combine function
// receives outputs in chunk order and must be a pure aggregation.
func reduce(outputs []string, combine func([]string) string) string {
	if combine == nil {
		combine = joinLines
	}
	return combine(append([]string(nil), outputs...))
}

func joinLines(outputs []string) string {
	return strings.Join(outputs, "\n")
}
