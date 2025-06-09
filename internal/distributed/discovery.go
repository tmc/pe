package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

// DiscoveryMethod defines how peers are discovered
type DiscoveryMethod string

const (
	DiscoveryMethodMDNS      DiscoveryMethod = "mdns"
	DiscoveryMethodBootstrap DiscoveryMethod = "bootstrap"
	DiscoveryMethodStatic    DiscoveryMethod = "static"
)

// PeerDiscovery handles peer discovery in the network
type PeerDiscovery struct {
	network         *P2PNetwork
	method          DiscoveryMethod
	bootstrapPeers  []string
	staticPeers     []string
	discoveredPeers map[string]*DiscoveredPeer
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	mdnsServer      *mdns.Server
	mdnsMu          sync.Mutex // Protects mdnsServer access
	serviceInfo     *mdns.MDNSService
}

// DiscoveredPeer represents a discovered peer with metadata
type DiscoveredPeer struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	PublicKey  string    `json:"public_key"`
	Capacity   int       `json:"capacity"`
	Version    string    `json:"version"`
	Services   []string  `json:"services"`
	LastSeen   time.Time `json:"last_seen"`
	Reputation float64   `json:"reputation"`
}

// NewPeerDiscovery creates a new peer discovery instance
func NewPeerDiscovery(network *P2PNetwork, method DiscoveryMethod) *PeerDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	return &PeerDiscovery{
		network:         network,
		method:          method,
		discoveredPeers: make(map[string]*DiscoveredPeer),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins the peer discovery process
func (pd *PeerDiscovery) Start() error {
	switch pd.method {
	case DiscoveryMethodMDNS:
		go pd.startMDNSDiscovery()
	case DiscoveryMethodBootstrap:
		go pd.startBootstrapDiscovery()
	case DiscoveryMethodStatic:
		go pd.startStaticDiscovery()
	default:
		return fmt.Errorf("unknown discovery method: %s", pd.method)
	}

	// Start periodic peer exchange
	go pd.peerExchangeLoop()

	return nil
}

// Stop halts the peer discovery process
func (pd *PeerDiscovery) Stop() error {
	pd.cancel()

	// Stop mDNS server if running
	pd.mdnsMu.Lock()
	if pd.mdnsServer != nil {
		pd.mdnsServer.Shutdown()
		pd.mdnsServer = nil
	}
	pd.mdnsMu.Unlock()

	return nil
}

// SetBootstrapPeers sets the list of bootstrap peers
func (pd *PeerDiscovery) SetBootstrapPeers(peers []string) {
	pd.bootstrapPeers = peers
}

// SetStaticPeers sets the list of static peers
func (pd *PeerDiscovery) SetStaticPeers(peers []string) {
	pd.staticPeers = peers
}

// GetDiscoveredPeers returns all discovered peers
func (pd *PeerDiscovery) GetDiscoveredPeers() []*DiscoveredPeer {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	peers := make([]*DiscoveredPeer, 0, len(pd.discoveredPeers))
	for _, peer := range pd.discoveredPeers {
		peers = append(peers, peer)
	}
	return peers
}

// startMDNSDiscovery implements mDNS-based peer discovery
func (pd *PeerDiscovery) startMDNSDiscovery() {
	// Setup mDNS service info
	info := []string{
		"version=v1.0.0", // TODO: Get from config
		"id=" + pd.network.nodeID,
	}

	// Get the actual port from network listener
	port := 8080 // Default port
	if pd.network.listener != nil {
		if addr, ok := pd.network.listener.Addr().(*net.TCPAddr); ok {
			port = addr.Port
		}
	}

	// Create mDNS service
	service, err := mdns.NewMDNSService(
		pd.network.nodeID,
		"_pe-distributed._tcp",
		"",
		"",
		port,
		nil, // IPs will be determined automatically
		info,
	)
	if err != nil {
		log.Printf("Failed to create mDNS service: %v", err)
		return
	}
	pd.serviceInfo = service

	// Create mDNS server
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		log.Printf("Failed to create mDNS server: %v", err)
		return
	}
	pd.mdnsMu.Lock()
	pd.mdnsServer = server
	pd.mdnsMu.Unlock()

	// Start periodic discovery
	go pd.runMDNSQueries()
}

