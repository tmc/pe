package testing

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens   int
	capacity int
	refill   int
	interval time.Duration
	lastRef  time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity, refill int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:   capacity,
		capacity: capacity,
		refill:   refill,
		interval: interval,
		lastRef:  time.Now(),
	}
}

// Allow checks if a request is allowed under current rate limits
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRef)

	// Add tokens based on elapsed time
	if elapsed >= rl.interval {
		tokensToAdd := int(elapsed/rl.interval) * rl.refill
		rl.tokens = min(rl.capacity, rl.tokens+tokensToAdd)
		rl.lastRef = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Wait blocks until a request can proceed
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond * 10):
			continue
		}
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
	Delay     time.Duration
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}
}

// RetryWithPolicy executes a function with retry logic
func RetryWithPolicy(ctx context.Context, policy *RetryPolicy, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if retryErr, ok := err.(*RetryableError); ok {
			if !retryErr.Retryable {
				return err
			}
		}

		// Don't sleep on last attempt
		if attempt == policy.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := time.Duration(float64(policy.BaseDelay) * power(policy.Multiplier, float64(attempt)))
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}

		// Add jitter if enabled
		if policy.Jitter {
			// Add 0-50% random jitter to delay (range: 0.5 to 1.0)
		jitterFactor := 0.5 + 0.5*rand.Float64()
		delay = time.Duration(float64(delay) * jitterFactor)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return lastErr
}

