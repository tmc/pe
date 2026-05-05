package performance

import (
	"context"
	"testing"
)

func TestCacheGetOrCompute(t *testing.T) {
	cache := NewCache[string, int]()
	calls := 0
	v, err := cache.GetOrCompute("x", func() (int, error) {
		calls++
		return 42, nil
	})
	if err != nil || v != 42 {
		t.Fatalf("GetOrCompute = %d, %v", v, err)
	}
	v, err = cache.GetOrCompute("x", func() (int, error) {
		calls++
		return 0, nil
	})
	if err != nil || v != 42 || calls != 1 {
		t.Fatalf("cached value = %d, calls = %d, err = %v", v, calls, err)
	}
}

func TestPool(t *testing.T) {
	pool := NewPool(func() []byte { return make([]byte, 0, 16) })
	buf := pool.Get()
	buf = append(buf, "pe"...)
	pool.Put(buf[:0])
	if got := cap(pool.Get()); got < 16 {
		t.Fatalf("pooled buffer cap = %d, want >= 16", got)
	}
}

func TestConnectionPool(t *testing.T) {
	pool := NewConnectionPool(1)
	release, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := pool.Acquire(ctx); err == nil {
		t.Fatal("Acquire with canceled context succeeded")
	}
	release()
}
