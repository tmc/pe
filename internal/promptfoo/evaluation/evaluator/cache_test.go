package evaluator

import (
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
	"github.com/tmc/pe/internal/promptfoo/storage"
)

// TestResponseCachePersists verifies a put is visible to a fresh cache built
// over the same Store, i.e. responses survive across runs.
func TestResponseCachePersists(t *testing.T) {
	store := storage.NewMemStore()

	c1 := newPersistentResponseCache(store)
	c1.put("k1", &promptfoo.ProviderResponse{Output: "cached", Cost: 0.02})

	// A new cache over the same store starts cold in memory but reads the
	// persisted entry on miss.
	c2 := newPersistentResponseCache(store)
	got, ok := c2.get("k1")
	if !ok {
		t.Fatal("expected persisted cache hit on a fresh cache")
	}
	if got.Output != "cached" || got.Cost != 0.02 {
		t.Fatalf("persisted response = %+v, want Output=cached Cost=0.02", got)
	}
}

// TestResponseCacheInMemoryOnly confirms a nil store keeps entries only in
// memory (no persistence across cache instances).
func TestResponseCacheInMemoryOnly(t *testing.T) {
	c1 := newPersistentResponseCache(nil)
	c1.put("k", &promptfoo.ProviderResponse{Output: "x"})
	if _, ok := c1.get("k"); !ok {
		t.Fatal("in-memory cache should hit within the same instance")
	}

	c2 := newPersistentResponseCache(nil)
	if _, ok := c2.get("k"); ok {
		t.Fatal("in-memory-only cache must not share entries across instances")
	}
}

// TestResponseCacheReturnsCopies ensures callers cannot mutate cached entries.
func TestResponseCacheReturnsCopies(t *testing.T) {
	c := newPersistentResponseCache(storage.NewMemStore())
	c.put("k", &promptfoo.ProviderResponse{Output: "orig"})
	got, _ := c.get("k")
	got.Output = "mutated"
	again, _ := c.get("k")
	if again.Output != "orig" {
		t.Fatalf("cache returned a shared pointer; got mutated value %q", again.Output)
	}
}