// runMDNSQueries performs periodic mDNS queries to discover peers
func (pd *PeerDiscovery) runMDNSQueries() {
	// Initial delay to let service start
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run initial query
	pd.queryMDNSPeers()

	for {
		select {
		case <-pd.ctx.Done():
			return
		case <-ticker.C:
			pd.queryMDNSPeers()
		}
	}
}

// queryMDNSPeers queries for PE distributed nodes on the network
func (pd *PeerDiscovery) queryMDNSPeers() {
	// Create a channel to receive entries
	entriesCh := make(chan *mdns.ServiceEntry, 10)

	go func() {
		for entry := range entriesCh {
			// Skip our own service
			if entry.Name == pd.network.nodeID {
				continue
			}

			// Extract peer info from TXT records
			var peerID string
			var version string
			for _, txt := range entry.InfoFields {
				if strings.HasPrefix(txt, "id=") {
					peerID = strings.TrimPrefix(txt, "id=")
				} else if strings.HasPrefix(txt, "version=") {
					version = strings.TrimPrefix(txt, "version=")
				}
			}

			if peerID == "" {
				peerID = entry.Name
			}

			// Create discovered peer
			addr := fmt.Sprintf("%s:%d", entry.AddrV4, entry.Port)
			if entry.AddrV4 == nil && entry.AddrV6 != nil {
				addr = fmt.Sprintf("[%s]:%d", entry.AddrV6, entry.Port)
			}

			peer := &DiscoveredPeer{
				ID:       peerID,
				Address:  addr,
				Version:  version,
				Services: []string{"_pe-distributed._tcp"},
				LastSeen: time.Now(),
			}

			// Add to discovered peers
			pd.mu.Lock()
			pd.discoveredPeers[peer.ID] = peer
			pd.mu.Unlock()

			// Try to connect
			if err := pd.network.Connect(peer.ID, peer.Address); err != nil {
				log.Printf("Failed to connect to discovered peer %s at %s: %v", peer.ID, peer.Address, err)
			} else {
				log.Printf("Connected to discovered peer %s at %s", peer.ID, peer.Address)
			}
		}
	}()

	// Start the query
	params := mdns.DefaultParams("_pe-distributed._tcp")
	params.Entries = entriesCh
	params.Timeout = 5 * time.Second

	if err := mdns.Query(params); err != nil {
		log.Printf("mDNS query failed: %v", err)
	}

	close(entriesCh)
}

// startBootstrapDiscovery connects to bootstrap peers and discovers network
func (pd *PeerDiscovery) startBootstrapDiscovery() {
	// Connect to bootstrap peers
	for _, addr := range pd.bootstrapPeers {
		parts := strings.Split(addr, "@")
		var peerID, peerAddr string

		if len(parts) == 2 {
			peerID = parts[0]
			peerAddr = parts[1]
		} else {
			peerID = fmt.Sprintf("bootstrap-%s", addr)
			peerAddr = addr
		}

		peer := &DiscoveredPeer{
			ID:       peerID,
			Address:  peerAddr,
			Services: []string{"bootstrap"},
			LastSeen: time.Now(),
		}

		pd.mu.Lock()
		pd.discoveredPeers[peer.ID] = peer
		pd.mu.Unlock()

		// Connect to bootstrap peer
		if err := pd.network.Connect(peer.ID, peer.Address); err != nil {
			fmt.Printf("Failed to connect to bootstrap peer %s: %v\n", peer.ID, err)
			continue
		}

		// Request peer list from bootstrap
		pd.requestPeerList(peer.ID)
	}
}

// startStaticDiscovery connects to statically configured peers
func (pd *PeerDiscovery) startStaticDiscovery() {
	for _, addr := range pd.staticPeers {
		parts := strings.Split(addr, "@")
		var peerID, peerAddr string

		if len(parts) == 2 {
			peerID = parts[0]
			peerAddr = parts[1]
		} else {
			peerID = fmt.Sprintf("static-%s", addr)
			peerAddr = addr
		}

		peer := &DiscoveredPeer{
			ID:       peerID,
			Address:  peerAddr,
			Services: []string{"static"},
			LastSeen: time.Now(),
		}

		pd.mu.Lock()
		pd.discoveredPeers[peer.ID] = peer
		pd.mu.Unlock()

		// Connect to static peer
		if err := pd.network.Connect(peer.ID, peer.Address); err != nil {
			fmt.Printf("Failed to connect to static peer %s: %v\n", peer.ID, err)
		}
	}
}

