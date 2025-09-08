// Package testing provides testing utilities and framework for the PE toolkit.
package testing

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/inference"
)

// TestFramework provides comprehensive testing utilities for PE components.
type TestFramework struct {
	t        *testing.T
	ctx      context.Context
	rand     *rand.Rand
	timeout  time.Duration
	parallel bool
}

// NewTestFramework creates a new test framework instance.
func NewTestFramework(t *testing.T) *TestFramework {
	return &TestFramework{
		t:       t,
		ctx:     context.Background(),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		timeout: 30 * time.Second,
	}
}

// WithContext sets a custom context for the test framework.
func (tf *TestFramework) WithContext(ctx context.Context) *TestFramework {
	tf.ctx = ctx
	return tf
}

// WithTimeout sets a timeout for test operations.
func (tf *TestFramework) WithTimeout(timeout time.Duration) *TestFramework {
	tf.timeout = timeout
	return tf
}

// WithParallel enables parallel execution for tests.
func (tf *TestFramework) WithParallel() *TestFramework {
	tf.parallel = true
	return tf
}

// Run executes a test function with the configured context and timeout.
func (tf *TestFramework) Run(name string, fn func(*TestFramework)) {
	tf.t.Run(name, func(t *testing.T) {
		if tf.parallel {
			t.Parallel()
		}
		
		ctx, cancel := context.WithTimeout(tf.ctx, tf.timeout)
		defer cancel()
		
		subFramework := &TestFramework{
			t:        t,
			ctx:      ctx,
			rand:     tf.rand,
			timeout:  tf.timeout,
			parallel: tf.parallel,
		}
		
		fn(subFramework)
	})
}

// Context returns the test context.
func (tf *TestFramework) Context() context.Context {
	return tf.ctx
}

// T returns the testing.T instance.
func (tf *TestFramework) T() *testing.T {
	return tf.t
}

// Rand returns the random number generator.
func (tf *TestFramework) Rand() *rand.Rand {
	return tf.rand
}

// AssertNoError fails the test if err is not nil.
func (tf *TestFramework) AssertNoError(err error, msg ...string) {
	tf.t.Helper()
	if err != nil {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: %v", strings.Join(msg, " "), err)
		} else {
			tf.t.Fatalf("unexpected error: %v", err)
		}
	}
}

// AssertError fails the test if err is nil.
func (tf *TestFramework) AssertError(err error, msg ...string) {
	tf.t.Helper()
	if err == nil {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: expected error but got nil", strings.Join(msg, " "))
		} else {
			tf.t.Fatal("expected error but got nil")
		}
	}
}

// AssertEqual fails the test if got != want.
func (tf *TestFramework) AssertEqual(got, want interface{}, msg ...string) {
	tf.t.Helper()
	if got != want {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: got %v, want %v", strings.Join(msg, " "), got, want)
		} else {
			tf.t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// AssertNotEqual fails the test if got == want.
func (tf *TestFramework) AssertNotEqual(got, want interface{}, msg ...string) {
	tf.t.Helper()
	if got == want {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: got %v, expected different value", strings.Join(msg, " "), got)
		} else {
			tf.t.Fatalf("got %v, expected different value", got)
		}
	}
}

// AssertContains fails the test if s does not contain substr.
func (tf *TestFramework) AssertContains(s, substr string, msg ...string) {
	tf.t.Helper()
	if !strings.Contains(s, substr) {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: %q does not contain %q", strings.Join(msg, " "), s, substr)
		} else {
			tf.t.Fatalf("%q does not contain %q", s, substr)
		}
	}
}

// AssertNotEmpty fails the test if s is empty.
func (tf *TestFramework) AssertNotEmpty(s string, msg ...string) {
	tf.t.Helper()
	if s == "" {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: expected non-empty string", strings.Join(msg, " "))
		} else {
			tf.t.Fatal("expected non-empty string")
		}
	}
}

// AssertTrue fails the test if condition is false.
func (tf *TestFramework) AssertTrue(condition bool, msg ...string) {
	tf.t.Helper()
	if !condition {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: condition was false", strings.Join(msg, " "))
		} else {
			tf.t.Fatal("condition was false")
		}
	}
}

// AssertFalse fails the test if condition is true.
func (tf *TestFramework) AssertFalse(condition bool, msg ...string) {
	tf.t.Helper()
	if condition {
		if len(msg) > 0 {
			tf.t.Fatalf("%s: condition was true", strings.Join(msg, " "))
		} else {
			tf.t.Fatal("condition was true")
		}
	}
}

// AssertPanics fails the test if fn does not panic.
func (tf *TestFramework) AssertPanics(fn func(), msg ...string) {
	tf.t.Helper()
	defer func() {
		if r := recover(); r == nil {
			if len(msg) > 0 {
				tf.t.Fatalf("%s: expected panic but none occurred", strings.Join(msg, " "))
			} else {
				tf.t.Fatal("expected panic but none occurred")
			}
		}
	}()
	fn()
}

