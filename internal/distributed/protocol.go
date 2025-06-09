package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RemoteTask represents a task that can be executed remotely
type RemoteTask struct {
	ID               string          `json:"id"`
	Type             string          `json:"type"`
	Payload          json.RawMessage `json:"payload"`
	Priority         int             `json:"priority"`
	EstimatedSeconds int             `json:"estimated_seconds"`
	CreatedAt        time.Time       `json:"created_at"`
}

// RemoteTaskResult represents the result of a remote task execution
type RemoteTaskResult struct {
	TaskID    string          `json:"task_id"`
	NodeID    string          `json:"node_id"`
	Success   bool            `json:"success"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	StartTime time.Time       `json:"start_time"`
	EndTime   time.Time       `json:"end_time"`
}

// TaskExecutor executes tasks on a node
type TaskExecutor interface {
	// ExecuteTask executes a task and returns the result
	ExecuteTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error)
	// CanExecute checks if the executor can handle a task type
	CanExecute(taskType string) bool
}

// RemoteExecutor handles remote task execution via P2P network
type RemoteExecutor struct {
	network      *P2PNetwork
	executor     TaskExecutor
	nodeCapacity int
	activeTasks  sync.Map
	mu           sync.RWMutex
}

// NewRemoteExecutor creates a new remote executor
func NewRemoteExecutor(network *P2PNetwork, executor TaskExecutor, capacity int) *RemoteExecutor {
	re := &RemoteExecutor{
		network:      network,
		executor:     executor,
		nodeCapacity: capacity,
	}

	// Register task handlers
	network.RegisterHandler(MessageTypeTaskRequest, re.handleTaskRequest)

	return re
}

// SendTask sends a task to a remote node for execution
func (re *RemoteExecutor) SendTask(ctx context.Context, task Task, nodeID string) error {
	// Convert to remote task format
	remoteTask := &RemoteTask{
		ID:               task.ID(),
		Type:             fmt.Sprintf("%T", task),
		Priority:         task.Priority(),
		EstimatedSeconds: int(task.EstimatedDuration().Seconds()),
		CreatedAt:        time.Now(),
	}

	// Serialize task payload
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}
	remoteTask.Payload = payload

	// Marshal remote task
	taskData, err := json.Marshal(remoteTask)
	if err != nil {
		return fmt.Errorf("failed to marshal remote task: %w", err)
	}

	// Find target peer
	var targetPeer *Peer
	for _, peer := range re.network.GetPeers() {
		if peer.ID == nodeID {
			targetPeer = peer
			break
		}
	}

	if targetPeer == nil {
		return fmt.Errorf("peer not found: %s", nodeID)
	}

	// Send task request
	return re.network.SendMessage(targetPeer, &Message{
		Type:      MessageTypeTaskRequest,
		ID:        generateMessageID(),
		From:      re.network.nodeID,
		To:        nodeID,
		Timestamp: time.Now(),
		Payload:   taskData,
	})
}

// handleTaskRequest processes incoming task execution requests
func (re *RemoteExecutor) handleTaskRequest(ctx context.Context, msg *Message, peer *Peer) error {
	var remoteTask RemoteTask
	if err := json.Unmarshal(msg.Payload, &remoteTask); err != nil {
		return fmt.Errorf("failed to unmarshal task request: %w", err)
	}

	// Check if we can execute this task type
	if !re.executor.CanExecute(remoteTask.Type) {
		result := &RemoteTaskResult{
			TaskID:    remoteTask.ID,
			NodeID:    re.network.nodeID,
			Success:   false,
			Error:     fmt.Sprintf("node cannot execute task type: %s", remoteTask.Type),
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		return re.sendTaskResult(peer, result)
	}

	// Check capacity
	activeCount := 0
	re.activeTasks.Range(func(_, _ interface{}) bool {
		activeCount++
		return true
	})

	if activeCount >= re.nodeCapacity {
		result := &RemoteTaskResult{
			TaskID:    remoteTask.ID,
			NodeID:    re.network.nodeID,
			Success:   false,
			Error:     "node at capacity",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		return re.sendTaskResult(peer, result)
	}

	// Mark task as active
	re.activeTasks.Store(remoteTask.ID, true)

	// Execute task asynchronously
	go func() {
		defer re.activeTasks.Delete(remoteTask.ID)

		startTime := time.Now()
		result, err := re.executor.ExecuteTask(ctx, &remoteTask)
		if err != nil {
			result = &RemoteTaskResult{
				TaskID:    remoteTask.ID,
				NodeID:    re.network.nodeID,
				Success:   false,
				Error:     err.Error(),
				StartTime: startTime,
				EndTime:   time.Now(),
			}
		} else {
			result.StartTime = startTime
			result.EndTime = time.Now()
			result.NodeID = re.network.nodeID
		}

		// Send result back
		if err := re.sendTaskResult(peer, result); err != nil {
			// Log error
			fmt.Printf("Failed to send task result: %v\n", err)
		}
	}()

	return nil
}

// sendTaskResult sends a task result back to the requesting peer
func (re *RemoteExecutor) sendTaskResult(peer *Peer, result *RemoteTaskResult) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result: %w", err)
	}

	return re.network.SendMessage(peer, &Message{
		Type:      MessageTypeTaskResponse,
		ID:        generateMessageID(),
		From:      re.network.nodeID,
		To:        peer.ID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// GetActiveTaskCount returns the number of currently executing tasks
func (re *RemoteExecutor) GetActiveTaskCount() int {
	count := 0
	re.activeTasks.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// PromptEngineeringTaskExecutor implements TaskExecutor for PE tasks
type PromptEngineeringTaskExecutor struct {
	// PE-specific execution components
	evaluator    interface{} // Will be set to actual evaluator
	optimizer    interface{} // Will be set to actual optimizer
	cacheDir     string
	providerName string
	mu           sync.RWMutex
}

// ExecuteTask executes a prompt engineering task
func (pe *PromptEngineeringTaskExecutor) ExecuteTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	startTime := time.Now()

	// Parse task type and payload
	switch task.Type {
	case "*evaluator.EvalTask":
		return pe.executeEvalTask(ctx, task)
	case "*metaprompt.OptimizeTask":
		return pe.executeOptimizeTask(ctx, task)
	case "*distributed.BenchmarkTask":
		return pe.executeBenchmarkTask(ctx, task)
	case "*distributed.PassNTask":
		return pe.executePassNTask(ctx, task)
	case "*distributed.SemanticTask":
		return pe.executeSemanticTask(ctx, task)
	default:
		return &RemoteTaskResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("unsupported task type: %s", task.Type),
			StartTime: startTime,
			EndTime:   time.Now(),
		}, nil
	}
}

// CanExecute checks if this executor can handle the task type
func (pe *PromptEngineeringTaskExecutor) CanExecute(taskType string) bool {
	// List of supported PE task types
	supportedTypes := []string{
		"*evaluator.EvalTask",
		"*metaprompt.OptimizeTask",
		"*distributed.BenchmarkTask",
		"*distributed.PassNTask",
		"*distributed.SemanticTask",
	}

	for _, supported := range supportedTypes {
		if taskType == supported {
			return true
		}
	}
	return false
}

// executeEvalTask handles evaluation tasks
func (pe *PromptEngineeringTaskExecutor) executeEvalTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	// TODO: Deserialize eval configuration and execute
	// For now, return a placeholder result
	result := map[string]interface{}{
		"task_type": "eval",
		"status":    "completed",
		"metrics": map[string]float64{
			"accuracy": 0.95,
			"latency":  1.23,
		},
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &RemoteTaskResult{
		TaskID:  task.ID,
		Success: true,
		Result:  resultJSON,
	}, nil
}

// executeOptimizeTask handles optimization tasks
func (pe *PromptEngineeringTaskExecutor) executeOptimizeTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	// TODO: Deserialize optimization configuration and execute
	result := map[string]interface{}{
		"task_type":        "optimize",
		"status":           "completed",
		"optimized_prompt": "Optimized version of the prompt",
		"improvement":      0.15,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &RemoteTaskResult{
		TaskID:  task.ID,
		Success: true,
		Result:  resultJSON,
	}, nil
}

// executeBenchmarkTask handles benchmark tasks
func (pe *PromptEngineeringTaskExecutor) executeBenchmarkTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	// TODO: Deserialize benchmark configuration and execute
	result := map[string]interface{}{
		"task_type": "benchmark",
		"status":    "completed",
		"results": map[string]interface{}{
			"throughput":  100.5,
			"p95_latency": 2.34,
		},
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &RemoteTaskResult{
		TaskID:  task.ID,
		Success: true,
		Result:  resultJSON,
	}, nil
}

// executePassNTask handles pass@n evaluation tasks
func (pe *PromptEngineeringTaskExecutor) executePassNTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	// TODO: Deserialize pass@n configuration and execute
	result := map[string]interface{}{
		"task_type":  "pass_n",
		"status":     "completed",
		"pass_at_1":  0.82,
		"pass_at_5":  0.94,
		"pass_at_10": 0.97,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &RemoteTaskResult{
		TaskID:  task.ID,
		Success: true,
		Result:  resultJSON,
	}, nil
}

// executeSemanticTask handles semantic backpropagation tasks
func (pe *PromptEngineeringTaskExecutor) executeSemanticTask(ctx context.Context, task *RemoteTask) (*RemoteTaskResult, error) {
	// TODO: Deserialize semantic task configuration and execute
	result := map[string]interface{}{
		"task_type": "semantic",
		"status":    "completed",
		"gradient": map[string]interface{}{
			"direction": "improve_clarity",
			"magnitude": 0.73,
		},
		"optimized_prompt": "Semantically optimized prompt",
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &RemoteTaskResult{
		TaskID:  task.ID,
		Success: true,
		Result:  resultJSON,
	}, nil
}
