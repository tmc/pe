package storage

import (
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

// stores returns the Store implementations under test, so every behavior is
// verified against both the filesystem and in-memory backends.
func stores(t *testing.T) map[string]Store {
	t.Helper()
	return map[string]Store{
		"mem": NewMemStore(),
		"dir": NewDirStore(t.TempDir()),
	}
}

func TestCacheRoundTrip(t *testing.T) {
	for name, store := range stores(t) {
		t.Run(name, func(t *testing.T) {
			c := NewCache(store)

			// Miss on an absent key.
			if _, ok, err := c.Get("abc"); err != nil || ok {
				t.Fatalf("Get(absent) = ok %v err %v, want false nil", ok, err)
			}

			resp := &promptfoo.ProviderResponse{Output: "hello", Cost: 0.01, LatencyMs: 42}
			if err := c.Set("abc", resp); err != nil {
				t.Fatalf("Set: %v", err)
			}

			got, ok, err := c.Get("abc")
			if err != nil || !ok {
				t.Fatalf("Get(present) = ok %v err %v, want true nil", ok, err)
			}
			if got.Output != "hello" || got.Cost != 0.01 || got.LatencyMs != 42 {
				t.Fatalf("round trip mismatch: %+v", got)
			}

			if err := c.Delete("abc"); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			if _, ok, _ := c.Get("abc"); ok {
				t.Fatal("Get after Delete should miss")
			}
		})
	}
}

func TestRunStoreSaveLoadList(t *testing.T) {
	for name, store := range stores(t) {
		t.Run(name, func(t *testing.T) {
			rs := NewRunStore(store)

			// Empty list, not an error.
			ids, err := rs.List()
			if err != nil {
				t.Fatalf("List(empty): %v", err)
			}
			if len(ids) != 0 {
				t.Fatalf("List(empty) = %v, want none", ids)
			}

			a := promptfoo.EvaluationResult{EvalID: "eval-a"}
			b := promptfoo.EvaluationResult{EvalID: "eval-b"}
			if err := rs.Save(a); err != nil {
				t.Fatalf("Save a: %v", err)
			}
			if err := rs.Save(b); err != nil {
				t.Fatalf("Save b: %v", err)
			}

			ids, err = rs.List()
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(ids) != 2 || ids[0] != "eval-a" || ids[1] != "eval-b" {
				t.Fatalf("List = %v, want [eval-a eval-b] sorted", ids)
			}

			loaded, err := rs.Load("eval-a")
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if loaded.EvalID != "eval-a" {
				t.Fatalf("Load EvalID = %q, want eval-a", loaded.EvalID)
			}

			if err := rs.Delete("eval-a"); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			ids, _ = rs.List()
			if len(ids) != 1 || ids[0] != "eval-b" {
				t.Fatalf("List after delete = %v, want [eval-b]", ids)
			}
		})
	}
}

func TestRunStoreRequiresEvalID(t *testing.T) {
	rs := NewRunStore(NewMemStore())
	if err := rs.Save(promptfoo.EvaluationResult{}); err == nil {
		t.Fatal("Save with empty EvalID should error")
	}
}
