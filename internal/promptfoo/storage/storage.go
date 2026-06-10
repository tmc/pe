// Package storage defines the persistence interfaces pe uses for cached
// inference and evaluation run results, with filesystem implementations.
//
// The design is layered so callers can swap in a database or any fs.FS:
//
//   - [Store] is the byte layer: an [io/fs.FS] for reads plus Put and Remove
//     for writes. A directory, an in-memory map, or a DB blob table can all
//     implement it.
//   - [Cache] is a typed request/response cache built on a Store.
//   - [RunStore] is a typed evaluation-result store built on a Store.
//
// Keys are slash-separated names, like fs.FS paths, so the same key works for
// a directory layout and a key/value store.
package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"sort"
	"strings"

	"github.com/tmc/pe/internal/promptfoo"
)

// Store is the byte-level persistence interface. It is an fs.FS (so reads,
// listing, and globbing come from the standard library for free) extended with
// the two write operations a cache needs. Implementations must accept
// slash-separated names and create intermediate structure as needed.
type Store interface {
	fs.FS
	// Put writes data under name, replacing any existing value.
	Put(name string, data []byte) error
	// Remove deletes name. Removing a missing name is not an error.
	Remove(name string) error
}

// readFile reads a whole entry from a Store via its fs.FS view.
func readFile(s Store, name string) ([]byte, error) {
	return fs.ReadFile(s, name)
}

// Cache is a typed cache of provider responses keyed by an opaque string
// (typically a content hash of provider+prompt+vars). It persists each entry
// as JSON under "<key>.json" in the backing Store, so any Store implementation
// yields a persistent cache.
type Cache struct {
	store Store
}

// NewCache returns a Cache backed by store.
func NewCache(store Store) *Cache {
	return &Cache{store: store}
}

func cacheName(key string) string {
	return key + ".json"
}

// Get returns the cached response for key. The boolean is false when the key is
// absent. A corrupt entry is reported as an error.
func (c *Cache) Get(key string) (*promptfoo.ProviderResponse, bool, error) {
	data, err := readFile(c.store, cacheName(key))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var response promptfoo.ProviderResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, false, err
	}
	return &response, true, nil
}

// Set stores response under key.
func (c *Cache) Set(key string, response *promptfoo.ProviderResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return c.store.Put(cacheName(key), data)
}

// Delete removes the cached entry for key.
func (c *Cache) Delete(key string) error {
	return c.store.Remove(cacheName(key))
}

// RunStore persists evaluation results by their EvalID. Each run is stored as
// JSON under "<evalID>.json" in the backing Store.
type RunStore struct {
	store Store
}

// NewRunStore returns a RunStore backed by store.
func NewRunStore(store Store) *RunStore {
	return &RunStore{store: store}
}

func runName(evalID string) string {
	return evalID + ".json"
}

// Save persists an evaluation result under its EvalID.
func (r *RunStore) Save(result promptfoo.EvaluationResult) error {
	if result.EvalID == "" {
		return errors.New("run result has no EvalID")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return r.store.Put(runName(result.EvalID), data)
}

// Load reads the evaluation result stored under evalID.
func (r *RunStore) Load(evalID string) (promptfoo.EvaluationResult, error) {
	var result promptfoo.EvaluationResult
	data, err := readFile(r.store, runName(evalID))
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// List returns the EvalIDs of all stored runs, sorted. A store with no runs
// yields an empty slice and no error.
func (r *RunStore) List() ([]string, error) {
	entries, err := fs.ReadDir(r.store, ".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if id, ok := strings.CutSuffix(name, ".json"); ok {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// Delete removes the run stored under evalID.
func (r *RunStore) Delete(evalID string) error {
	return r.store.Remove(runName(evalID))
}
