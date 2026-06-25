package testing

import (
	"context"
	"math/rand"
	"sync/atomic"
	stdtesting "testing"
	"time"
)

func TestFrameworkAccessorsAndAssertions(t *stdtesting.T) {
	ctx := context.WithValue(context.Background(), contextKey("k"), "v")
	tf := NewTestFramework(t).WithContext(ctx).WithTimeout(time.Second)
	if tf.Context() != ctx {
		t.Fatalf("context not set")
	}
	if tf.T() != t {
		t.Fatalf("testing handle not set")
	}
	if tf.Rand() == nil {
		t.Fatalf("nil random source")
	}

	tf.AssertNoError(nil)
	tf.AssertError(context.Canceled)
	tf.AssertEqual("a", "a")
	tf.AssertNotEqual("a", "b")
	tf.AssertContains("abcdef", "bcd")
	tf.AssertNotEmpty("x")
	tf.AssertTrue(true)
	tf.AssertFalse(false)
	tf.AssertPanics(func() { panic("boom") })
	tf.AssertNotPanics(func() {})
}

func TestFrameworkRunPropertyBenchmarkParallelEventually(t *stdtesting.T) {
	tf := NewTestFramework(t).WithTimeout(200 * time.Millisecond)
	runCalled := false
	tf.Run("sub", func(sub *TestFramework) {
		runCalled = true
		if sub.Context() == nil {
			t.Fatal("nil sub context")
		}
	})
	if !runCalled {
		t.Fatal("Run did not call subtest")
	}

	properties := []PropertyTest{{
		Name:       "positive ints",
		Iterations: 3,
		Generator:  func(_ *rand.Rand) interface{} { return 1 },
	}}
	_ = properties
	tf.RunPropertyTests([]PropertyTest{{
		Name:       "strings are non-empty",
		Iterations: 4,
		Generator:  func(_ *rand.Rand) interface{} { return "x" },
		Property: func(v interface{}) bool {
			return v.(string) != ""
		},
	}})

	tf.Benchmark("noop", func(*TestFramework) error { return nil })

	var count atomic.Int64
	tf.Parallel(
		func(*TestFramework) { count.Add(1) },
		func(*TestFramework) { count.Add(1) },
	)
	if got := count.Load(); got != 2 {
		t.Fatalf("parallel count = %d, want 2", got)
	}

	var ready atomic.Bool
	go func() {
		time.Sleep(10 * time.Millisecond)
		ready.Store(true)
	}()
	tf.Eventually(func() bool { return ready.Load() }, time.Millisecond)
}

func TestTestDataDeterministicGeneration(t *stdtesting.T) {
	first := NewTestData(42)
	second := NewTestData(42)
	if first.RandomString(12) != second.RandomString(12) {
		t.Fatal("RandomString not deterministic for same seed")
	}
	if got := len(NewTestData(1).RandomString(8)); got != 8 {
		t.Fatalf("RandomString length = %d, want 8", got)
	}

	data := NewTestData(7)
	if data.RandomPrompt() == "" {
		t.Fatal("empty random prompt")
	}
	if data.RandomModel() == "" {
		t.Fatal("empty random model")
	}
	req := data.RandomInferenceRequest()
	if req.Prompt == "" || req.Model == "" || req.MaxTokens < 100 || req.MaxTokens > 999 {
		t.Fatalf("request = %#v", req)
	}
	resp := data.RandomResponse("model-x")
	if resp.Model != "model-x" || resp.Content == "" || resp.Metadata["test_data"] != true {
		t.Fatalf("response = %#v", resp)
	}
}

type contextKey string
