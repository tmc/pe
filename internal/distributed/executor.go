package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExecutionMode defines how tasks should be executed
type ExecutionMode string

const (
	// LocalMode executes all tasks locally
	LocalMode ExecutionMode = "local"
	// ClusterMode distributes tasks across multiple nodes
	ClusterMode ExecutionMode = "cluster"
	// HybridMode uses both local and remote resources
	HybridMode ExecutionMode = "hybrid"
)

// Task represents a unit of work to be executed
type Task interface {
	// ID returns a unique identifier for the task
	ID() string
	// Execute runs the task and returns results or error
	Execute(ctx context.Context) (interface{}, error)
	// Priority returns the task priority (higher = more important)
	Priority() int
	// EstimatedDuration returns expected execution time
	EstimatedDuration() time.Duration
}

// Node represents a compute node in the distributed system
type Node struct {
	ID       string
	Address  string
	Capacity int // Number of concurrent tasks
	Load     int // Current number of running tasks
	mu       sync.RWMutex
}

// Executor manages distributed task execution
type Executor struct {
	mode     ExecutionMode
	nodes    []*Node
	taskChan chan Task
	results  sync.Map
	wg       sync.WaitGroup
	mu       sync.RWMutex

	// Configuration
	maxRetries     int
	retryDelay     time.Duration
	loadBalancer   LoadBalancer
	faultTolerance bool
}

// NewExecutor creates a new distributed executor
func NewExecutor(mode ExecutionMode, opts ...ExecutorOption) *Executor {
	e := &Executor{
		mode:           mode,
		nodes:          make([]*Node, 0),
		taskChan:       make(chan Task, 1000),
		maxRetries:     3,
		retryDelay:     time.Second,
		faultTolerance: true,
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Set default load balancer if not provided
	if e.loadBalancer == nil {
		e.loadBalancer = NewRoundRobinBalancer()
	}

	return e
}

// ExecutorOption configures the executor
type ExecutorOption func(*Executor)

// WithMaxRetries sets the maximum number of retries for failed tasks
func WithMaxRetries(n int) ExecutorOption {
	return func(e *Executor) {
		e.maxRetries = n
	}
}

// WithLoadBalancer sets a custom load balancer
func WithLoadBalancer(lb LoadBalancer) ExecutorOption {
	return func(e *Executor) {
		e.loadBalancer = lb
	}
}

// WithFaultTolerance enables/disables fault tolerance
func WithFaultTolerance(enabled bool) ExecutorOption {
	return func(e *Executor) {
		e.faultTolerance = enabled
	}
}

// AddNode adds a compute node to the executor
func (e *Executor) AddNode(node *Node) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nodes = append(e.nodes, node)
}

// RemoveNode removes a compute node from the executor
func (e *Executor) RemoveNode(nodeID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	filtered := make([]*Node, 0, len(e.nodes))
	for _, node := range e.nodes {
		if node.ID != nodeID {
			filtered = append(filtered, node)
		}
	}
	e.nodes = filtered
}

// Submit adds a task to the execution queue
func (e *Executor) Submit(task Task) error {
	select {
	case e.taskChan <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitBatch adds multiple tasks to the execution queue
func (e *Executor) SubmitBatch(tasks []Task) error {
	for _, task := range tasks {
		if err := e.Submit(task); err != nil {
			return fmt.Errorf("failed to submit task %s: %w", task.ID(), err)
		}
	}
	return nil
}

// Start begins processing tasks
func (e *Executor) Start(ctx context.Context) error {
	switch e.mode {
	case LocalMode:
		return e.startLocal(ctx)
	case ClusterMode:
		return e.startCluster(ctx)
	case HybridMode:
		return e.startHybrid(ctx)
	default:
		return fmt.Errorf("unknown execution mode: %s", e.mode)
	}
}

// startLocal runs all tasks on the local machine
func (e *Executor) startLocal(ctx context.Context) error {
	// Determine number of workers based on CPU
	numWorkers := 4 // TODO: Use runtime.NumCPU()

	for i := 0; i < numWorkers; i++ {
		e.wg.Add(1)
		go e.localWorker(ctx, i)
	}

	return nil
}

// localWorker processes tasks locally
func (e *Executor) localWorker(ctx context.Context, id int) {
	defer e.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-e.taskChan:
			if !ok {
				return
			}

			result, err := e.executeWithRetry(ctx, task)
			if err != nil {
				e.results.Store(task.ID(), &TaskResult{
					TaskID: task.ID(),
					Error:  err,
				})
			} else {
				e.results.Store(task.ID(), &TaskResult{
					TaskID: task.ID(),
					Result: result,
				})
			}
		}
	}
}

// executeWithRetry executes a task with retry logic
func (e *Executor) executeWithRetry(ctx context.Context, task Task) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(e.retryDelay * time.Duration(attempt)):
				// Exponential backoff
			}
		}

		result, err := task.Execute(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !e.shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("task failed after %d attempts: %w", e.maxRetries+1, lastErr)
}

// shouldRetry determines if an error is retryable
func (e *Executor) shouldRetry(err error) bool {
	if !e.faultTolerance {
		return false
	}

	// TODO: Implement retry logic based on error type
	// For now, retry all errors
	return true
}

