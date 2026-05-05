package performance

import "sync"

// Cache is a small concurrent in-memory cache.
type Cache[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewCache creates an empty cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{m: make(map[K]V)}
}

// Get returns a cached value.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}

// Set stores a value.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}

// GetOrCompute returns a cached value or computes and stores it.
func (c *Cache[K, V]) GetOrCompute(key K, compute func() (V, error)) (V, error) {
	if v, ok := c.Get(key); ok {
		return v, nil
	}
	v, err := compute()
	if err != nil {
		var zero V
		return zero, err
	}
	c.Set(key, v)
	return v, nil
}
