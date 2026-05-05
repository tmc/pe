package errors

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Operation is work that may be retried or recovered.
type Operation func(context.Context) error

// RecoveryStrategy executes an operation with a recovery policy.
type RecoveryStrategy interface {
	Recover(context.Context, Operation) error
}

// RecoveryMetrics records recovery activity.
type RecoveryMetrics struct {
	Attempts  int
	Failures  int
	Recovered int
}

// Backoff computes the delay before attempt. Attempt is 1-based.
type Backoff interface {
	Delay(attempt int) time.Duration
}

// ExponentialBackoff grows delays by Factor up to Max.
type ExponentialBackoff struct {
	Base   time.Duration
	Max    time.Duration
	Factor int
}

// Delay returns the delay for attempt.
func (b ExponentialBackoff) Delay(attempt int) time.Duration {
	if attempt <= 1 {
		return 0
	}
	base := b.Base
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	factor := b.Factor
	if factor <= 1 {
		factor = 2
	}
	delay := base
	for i := 2; i < attempt; i++ {
		delay *= time.Duration(factor)
		if b.Max > 0 && delay > b.Max {
			return b.Max
		}
	}
	if b.Max > 0 && delay > b.Max {
		return b.Max
	}
	return delay
}

// RetryStrategy retries retryable failures.
type RetryStrategy struct {
	MaxAttempts int
	Backoff     Backoff
	Sleep       func(context.Context, time.Duration) error
	metrics     RecoveryMetrics
}

// NewRetryStrategy returns a retry strategy.
func NewRetryStrategy(maxAttempts int, backoff Backoff) *RetryStrategy {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return &RetryStrategy{MaxAttempts: maxAttempts, Backoff: backoff}
}

// Recover runs op until it succeeds, exhausts attempts, or sees a non-retryable error.
func (s *RetryStrategy) Recover(ctx context.Context, op Operation) error {
	var last error
	for attempt := 1; attempt <= s.MaxAttempts; attempt++ {
		if delay := backoffDelay(s.Backoff, attempt); delay > 0 {
			if err := sleep(ctx, s.Sleep, delay); err != nil {
				return err
			}
		}
		s.metrics.Attempts++
		err := op(ctx)
		if err == nil {
			if attempt > 1 {
				s.metrics.Recovered++
			}
			return nil
		}
		last = err
		s.metrics.Failures++
		if !IsRetryable(err) {
			return err
		}
	}
	return last
}

// Metrics returns a copy of the strategy metrics.
func (s *RetryStrategy) Metrics() RecoveryMetrics {
	return s.metrics
}

// FallbackStrategy runs Fallback if Primary fails.
type FallbackStrategy struct {
	Primary  Operation
	Fallback Operation
}

// Recover runs the primary operation and falls back on failure.
func (s FallbackStrategy) Recover(ctx context.Context, _ Operation) error {
	if s.Primary == nil || s.Fallback == nil {
		return fmt.Errorf("fallback strategy requires primary and fallback operations")
	}
	if err := s.Primary(ctx); err == nil {
		return nil
	}
	return s.Fallback(ctx)
}

// CircuitBreaker rejects work after repeated failures until the reset timeout.
type CircuitBreaker struct {
	FailureThreshold int
	ResetAfter       time.Duration

	mu       sync.Mutex
	failures int
	openedAt time.Time
}

// NewCircuitBreaker returns a circuit breaker.
func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 1
	}
	return &CircuitBreaker{FailureThreshold: threshold, ResetAfter: resetAfter}
}

// Recover runs op when the circuit is closed.
func (b *CircuitBreaker) Recover(ctx context.Context, op Operation) error {
	b.mu.Lock()
	if b.openedAt.IsZero() {
		// closed
	} else if b.ResetAfter <= 0 || time.Since(b.openedAt) < b.ResetAfter {
		b.mu.Unlock()
		return New(ErrCodeProviderUnavailable, "circuit breaker is open").WithComponent("recovery")
	} else {
		b.failures = 0
		b.openedAt = time.Time{}
	}
	b.mu.Unlock()

	err := op(ctx)
	b.mu.Lock()
	defer b.mu.Unlock()
	if err == nil {
		b.failures = 0
		return nil
	}
	b.failures++
	if b.failures >= b.FailureThreshold {
		b.openedAt = time.Now()
	}
	return err
}

func backoffDelay(backoff Backoff, attempt int) time.Duration {
	if backoff == nil {
		return 0
	}
	return backoff.Delay(attempt)
}

func sleep(ctx context.Context, sleeper func(context.Context, time.Duration) error, delay time.Duration) error {
	if sleeper != nil {
		return sleeper(ctx, delay)
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