// startCluster runs tasks across multiple nodes
func (e *Executor) startCluster(ctx context.Context) error {
	if len(e.nodes) == 0 {
		return fmt.Errorf("no nodes available for cluster mode")
	}

	// Start task dispatcher
	e.wg.Add(1)
	go e.clusterDispatcher(ctx)

	// Start result collector for each node
	for _, node := range e.nodes {
		e.wg.Add(1)
		go e.nodeResultCollector(ctx, node)
	}

	return nil
}

// startHybrid runs tasks using both local and remote resources
func (e *Executor) startHybrid(ctx context.Context) error {
	// Start local workers
	if err := e.startLocal(ctx); err != nil {
		return fmt.Errorf("failed to start local workers: %w", err)
	}

	// Start cluster workers if nodes are available
	if len(e.nodes) > 0 {
		e.wg.Add(1)
		go e.clusterDispatcher(ctx)

		for _, node := range e.nodes {
			e.wg.Add(1)
			go e.nodeResultCollector(ctx, node)
		}
	}

	return nil
}

// Wait blocks until all tasks are complete
func (e *Executor) Wait() {
	e.wg.Wait()
}

// Shutdown gracefully shuts down the executor
func (e *Executor) Shutdown() error {
	close(e.taskChan)
	e.Wait()
	return nil
}

// GetResult retrieves the result for a specific task
func (e *Executor) GetResult(taskID string) (*TaskResult, bool) {
	result, ok := e.results.Load(taskID)
	if !ok {
		return nil, false
	}
	return result.(*TaskResult), true
}

// GetAllResults retrieves all task results
func (e *Executor) GetAllResults() map[string]*TaskResult {
	results := make(map[string]*TaskResult)
	e.results.Range(func(key, value interface{}) bool {
		results[key.(string)] = value.(*TaskResult)
		return true
	})
	return results
}

// TaskResult holds the result of a task execution
type TaskResult struct {
	TaskID string
	Result interface{}
	Error  error
}

// LoadBalancer distributes tasks across nodes
type LoadBalancer interface {
	// SelectNode chooses a node for task execution
	SelectNode(nodes []*Node, task Task) (*Node, error)
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
	current uint64
	mu      sync.Mutex
}

// NewRoundRobinBalancer creates a new round-robin balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// SelectNode selects the next node in round-robin fashion
func (b *RoundRobinBalancer) SelectNode(nodes []*Node, task Task) (*Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	idx := b.current % uint64(len(nodes))
	b.current++

	return nodes[idx], nil
}

// clusterDispatcher distributes tasks to cluster nodes
func (e *Executor) clusterDispatcher(ctx context.Context) {
	defer e.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-e.taskChan:
			if !ok {
				return
			}

			// Select a node for the task
			e.mu.RLock()
			nodes := make([]*Node, len(e.nodes))
			copy(nodes, e.nodes)
			e.mu.RUnlock()

			node, err := e.loadBalancer.SelectNode(nodes, task)
			if err != nil {
				// No nodes available, store error
				e.results.Store(task.ID(), &TaskResult{
					TaskID: task.ID(),
					Error:  fmt.Errorf("no nodes available: %w", err),
				})
				continue
			}

			// Dispatch task to selected node
			if err := e.dispatchToNode(ctx, task, node); err != nil {
				// Retry with different node or handle error
				e.handleDispatchError(ctx, task, err)
			}
		}
	}
}

// dispatchToNode sends a task to a specific node for execution
func (e *Executor) dispatchToNode(ctx context.Context, task Task, node *Node) error {
	// Update node load
	node.mu.Lock()
	if node.Load >= node.Capacity {
		node.mu.Unlock()
		return fmt.Errorf("node %s at capacity", node.ID)
	}
	node.Load++
	node.mu.Unlock()

	// TODO: Implement actual remote task execution
	// For now, simulate task execution
	go func() {
		defer func() {
			node.mu.Lock()
			node.Load--
			node.mu.Unlock()
		}()

		result, err := e.executeWithRetry(ctx, task)
		if err != nil {
			e.results.Store(task.ID(), &TaskResult{
				TaskID: task.ID(),
				Error:  err,
			})
		} else {
			e.results.Store(task.ID(), &TaskResult{
				TaskID: task.ID(),
				Result: result,
			})
		}
	}()

	return nil
}

// handleDispatchError handles errors when dispatching tasks
func (e *Executor) handleDispatchError(ctx context.Context, task Task, err error) {
	if e.faultTolerance {
		// Try to re-submit the task
		select {
		case e.taskChan <- task:
			// Successfully re-queued
		default:
			// Queue is full, store error
			e.results.Store(task.ID(), &TaskResult{
				TaskID: task.ID(),
				Error:  fmt.Errorf("dispatch failed and queue full: %w", err),
			})
		}
	} else {
		e.results.Store(task.ID(), &TaskResult{
			TaskID: task.ID(),
			Error:  err,
		})
	}
}

// nodeResultCollector collects results from a specific node
func (e *Executor) nodeResultCollector(ctx context.Context, node *Node) {
	defer e.wg.Done()

	// TODO: Implement actual result collection from remote nodes
	// This will involve listening for results sent back from the node

	<-ctx.Done()
}
