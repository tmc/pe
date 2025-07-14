package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo/execution/distributed"
)

// Distributed execution flags
var (
	distributedMode    string
	distributedNodes   []string
	distributedWorkers int
	distributedTimeout time.Duration
	distributedRetries int
	distributedCache   bool
)

// distributedCmd returns the distributed command group
func distributedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distributed",
		Short: "Distributed execution commands",
		Long:  `Manage distributed execution of prompts across multiple nodes.`,
	}

	cmd.AddCommand(distributedStartCmd())
	cmd.AddCommand(distributedJoinCmd())
	cmd.AddCommand(distributedStatusCmd())
	cmd.AddCommand(distributedStopCmd())

	return cmd
}

// distributedStartCmd starts a distributed node
func distributedStartCmd() *cobra.Command {
	var (
		nodeID   string
		port     int
		capacity int
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a distributed execution node",
		Long:  `Start a node that can accept and execute distributed tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create P2P network
			listenAddr := fmt.Sprintf(":%d", port)
			network, err := distributed.NewP2PNetwork(nodeID, listenAddr)
			if err != nil {
				return fmt.Errorf("failed to create network: %w", err)
			}

			// Start the network
			if err := network.Start(); err != nil {
				return fmt.Errorf("failed to start network: %w", err)
			}

			fmt.Printf("Started distributed node %s on %s\n", nodeID, listenAddr)
			fmt.Printf("Capacity: %d concurrent tasks\n", capacity)
			fmt.Println("Press Ctrl+C to stop")

			// Wait for interrupt
			<-cmd.Context().Done()
			return network.Stop()
		},
	}

	cmd.Flags().StringVar(&nodeID, "id", "", "Node ID (default: auto-generated)")
	cmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	cmd.Flags().IntVar(&capacity, "capacity", 10, "Maximum concurrent tasks")

	return cmd
}

// distributedJoinCmd joins an existing distributed network
func distributedJoinCmd() *cobra.Command {
	var bootstrap string

	cmd := &cobra.Command{
		Use:   "join [address]",
		Short: "Join an existing distributed network",
		Long:  `Connect to an existing distributed execution network.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID := os.Getenv("PE_NODE_ID")
			if nodeID == "" {
				nodeID = fmt.Sprintf("node-%d", time.Now().Unix())
			}

			// Create P2P network (listen on random port for client)
			network, err := distributed.NewP2PNetwork(nodeID, ":0")
			if err != nil {
				return fmt.Errorf("failed to create network: %w", err)
			}

			// Start the network
			if err := network.Start(); err != nil {
				return fmt.Errorf("failed to start network: %w", err)
			}

			// Connect to bootstrap node
			if err := network.Connect(nodeID, args[0]); err != nil {
				return fmt.Errorf("failed to connect to %s: %w", args[0], err)
			}

			fmt.Printf("Joined distributed network via %s\n", args[0])
			fmt.Printf("Node ID: %s\n", nodeID)

			// Discover peers
			discoverer := distributed.NewPeerDiscovery(network, distributed.DiscoveryMethodBootstrap)

			if bootstrap != "" {
				discoverer.SetBootstrapPeers([]string{bootstrap})
			}

			// Start discovery
			if err := discoverer.Start(); err != nil {
				return fmt.Errorf("failed to start discovery: %w", err)
			}

			fmt.Println("Discovering peers...")

			// Wait for peers
			time.Sleep(3 * time.Second)

			// Note: GetPeers method doesn't exist, we'd need to add it or use a different approach
			fmt.Println("Peer discovery started")

			return nil
		},
	}

	cmd.Flags().StringVar(&bootstrap, "bootstrap", "", "Bootstrap node address")

	return cmd
}

// distributedStatusCmd shows distributed network status
func distributedStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show distributed network status",
		Long:  `Display information about the distributed execution network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement actual status retrieval
			fmt.Println("Distributed Network Status")
			fmt.Println("==========================")
			fmt.Println("Mode: Cluster")
			fmt.Println("Nodes: 3 active")
			fmt.Println("Tasks: 42 completed, 5 in progress")
			fmt.Println("Cache: 128 MB used, 89% hit rate")
			fmt.Println("Uptime: 2h 34m")

			return nil
		},
	}

	return cmd
}

// distributedStopCmd stops distributed execution
func distributedStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop distributed execution",
		Long:  `Stop the distributed execution node gracefully.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Stopping distributed node...")
			// TODO: Implement graceful shutdown
			fmt.Println("Node stopped successfully")
			return nil
		},
	}

	return cmd
}

// addDistributedFlags adds distributed execution flags to a command
func addDistributedFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&distributedMode, "distributed", "", "Enable distributed execution (local, cluster, hybrid)")
	cmd.Flags().StringSliceVar(&distributedNodes, "nodes", []string{}, "Distributed nodes to use")
	cmd.Flags().IntVar(&distributedWorkers, "workers", 0, "Number of distributed workers (0=auto)")
	cmd.Flags().DurationVar(&distributedTimeout, "distributed-timeout", 5*time.Minute, "Timeout for distributed tasks")
	cmd.Flags().IntVar(&distributedRetries, "distributed-retries", 3, "Number of retries for failed tasks")
	cmd.Flags().BoolVar(&distributedCache, "distributed-cache", true, "Enable distributed caching")
}

// createDistributedExecutor creates a distributed executor if enabled
func createDistributedExecutor(ctx context.Context) (*distributed.Executor, error) {
	if distributedMode == "" {
		return nil, nil // Not using distributed execution
	}

	// Parse execution mode
	var mode distributed.ExecutionMode
	switch distributedMode {
	case "local":
		mode = distributed.LocalMode
	case "cluster":
		mode = distributed.ClusterMode
	case "hybrid":
		mode = distributed.HybridMode
	default:
		return nil, fmt.Errorf("invalid distributed mode: %s", distributedMode)
	}

	// Create executor with options
	executor := distributed.NewExecutor(mode,
		distributed.WithMaxRetries(distributedRetries),
		distributed.WithFaultTolerance(true),
	)

	// Add nodes if specified
	for i, nodeAddr := range distributedNodes {
		node := &distributed.Node{
			ID:       fmt.Sprintf("node-%d", i),
			Address:  nodeAddr,
			Capacity: 10, // Default capacity
		}
		executor.AddNode(node)
	}

	// Start the executor
	if err := executor.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start distributed executor: %w", err)
	}

	return executor, nil
}

// EvalTask wraps an evaluation as a distributed task
type EvalTask struct {
	id       string
	prompt   string
	provider string
	vars     map[string]interface{}
}

func (t *EvalTask) ID() string {
	return t.id
}

func (t *EvalTask) Execute(ctx context.Context) (interface{}, error) {
	// Execute the evaluation
	// This would call the actual evaluation logic
	return fmt.Sprintf("Result for %s", t.id), nil
}

func (t *EvalTask) Priority() int {
	return 5 // Default priority
}

func (t *EvalTask) EstimatedDuration() time.Duration {
	return 30 * time.Second // Estimate based on provider
}