// Helper function for power calculation
func power(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Tests for Rate Limiter

func TestRateLimiter_Basic(t *testing.T) {
	// Create rate limiter: 2 requests per 100ms
	rl := NewRateLimiter(2, 2, 100*time.Millisecond)

	// First two requests should be allowed
	if !rl.Allow() {
		t.Error("First request should be allowed")
	}
	if !rl.Allow() {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if rl.Allow() {
		t.Error("Third request should be denied")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create rate limiter: 1 request per 50ms
	rl := NewRateLimiter(1, 1, 50*time.Millisecond)

	// Use initial token
	if !rl.Allow() {
		t.Error("First request should be allowed")
	}

	// Should be denied immediately
	if rl.Allow() {
		t.Error("Second request should be denied")
	}

	// Wait for refill
	time.Sleep(60 * time.Millisecond)

	// Should be allowed after refill
	if !rl.Allow() {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	// Create rate limiter: 10 requests per 100ms
	rl := NewRateLimiter(10, 5, 100*time.Millisecond)

	const numGoroutines = 20
	const requestsPerGoroutine = 5

	allowed := make(chan bool, numGoroutines*requestsPerGoroutine)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				allowed <- rl.Allow()
				time.Sleep(time.Millisecond) // Small delay between requests
			}
		}()
	}

	wg.Wait()
	close(allowed)

	// Count allowed and denied requests
	allowedCount := 0
	deniedCount := 0
	for result := range allowed {
		if result {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	t.Logf("Allowed: %d, Denied: %d", allowedCount, deniedCount)

	// Should have some allowed and some denied
	if allowedCount == 0 {
		t.Error("Expected some requests to be allowed")
	}
	if deniedCount == 0 {
		t.Error("Expected some requests to be denied")
	}

	// Total should match
	if allowedCount+deniedCount != numGoroutines*requestsPerGoroutine {
		t.Errorf("Expected %d total requests, got %d", numGoroutines*requestsPerGoroutine, allowedCount+deniedCount)
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	// Create rate limiter: 1 request per 50ms
	rl := NewRateLimiter(1, 1, 50*time.Millisecond)

	// Use initial token
	if !rl.Allow() {
		t.Error("First request should be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait should succeed: %v", err)
	}

	// Should have waited for refill
	if elapsed < 40*time.Millisecond {
		t.Errorf("Should have waited for refill, elapsed: %v", elapsed)
	}
}

func TestRateLimiter_WaitTimeout(t *testing.T) {
	// Create rate limiter: 1 request per 1 second (very slow refill)
	rl := NewRateLimiter(1, 1, time.Second)

	// Use initial token
	if !rl.Allow() {
		t.Error("First request should be allowed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("Wait should timeout")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// Tests for Retry Logic

func TestRetryWithPolicy_Success(t *testing.T) {
	policy := DefaultRetryPolicy()
	attempts := 0

	err := RetryWithPolicy(context.Background(), policy, func() error {
		attempts++
		return nil // Success on first try
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithPolicy_EventualSuccess(t *testing.T) {
	policy := DefaultRetryPolicy()
	attempts := 0

	err := RetryWithPolicy(context.Background(), policy, func() error {
		attempts++
		if attempts < 3 {
			return &RetryableError{
				Err:       NewError("temporary failure"),
				Retryable: true,
			}
		}
		return nil // Success on third try
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithPolicy_MaxAttempts(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts: 2,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      false,
	}
	attempts := 0

	err := RetryWithPolicy(context.Background(), policy, func() error {
		attempts++
		return &RetryableError{
			Err:       NewError("persistent failure"),
			Retryable: true,
		}
	})

	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempts != policy.MaxAttempts {
		t.Errorf("Expected %d attempts, got %d", policy.MaxAttempts, attempts)
	}
}

func TestRetryWithPolicy_NonRetryableError(t *testing.T) {
	policy := DefaultRetryPolicy()
	attempts := 0

	err := RetryWithPolicy(context.Background(), policy, func() error {
		attempts++
		return &RetryableError{
			Err:       NewError("non-retryable failure"),
			Retryable: false,
		}
	})

	if err == nil {
		t.Error("Expected error for non-retryable failure")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetryWithPolicy_ContextCancellation(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts: 10,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    time.Second,
		Multiplier:  2.0,
		Jitter:      false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	attempts := 0
	start := time.Now()

	err := RetryWithPolicy(ctx, policy, func() error {
		attempts++
		return &RetryableError{
			Err:       NewError("failure"),
			Retryable: true,
		}
	})

	elapsed := time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}

	// Should have attempted at least once, but not all 10 times
	if attempts < 1 || attempts >= policy.MaxAttempts {
		t.Errorf("Expected 1 <= attempts < %d, got %d", policy.MaxAttempts, attempts)
	}

	// Should have been cancelled around the timeout
	if elapsed < 140*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected cancellation around 150ms, took %v", elapsed)
	}
}

func TestRetryWithPolicy_BackoffTiming(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    time.Second,
		Multiplier:  2.0,
		Jitter:      false, // No jitter for predictable timing
	}

	attempts := 0
	start := time.Now()
	var attemptTimes []time.Time

	RetryWithPolicy(context.Background(), policy, func() error {
		attempts++
		attemptTimes = append(attemptTimes, time.Now())
		return &RetryableError{
			Err:       NewError("failure"),
			Retryable: true,
		}
	})

	if len(attemptTimes) != 3 {
		t.Fatalf("Expected 3 attempts, got %d", len(attemptTimes))
	}

	// Check delays between attempts
	delay1 := attemptTimes[1].Sub(attemptTimes[0])
	delay2 := attemptTimes[2].Sub(attemptTimes[1])

	// First delay should be ~50ms
	if delay1 < 40*time.Millisecond || delay1 > 70*time.Millisecond {
		t.Errorf("Expected first delay ~50ms, got %v", delay1)
	}

	// Second delay should be ~100ms (50ms * 2)
	if delay2 < 90*time.Millisecond || delay2 > 120*time.Millisecond {
		t.Errorf("Expected second delay ~100ms, got %v", delay2)
	}

	totalElapsed := time.Since(start)
	expectedMin := 150 * time.Millisecond // 50ms + 100ms
	if totalElapsed < expectedMin {
		t.Errorf("Expected total time >= %v, got %v", expectedMin, totalElapsed)
	}
}

func TestRetryWithPolicy_Jitter(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    time.Second,
		Multiplier:  1.0, // No multiplication to isolate jitter
		Jitter:      true,
	}

	var delays []time.Duration
	for i := 0; i < 10; i++ {
		attempts := 0
		var attemptTimes []time.Time

		RetryWithPolicy(context.Background(), policy, func() error {
			attempts++
			attemptTimes = append(attemptTimes, time.Now())
			if attempts < 2 {
				return &RetryableError{
					Err:       NewError("failure"),
					Retryable: true,
				}
			}
			return nil
		})

		if len(attemptTimes) >= 2 {
			delay := attemptTimes[1].Sub(attemptTimes[0])
			delays = append(delays, delay)
		}
	}

	if len(delays) < 5 {
		t.Fatalf("Expected at least 5 delay measurements, got %d", len(delays))
	}

	// With jitter, delays should vary
	allSame := true
	first := delays[0]
	for _, delay := range delays[1:] {
		if abs(delay-first) > 10*time.Millisecond { // Allow some tolerance
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Expected jitter to cause delay variation")
	}
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// Benchmark Tests

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000, 100, 10*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkRetryWithPolicy(b *testing.B) {
	policy := DefaultRetryPolicy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RetryWithPolicy(context.Background(), policy, func() error {
			return nil // Always succeed
		})
	}
}

// Mock error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// NewError creates a new test error
func NewError(message string) error {
	return &testError{message: message}
}