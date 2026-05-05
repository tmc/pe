package errors

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryStrategyRecoversRetryableError(t *testing.T) {
	attempts := 0
	strategy := NewRetryStrategy(3, ExponentialBackoff{Base: time.Nanosecond})
	strategy.Sleep = func(context.Context, time.Duration) error { return nil }

	err := strategy.Recover(context.Background(), func(context.Context) error {
		attempts++
		if attempts == 1 {
			err := New(ErrCodeNetworkTimeout, "timeout")
			err.Retryable = true
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if strategy.Metrics().Recovered != 1 {
		t.Fatalf("Recovered = %d, want 1", strategy.Metrics().Recovered)
	}
}

func TestRetryStrategyStopsOnNonRetryableError(t *testing.T) {
	attempts := 0
	strategy := NewRetryStrategy(3, nil)
	err := strategy.Recover(context.Background(), func(context.Context) error {
		attempts++
		return New(ErrCodeInvalidInput, "bad input")
	})
	if err == nil {
		t.Fatal("Recover succeeded, want error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestFallbackStrategy(t *testing.T) {
	fallbackCalled := false
	strategy := FallbackStrategy{
		Primary: func(context.Context) error {
			return errors.New("primary failed")
		},
		Fallback: func(context.Context) error {
			fallbackCalled = true
			return nil
		},
	}
	if err := strategy.Recover(context.Background(), nil); err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if !fallbackCalled {
		t.Fatal("fallback was not called")
	}
}

func TestCircuitBreakerOpens(t *testing.T) {
	breaker := NewCircuitBreaker(1, time.Hour)
	err := breaker.Recover(context.Background(), func(context.Context) error {
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("first call succeeded, want error")
	}
	err = breaker.Recover(context.Background(), func(context.Context) error {
		t.Fatal("operation should not run while circuit is open")
		return nil
	})
	if !IsCode(err, ErrCodeProviderUnavailable) {
		t.Fatalf("err = %v, want provider unavailable", err)
	}
}

func TestExponentialBackoffCapsDelay(t *testing.T) {
	backoff := ExponentialBackoff{Base: time.Second, Max: 3 * time.Second, Factor: 2}
	if got := backoff.Delay(4); got != 3*time.Second {
		t.Fatalf("Delay(4) = %v, want 3s", got)
	}
}
