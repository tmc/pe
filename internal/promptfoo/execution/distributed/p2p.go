package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// MessageType defines the type of P2P message
type MessageType string

const (
	MessageTypePing            MessageType = "ping"
	MessageTypePong            MessageType = "pong"
	MessageTypeTaskRequest     MessageType = "task_request"
	MessageTypeTaskResponse    MessageType = "task_response"
	MessageTypeCacheRequest    MessageType = "cache_request"
	MessageTypeCacheResponse   MessageType = "cache_response"
	MessageTypeWitnessRequest  MessageType = "witness_request"
	MessageTypeWitnessResponse MessageType = "witness_response"
)

// Message represents a P2P message
type Message struct {
	Type      MessageType     `json:"type"`
	ID        string          `json:"id"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// Peer represents a peer in the P2P network
type Peer struct {
	ID        string
	Address   string
	PublicKey string
	LastSeen  time.Time
	conn      net.Conn
	mu        sync.RWMutex
}

// P2PNetwork manages peer-to-peer communications
type P2PNetwork struct {
	nodeID   string
	listener net.Listener
	peers    map[string]*Peer
	handlers map[MessageType]MessageHandler
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, msg *Message, peer *Peer) error

// NewP2PNetwork creates a new P2P network instance
func NewP2PNetwork(nodeID string, listenAddr string) (*P2PNetwork, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	network := &P2PNetwork{
		nodeID:   nodeID,
		listener: listener,
		peers:    make(map[string]*Peer),
		handlers: make(map[MessageType]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register default handlers
	network.RegisterHandler(MessageTypePing, network.handlePing)
	network.RegisterHandler(MessageTypePong, network.handlePong)

	return network, nil
}

// Start begins accepting connections
func (n *P2PNetwork) Start() error {
	go n.acceptConnections()
	go n.maintainPeers()
	return nil
}

// Stop shuts down the P2P network
func (n *P2PNetwork) Stop() error {
	n.cancel()
	return n.listener.Close()
}

// RegisterHandler registers a message handler
func (n *P2PNetwork) RegisterHandler(msgType MessageType, handler MessageHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers[msgType] = handler
}

// Connect establishes a connection to a peer
func (n *P2PNetwork) Connect(peerID, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerID, err)
	}

	peer := &Peer{
		ID:       peerID,
		Address:  address,
		LastSeen: time.Now(),
		conn:     conn,
	}

	n.mu.Lock()
	n.peers[peerID] = peer
	n.mu.Unlock()

	// Start handling messages from this peer
	go n.handlePeer(peer)

	// Send initial ping
	return n.SendMessage(peer, &Message{
		Type:      MessageTypePing,
		ID:        generateMessageID(),
		From:      n.nodeID,
		To:        peerID,
		Timestamp: time.Now(),
	})
}

// SendMessage sends a message to a peer
func (n *P2PNetwork) SendMessage(peer *Peer, msg *Message) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	encoder := json.NewEncoder(peer.conn)
	return encoder.Encode(msg)
}

// Broadcast sends a message to all peers
func (n *P2PNetwork) Broadcast(msg *Message) error {
	n.mu.RLock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	n.mu.RUnlock()

	var lastErr error
	for _, peer := range peers {
		if err := n.SendMessage(peer, msg); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// acceptConnections accepts incoming peer connections
func (n *P2PNetwork) acceptConnections() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				continue
			}

			// Handle new peer connection
			go n.handleIncomingPeer(conn)
		}
	}
}

// handleIncomingPeer handles a new incoming peer connection
func (n *P2PNetwork) handleIncomingPeer(conn net.Conn) {
	// TODO: Implement peer authentication/handshake

	peer := &Peer{
		ID:       fmt.Sprintf("peer-%s", conn.RemoteAddr()),
		Address:  conn.RemoteAddr().String(),
		LastSeen: time.Now(),
		conn:     conn,
	}

	n.mu.Lock()
	n.peers[peer.ID] = peer
	n.mu.Unlock()

	n.handlePeer(peer)
}

// handlePeer processes messages from a peer
func (n *P2PNetwork) handlePeer(peer *Peer) {
	defer func() {
		peer.conn.Close()
		n.mu.Lock()
		delete(n.peers, peer.ID)
		n.mu.Unlock()
	}()

	decoder := json.NewDecoder(peer.conn)
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			var msg Message
			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			// Update last seen
			peer.mu.Lock()
			peer.LastSeen = time.Now()
			peer.mu.Unlock()

			// Handle message
			n.handleMessage(&msg, peer)
		}
	}
}

// handleMessage processes an incoming message
func (n *P2PNetwork) handleMessage(msg *Message, peer *Peer) {
	n.mu.RLock()
	handler, exists := n.handlers[msg.Type]
	n.mu.RUnlock()

	if !exists {
		return
	}

	if err := handler(n.ctx, msg, peer); err != nil {
		// Log error
		fmt.Printf("Error handling message %s from %s: %v\n", msg.Type, peer.ID, err)
	}
}

// handlePing responds to ping messages
func (n *P2PNetwork) handlePing(ctx context.Context, msg *Message, peer *Peer) error {
	return n.SendMessage(peer, &Message{
		Type:      MessageTypePong,
		ID:        generateMessageID(),
		From:      n.nodeID,
		To:        peer.ID,
		Timestamp: time.Now(),
	})
}

// handlePong processes pong responses
func (n *P2PNetwork) handlePong(ctx context.Context, msg *Message, peer *Peer) error {
	// Pong received, peer is alive
	return nil
}

// maintainPeers periodically checks peer health
func (n *P2PNetwork) maintainPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.checkPeerHealth()
		}
	}
}

// checkPeerHealth removes inactive peers
func (n *P2PNetwork) checkPeerHealth() {
	n.mu.Lock()
	defer n.mu.Unlock()

	timeout := 2 * time.Minute
	now := time.Now()

	for id, peer := range n.peers {
		peer.mu.RLock()
		lastSeen := peer.LastSeen
		peer.mu.RUnlock()

		if now.Sub(lastSeen) > timeout {
			peer.conn.Close()
			delete(n.peers, id)
		}
	}
}

// GetPeers returns a list of connected peers
func (n *P2PNetwork) GetPeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	return peers
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// TaskDistributor distributes tasks across peers
type TaskDistributor struct {
	network      *P2PNetwork
	executor     *Executor
	pendingTasks map[string]Task
	mu           sync.RWMutex
}

// NewTaskDistributor creates a new task distributor
func NewTaskDistributor(network *P2PNetwork, executor *Executor) *TaskDistributor {
	td := &TaskDistributor{
		network:      network,
		executor:     executor,
		pendingTasks: make(map[string]Task),
	}

	// Register task handlers
	network.RegisterHandler(MessageTypeTaskRequest, td.handleTaskRequest)
	network.RegisterHandler(MessageTypeTaskResponse, td.handleTaskResponse)

	return td
}

// DistributeTask sends a task to a remote peer for execution
func (td *TaskDistributor) DistributeTask(task Task, peerID string) error {
	td.mu.Lock()
	td.pendingTasks[task.ID()] = task
	td.mu.Unlock()

	// Serialize task
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	// Find peer
	var targetPeer *Peer
	for _, peer := range td.network.GetPeers() {
		if peer.ID == peerID {
			targetPeer = peer
			break
		}
	}

	if targetPeer == nil {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// Send task request
	return td.network.SendMessage(targetPeer, &Message{
		Type:      MessageTypeTaskRequest,
		ID:        generateMessageID(),
		From:      td.network.nodeID,
		To:        peerID,
		Timestamp: time.Now(),
		Payload:   taskData,
	})
}

// handleTaskRequest processes incoming task requests
func (td *TaskDistributor) handleTaskRequest(ctx context.Context, msg *Message, peer *Peer) error {
	// TODO: Deserialize and execute task
	// For now, just acknowledge
	return td.network.SendMessage(peer, &Message{
		Type:      MessageTypeTaskResponse,
		ID:        generateMessageID(),
		From:      td.network.nodeID,
		To:        peer.ID,
		Timestamp: time.Now(),
	})
}

// handleTaskResponse processes task execution responses
func (td *TaskDistributor) handleTaskResponse(ctx context.Context, msg *Message, peer *Peer) error {
	// TODO: Process task results
	return nil
}

// CacheSynchronizer handles cache synchronization between peers
type CacheSynchronizer struct {
	network *P2PNetwork
	cache   *VerifiableCache
	mu      sync.RWMutex
}

// CacheSyncRequest represents a cache synchronization request
type CacheSyncRequest struct {
	Keys      []string `json:"keys,omitempty"`      // Specific keys to sync, empty for all
	Since     int64    `json:"since,omitempty"`     // Unix timestamp for incremental sync
	Witnesses bool     `json:"witnesses,omitempty"` // Include witness data
}

// CacheSyncResponse contains synchronized cache entries
type CacheSyncResponse struct {
	Entries []CacheEntryWithWitnesses `json:"entries"`
	More    bool                      `json:"more"` // Indicates if more entries are available
}

// CacheEntryWithWitnesses includes the cache entry and its witnesses
type CacheEntryWithWitnesses struct {
	Key       string        `json:"key"`
	Entry     CacheEntry    `json:"entry"`
	Witnesses []WitnessData `json:"witnesses"`
}

// NewCacheSynchronizer creates a new cache synchronizer
func NewCacheSynchronizer(network *P2PNetwork, cache *VerifiableCache) *CacheSynchronizer {
	cs := &CacheSynchronizer{
		network: network,
		cache:   cache,
	}

	// Register cache handlers
	network.RegisterHandler(MessageTypeCacheRequest, cs.handleCacheRequest)
	network.RegisterHandler(MessageTypeCacheResponse, cs.handleCacheResponse)

	return cs
}

// SyncWithPeer synchronizes cache with a specific peer
func (cs *CacheSynchronizer) SyncWithPeer(peerID string, keys []string) error {
	request := CacheSyncRequest{
		Keys:      keys,
		Witnesses: true,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal sync request: %w", err)
	}

	// Find peer
	var targetPeer *Peer
	for _, peer := range cs.network.GetPeers() {
		if peer.ID == peerID {
			targetPeer = peer
			break
		}
	}

	if targetPeer == nil {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// Send cache sync request
	return cs.network.SendMessage(targetPeer, &Message{
		Type:      MessageTypeCacheRequest,
		ID:        generateMessageID(),
		From:      cs.network.nodeID,
		To:        peerID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// BroadcastCacheUpdate broadcasts a cache update to all peers
func (cs *CacheSynchronizer) BroadcastCacheUpdate(key string) error {
	entry, witnesses, err := cs.cache.GetWithWitnesses(key)
	if err != nil {
		return fmt.Errorf("failed to get cache entry: %w", err)
	}

	update := CacheSyncResponse{
		Entries: []CacheEntryWithWitnesses{
			{
				Key:       key,
				Entry:     *entry,
				Witnesses: witnesses,
			},
		},
		More: false,
	}

	payload, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal cache update: %w", err)
	}

	// Broadcast to all peers
	return cs.network.Broadcast(&Message{
		Type:      MessageTypeCacheResponse,
		ID:        generateMessageID(),
		From:      cs.network.nodeID,
		To:        "", // Empty for broadcast
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// handleCacheRequest processes incoming cache sync requests
func (cs *CacheSynchronizer) handleCacheRequest(ctx context.Context, msg *Message, peer *Peer) error {
	var request CacheSyncRequest
	if err := json.Unmarshal(msg.Payload, &request); err != nil {
		return fmt.Errorf("failed to unmarshal cache request: %w", err)
	}

	// Collect requested entries
	var entries []CacheEntryWithWitnesses

	if len(request.Keys) > 0 {
		// Sync specific keys
		for _, key := range request.Keys {
			entry, witnesses, err := cs.cache.GetWithWitnesses(key)
			if err != nil {
				continue // Skip missing keys
			}

			entries = append(entries, CacheEntryWithWitnesses{
				Key:       key,
				Entry:     *entry,
				Witnesses: witnesses,
			})
		}
	} else {
		// Sync all entries (limited batch)
		allEntries := cs.cache.store.List()
		limit := 100 // Batch size

		for i, kv := range allEntries {
			if i >= limit {
				break
			}

			entry, witnesses, err := cs.cache.GetWithWitnesses(kv.Key)
			if err != nil {
				continue
			}

			// Filter by timestamp if requested
			if request.Since > 0 && entry.Timestamp.Unix() <= request.Since {
				continue
			}

			entries = append(entries, CacheEntryWithWitnesses{
				Key:       kv.Key,
				Entry:     *entry,
				Witnesses: witnesses,
			})
		}
	}

	// Prepare response
	response := CacheSyncResponse{
		Entries: entries,
		More:    false, // TODO: Implement pagination
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal cache response: %w", err)
	}

	// Send response
	return cs.network.SendMessage(peer, &Message{
		Type:      MessageTypeCacheResponse,
		ID:        generateMessageID(),
		From:      cs.network.nodeID,
		To:        peer.ID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// handleCacheResponse processes incoming cache updates
func (cs *CacheSynchronizer) handleCacheResponse(ctx context.Context, msg *Message, peer *Peer) error {
	var response CacheSyncResponse
	if err := json.Unmarshal(msg.Payload, &response); err != nil {
		return fmt.Errorf("failed to unmarshal cache response: %w", err)
	}

	// Import received entries
	for _, item := range response.Entries {
		// Verify the entry signature
		if err := cs.cache.VerifyEntry(&item.Entry); err != nil {
			continue // Skip invalid entries
		}

		// Import the entry with witnesses
		cs.cache.store.Set(item.Key, item.Entry)

		// Add witnesses
		for _, witness := range item.Witnesses {
			cs.cache.addWitness(item.Key, witness)
		}
	}

	// TODO: Handle pagination if response.More is true

	return nil
}
