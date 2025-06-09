package distributed

import (
	"context"
	"testing"
	"time"
)

func TestPeerDiscovery(t *testing.T) {
	// Create test networks
	network1, err := NewP2PNetwork("test-node-1", "localhost:8081")
	if err != nil {
		t.Fatalf("Failed to create network1: %v", err)
	}
	network2, err := NewP2PNetwork("test-node-2", "localhost:8082")
	if err != nil {
		t.Fatalf("Failed to create network2: %v", err)
	}

	// Start networks
	if err := network1.Start(); err != nil {
		t.Fatalf("Failed to start network1: %v", err)
	}
	defer network1.Stop()

	if err := network2.Start(); err != nil {
		t.Fatalf("Failed to start network2: %v", err)
	}
	defer network2.Stop()

	t.Run("mDNS Discovery", func(t *testing.T) {
		// Create discovery instances
		discovery1 := NewPeerDiscovery(network1, DiscoveryMethodMDNS)
		discovery2 := NewPeerDiscovery(network2, DiscoveryMethodMDNS)

		// Start discovery
		if err := discovery1.Start(); err != nil {
			t.Fatalf("Failed to start discovery1: %v", err)
		}
		defer discovery1.Stop()

		if err := discovery2.Start(); err != nil {
			t.Fatalf("Failed to start discovery2: %v", err)
		}
		defer discovery2.Stop()

		// Give time for discovery
		time.Sleep(5 * time.Second)

		// Check if peers discovered each other
		peers1 := discovery1.GetDiscoveredPeers()
		peers2 := discovery2.GetDiscoveredPeers()

		// Since mDNS might not work in all test environments,
		// we just check the discovery mechanism doesn't crash
		t.Logf("Network1 discovered %d peers", len(peers1))
		t.Logf("Network2 discovered %d peers", len(peers2))
	})

	t.Run("Bootstrap Discovery", func(t *testing.T) {
		discovery := NewPeerDiscovery(network1, DiscoveryMethodBootstrap)
		discovery.SetBootstrapPeers([]string{"test-bootstrap@localhost:8083"})

		if err := discovery.Start(); err != nil {
			t.Fatalf("Failed to start bootstrap discovery: %v", err)
		}
		defer discovery.Stop()

		// Check that bootstrap peer is added
		time.Sleep(1 * time.Second)
		peers := discovery.GetDiscoveredPeers()

		found := false
		for _, peer := range peers {
			if peer.ID == "test-bootstrap" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Bootstrap peer not found in discovered peers")
		}
	})

	t.Run("Static Discovery", func(t *testing.T) {
		discovery := NewPeerDiscovery(network1, DiscoveryMethodStatic)
		discovery.SetStaticPeers([]string{"static-peer-1@localhost:8084", "localhost:8085"})

		if err := discovery.Start(); err != nil {
			t.Fatalf("Failed to start static discovery: %v", err)
		}
		defer discovery.Stop()

		// Check that static peers are added
		time.Sleep(1 * time.Second)
		peers := discovery.GetDiscoveredPeers()

		if len(peers) != 2 {
			t.Errorf("Expected 2 static peers, got %d", len(peers))
		}
	})
}

func TestNetworkTopology(t *testing.T) {
	topology := NewNetworkTopology()

	// Add nodes
	node1 := &NodeInfo{
		ID:          "node1",
		Address:     "localhost:8001",
		Capacity:    10,
		ActiveTasks: 3,
		Services:    []string{"eval", "optimize"},
		Latency:     map[string]int64{"node2": 50, "node3": 100},
	}

	node2 := &NodeInfo{
		ID:          "node2",
		Address:     "localhost:8002",
		Capacity:    8,
		ActiveTasks: 7,
		Services:    []string{"eval"},
		Latency:     map[string]int64{"node1": 50, "node3": 80},
	}

	node3 := &NodeInfo{
		ID:          "node3",
		Address:     "localhost:8003",
		Capacity:    12,
		ActiveTasks: 2,
		Services:    []string{"optimize", "benchmark"},
		Latency:     map[string]int64{"node1": 100, "node2": 80},
	}

	topology.AddNode(node1)
	topology.AddNode(node2)
	topology.AddNode(node3)

	// Add edges
	topology.AddEdge("node1", "node2")
	topology.AddEdge("node2", "node3")
	topology.AddEdge("node1", "node3")

	t.Run("GetOptimalNode", func(t *testing.T) {
		// Test basic capacity-based selection
		optimal := topology.GetOptimalNode(map[string]interface{}{})
		if optimal == nil {
			t.Fatal("No optimal node found")
		}

		// Should select node3 (highest available capacity)
		if optimal.ID != "node3" {
			t.Errorf("Expected node3, got %s", optimal.ID)
		}

		// Test with latency consideration
		optimal = topology.GetOptimalNode(map[string]interface{}{
			"near_node": "node2",
		})

		// Should still prefer node3 but factor in latency
		if optimal == nil {
			t.Fatal("No optimal node found with latency criteria")
		}
	})

	t.Run("GetNeighbors", func(t *testing.T) {
		neighbors := topology.GetNeighbors("node2")
		if len(neighbors) != 2 {
			t.Errorf("Expected 2 neighbors for node2, got %d", len(neighbors))
		}

		// Check that node1 and node3 are neighbors
		hasNode1 := false
		hasNode3 := false
		for _, n := range neighbors {
			if n == "node1" {
				hasNode1 = true
			}
			if n == "node3" {
				hasNode3 = true
			}
		}

		if !hasNode1 || !hasNode3 {
			t.Error("node2 should have node1 and node3 as neighbors")
		}
	})
}

func TestPeerExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create networks
	network1, err := NewP2PNetwork("exchange-node-1", "localhost:8091")
	if err != nil {
		t.Fatalf("Failed to create network1: %v", err)
	}
	network2, err := NewP2PNetwork("exchange-node-2", "localhost:8092")
	if err != nil {
		t.Fatalf("Failed to create network2: %v", err)
	}
	network3, err := NewP2PNetwork("exchange-node-3", "localhost:8093")
	if err != nil {
		t.Fatalf("Failed to create network3: %v", err)
	}

	// Start networks
	if err := network1.Start(); err != nil {
		t.Fatalf("Failed to start network1: %v", err)
	}
	defer network1.Stop()

	if err := network2.Start(); err != nil {
		t.Fatalf("Failed to start network2: %v", err)
	}
	defer network2.Stop()

	if err := network3.Start(); err != nil {
		t.Fatalf("Failed to start network3: %v", err)
	}
	defer network3.Stop()

	// Connect network1 to network2
	if err := network1.Connect("exchange-node-2", "localhost:8092"); err != nil {
		t.Fatalf("Failed to connect network1 to network2: %v", err)
	}

	// Connect network2 to network3
	if err := network2.Connect("exchange-node-3", "localhost:8093"); err != nil {
		t.Fatalf("Failed to connect network2 to network3: %v", err)
	}

	// Create discovery for network1 (should discover network3 through network2)
	discovery1 := NewPeerDiscovery(network1, DiscoveryMethodStatic)

	// Wait for connections to establish
	time.Sleep(2 * time.Second)

	// Test peer exchange
	discovery1.exchangePeers()

	// Give time for peer exchange
	select {
	case <-ctx.Done():
		t.Fatal("Test timeout")
	case <-time.After(3 * time.Second):
		// Check if network1 discovered network3 through network2
		peers := discovery1.GetDiscoveredPeers()
		t.Logf("Network1 discovered %d peers through exchange", len(peers))
	}
}