// peerExchangeLoop periodically exchanges peer lists with connected peers
func (pd *PeerDiscovery) peerExchangeLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pd.ctx.Done():
			return
		case <-ticker.C:
			pd.exchangePeers()
		}
	}
}

// exchangePeers exchanges peer lists with all connected peers
func (pd *PeerDiscovery) exchangePeers() {
	peers := pd.network.GetPeers()
	for _, peer := range peers {
		pd.requestPeerList(peer.ID)
	}
}

// requestPeerList requests a peer's known peers
func (pd *PeerDiscovery) requestPeerList(peerID string) {
	// Create peer list request
	request := map[string]interface{}{
		"type": "peer_list_request",
		"max":  50, // Limit number of peers
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return
	}

	// Find peer
	var targetPeer *Peer
	for _, peer := range pd.network.GetPeers() {
		if peer.ID == peerID {
			targetPeer = peer
			break
		}
	}

	if targetPeer == nil {
		return
	}

	// Send request (using a custom message type)
	pd.network.SendMessage(targetPeer, &Message{
		Type:      "peer_exchange",
		ID:        generateMessageID(),
		From:      pd.network.nodeID,
		To:        peerID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// NetworkTopology tracks the overall network structure
type NetworkTopology struct {
	nodes      map[string]*NodeInfo
	edges      map[string]map[string]bool // adjacency list
	mu         sync.RWMutex
	lastUpdate time.Time
}

// NodeInfo contains information about a node in the network
type NodeInfo struct {
	ID          string           `json:"id"`
	Address     string           `json:"address"`
	Capacity    int              `json:"capacity"`
	ActiveTasks int              `json:"active_tasks"`
	Services    []string         `json:"services"`
	Latency     map[string]int64 `json:"latency"` // ms to other nodes
	LastUpdate  time.Time        `json:"last_update"`
}

// NewNetworkTopology creates a new network topology tracker
func NewNetworkTopology() *NetworkTopology {
	return &NetworkTopology{
		nodes:      make(map[string]*NodeInfo),
		edges:      make(map[string]map[string]bool),
		lastUpdate: time.Now(),
	}
}

// AddNode adds or updates a node in the topology
func (nt *NetworkTopology) AddNode(info *NodeInfo) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	nt.nodes[info.ID] = info
	if _, exists := nt.edges[info.ID]; !exists {
		nt.edges[info.ID] = make(map[string]bool)
	}
	nt.lastUpdate = time.Now()
}

// AddEdge adds a connection between two nodes
func (nt *NetworkTopology) AddEdge(nodeA, nodeB string) {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	if _, exists := nt.edges[nodeA]; !exists {
		nt.edges[nodeA] = make(map[string]bool)
	}
	if _, exists := nt.edges[nodeB]; !exists {
		nt.edges[nodeB] = make(map[string]bool)
	}

	nt.edges[nodeA][nodeB] = true
	nt.edges[nodeB][nodeA] = true
	nt.lastUpdate = time.Now()
}

// GetOptimalNode returns the best node for task execution based on criteria
func (nt *NetworkTopology) GetOptimalNode(criteria map[string]interface{}) *NodeInfo {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	var bestNode *NodeInfo
	bestScore := -1.0

	for _, node := range nt.nodes {
		// Simple scoring based on capacity and active tasks
		availableCapacity := float64(node.Capacity - node.ActiveTasks)
		if availableCapacity <= 0 {
			continue
		}

		score := availableCapacity / float64(node.Capacity)

		// Factor in latency if specified
		if targetNode, ok := criteria["near_node"].(string); ok {
			if latency, exists := node.Latency[targetNode]; exists {
				// Prefer lower latency
				score *= (1.0 / (1.0 + float64(latency)/1000.0))
			}
		}

		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode
}

// GetNeighbors returns the neighbors of a node
func (nt *NetworkTopology) GetNeighbors(nodeID string) []string {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	neighbors := make([]string, 0)
	if edges, exists := nt.edges[nodeID]; exists {
		for neighbor := range edges {
			neighbors = append(neighbors, neighbor)
		}
	}
	return neighbors
}
