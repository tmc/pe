package performance

import (
	"context"
	"sync"
)

// Pool manages reusable values.
type Pool[T any] struct {
	pool sync.Pool
}

// NewPool creates a pool backed by newFn.
func NewPool[T any](newFn func() T) *Pool[T] {
	return &Pool[T]{pool: sync.Pool{New: func() any { return newFn() }}}
}

// Get returns a value from the pool.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns a value to the pool.
func (p *Pool[T]) Put(v T) {
	p.pool.Put(v)
}

// ConnectionPool bounds concurrent connection use.
type ConnectionPool struct {
	slots chan struct{}
}

// NewConnectionPool creates a pool with size slots.
func NewConnectionPool(size int) *ConnectionPool {
	if size <= 0 {
		size = 1
	}
	return &ConnectionPool{slots: make(chan struct{}, size)}
}

// Acquire reserves a slot until the returned release function is called.
func (p *ConnectionPool) Acquire(ctx context.Context) (func(), error) {
	select {
	case p.slots <- struct{}{}:
		return func() { <-p.slots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