// AssertNotPanics fails the test if fn panics.
func (tf *TestFramework) AssertNotPanics(fn func(), msg ...string) {
	tf.t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msg) > 0 {
				tf.t.Fatalf("%s: unexpected panic: %v", strings.Join(msg, " "), r)
			} else {
				tf.t.Fatalf("unexpected panic: %v", r)
			}
		}
	}()
	fn()
}

// TestData provides common test data generation utilities.
type TestData struct {
	rand *rand.Rand
}

// NewTestData creates a new test data generator.
func NewTestData(seed int64) *TestData {
	return &TestData{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// RandomString generates a random string of the specified length.
func (td *TestData) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[td.rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomPrompt generates a random prompt for testing.
func (td *TestData) RandomPrompt() string {
	prompts := []string{
		"What is the capital of France?",
		"Explain quantum computing in simple terms.",
		"Write a haiku about artificial intelligence.",
		"Calculate the area of a circle with radius 5.",
		"Translate 'hello world' to Spanish.",
		"What are the benefits of renewable energy?",
		"Describe the process of photosynthesis.",
		"Write a short story about time travel.",
	}
	return prompts[td.rand.Intn(len(prompts))]
}

// RandomInferenceRequest generates a random inference request for testing.
func (td *TestData) RandomInferenceRequest() inference.Request {
	return inference.Request{
		Prompt:        td.RandomPrompt(),
		Model:         td.RandomModel(),
		Temperature:   td.rand.Float32(),
		MaxTokens:     100 + td.rand.Intn(900),
		SystemPrompt:  td.RandomString(50),
		StopSequences: []string{td.RandomString(5)},
		Options:       map[string]interface{}{"test": true},
	}
}

// RandomModel generates a random model name for testing.
func (td *TestData) RandomModel() string {
	models := []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"claude-3-sonnet",
		"claude-3-haiku",
		"llama-2-7b",
		"llama-2-13b",
	}
	return models[td.rand.Intn(len(models))]
}

// RandomResponse generates a random inference response for testing.
func (td *TestData) RandomResponse(model string) *inference.Response {
	return &inference.Response{
		Content: td.RandomString(100),
		Model:   model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     td.rand.Intn(1000),
			CompletionTokens: td.rand.Intn(1000),
			TotalTokens:      td.rand.Intn(2000),
		},
		Metadata: map[string]interface{}{
			"provider":      "test",
			"test_data":     true,
			"finish_reason": "stop",
		},
	}
}

// PropertyTest represents a property-based test.
type PropertyTest struct {
	Name       string
	Property   func(interface{}) bool
	Generator  func(*rand.Rand) interface{}
	Iterations int
}

// RunPropertyTests runs a set of property-based tests.
func (tf *TestFramework) RunPropertyTests(tests []PropertyTest) {
	for _, test := range tests {
		tf.Run(test.Name, func(tf *TestFramework) {
			iterations := test.Iterations
			if iterations == 0 {
				iterations = 100 // Default iterations
			}
			
			for i := 0; i < iterations; i++ {
				testData := test.Generator(tf.rand)
				if !test.Property(testData) {
					tf.t.Fatalf("Property %s failed on iteration %d with data: %+v", test.Name, i+1, testData)
				}
			}
		})
	}
}

// BenchmarkFunction represents a benchmark function.
type BenchmarkFunction func(*TestFramework) error

// Benchmark runs a benchmark and reports timing.
func (tf *TestFramework) Benchmark(name string, fn BenchmarkFunction) {
	tf.t.Helper()
	
	start := time.Now()
	err := fn(tf)
	duration := time.Since(start)
	
	if err != nil {
		tf.t.Fatalf("Benchmark %s failed: %v", name, err)
	}
	
	tf.t.Logf("Benchmark %s completed in %v", name, duration)
}

// Parallel runs multiple test functions in parallel and waits for all to complete.
func (tf *TestFramework) Parallel(fns ...func(*TestFramework)) {
	tf.t.Helper()
	
	done := make(chan error, len(fns))
	
	for i, fn := range fns {
		go func(index int, testFn func(*TestFramework)) {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("goroutine %d panicked: %v", index, r)
					return
				}
				done <- nil
			}()
			
			testFn(tf)
		}(i, fn)
	}
	
	var errors []error
	for i := 0; i < len(fns); i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		tf.t.Fatalf("Parallel execution failed with errors: %v", errors)
	}
}

// Eventually waits for a condition to become true within the timeout.
func (tf *TestFramework) Eventually(condition func() bool, interval time.Duration, msg ...string) {
	tf.t.Helper()
	
	ctx, cancel := context.WithTimeout(tf.ctx, tf.timeout)
	defer cancel()
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		if condition() {
			return
		}
		
		select {
		case <-ctx.Done():
			if len(msg) > 0 {
				tf.t.Fatalf("%s: condition was not met within timeout", strings.Join(msg, " "))
			} else {
				tf.t.Fatal("condition was not met within timeout")
			}
		case <-ticker.C:
			continue
		}
	}
}