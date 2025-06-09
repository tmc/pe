package distributed

import (
	"context"
	"testing"
	"time"
)

// SimpleTask implements the Task interface for testing
type SimpleTask struct {
	id       string
	result   interface{}
	err      error
	duration time.Duration
}

func (t *SimpleTask) ID() string {
	return t.id
}

func (t *SimpleTask) Execute(ctx context.Context) (interface{}, error) {
	if t.duration > 0 {
		select {
		case <-time.After(t.duration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return t.result, t.err
}

func (t *SimpleTask) Priority() int {
	return 1
}

func (t *SimpleTask) EstimatedDuration() time.Duration {
	return t.duration
}

func TestExecutorLocalMode(t *testing.T) {
	executor := NewExecutor(LocalMode)
	ctx := context.Background()

	if err := executor.Start(ctx); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}

	// Submit a simple task
	task := &SimpleTask{
		id:     "test-task-1",
		result: "success",
	}

	if err := executor.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	// Check result
	result, ok := executor.GetResult(task.ID())
	if !ok {
		t.Fatal("Task result not found")
	}

	if result.Error != nil {
		t.Fatalf("Task failed: %v", result.Error)
	}

	if result.Result != "success" {
		t.Fatalf("Expected result 'success', got %v", result.Result)
	}

	executor.Shutdown()
}

func TestVerifiableCache(t *testing.T) {
	// Create security manager
	sm, err := NewSecurityManager("node1")
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Create cache
	signer := NewCryptoSigner(sm)
	verifier := NewCryptoVerifier(sm)
	cache := NewVerifiableCache("node1", signer, verifier)

	// Add self as trusted peer (for testing)
	sm.AddTrustedPeer("node1", sm.GetPublicKey(), 1.0)

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	// Set value
	if err := cache.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}

	// Get value
	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Fatalf("Expected value %s, got %s", value, retrieved)
	}
}

func TestContentAddressedStore(t *testing.T) {
	store := NewContentAddressedStore()

	data := []byte("test content")
	address := store.Put(data)

	// Retrieve by address
	retrieved, err := store.Get(address)
	if err != nil {
		t.Fatalf("Failed to get content: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Fatalf("Expected content %s, got %s", data, retrieved)
	}

	// Check existence
	if !store.Has(address) {
		t.Fatal("Expected content to exist")
	}
}

func TestP2PNetwork(t *testing.T) {
	// Create two networks
	network1, err := NewP2PNetwork("node1", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create network1: %v", err)
	}
	defer network1.Stop()

	network2, err := NewP2PNetwork("node2", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create network2: %v", err)
	}
	defer network2.Stop()

	// Start networks
	network1.Start()
	network2.Start()

	// Get actual listen addresses
	addr1 := network1.listener.Addr().String()

	// Connect network2 to network1
	if err := network2.Connect("node1", addr1); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Check peers
	peers1 := network1.GetPeers()
	peers2 := network2.GetPeers()

	if len(peers1) != 1 {
		t.Fatalf("Expected 1 peer for network1, got %d", len(peers1))
	}

	if len(peers2) != 1 {
		t.Fatalf("Expected 1 peer for network2, got %d", len(peers2))
	}
}

func TestSecurityManager(t *testing.T) {
	sm1, err := NewSecurityManager("node1")
	if err != nil {
		t.Fatalf("Failed to create security manager 1: %v", err)
	}

	sm2, err := NewSecurityManager("node2")
	if err != nil {
		t.Fatalf("Failed to create security manager 2: %v", err)
	}

	// Exchange public keys
	sm1.AddTrustedPeer("node2", sm2.GetPublicKey(), 0.8)
	sm2.AddTrustedPeer("node1", sm1.GetPublicKey(), 0.8)

	// Test signing and verification
	data := []byte("test message")

	// Node1 signs
	signature, err := sm1.Sign(data)
	if err != nil {
		t.Fatalf("Failed to sign: %v", err)
	}

	// Node2 verifies
	if err := sm2.Verify(data, signature, "node1"); err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	// Test trust level
	trust, err := sm1.GetTrustLevel("node2")
	if err != nil {
		t.Fatalf("Failed to get trust level: %v", err)
	}

	if trust != 0.8 {
		t.Fatalf("Expected trust level 0.8, got %f", trust)
	}
}
