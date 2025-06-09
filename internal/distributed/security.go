package distributed

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SecurityManager handles cryptographic operations and trust management
type SecurityManager struct {
	nodeID       string
	privateKey   ed25519.PrivateKey
	publicKey    ed25519.PublicKey
	trustedPeers map[string]*TrustedPeer
	mu           sync.RWMutex
}

// TrustedPeer represents a peer with trust information
type TrustedPeer struct {
	ID           string
	PublicKey    ed25519.PublicKey
	TrustLevel   float64 // 0.0 to 1.0
	LastVerified time.Time
	Reputation   int // Positive or negative based on behavior
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(nodeID string) (*SecurityManager, error) {
	// Generate key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &SecurityManager{
		nodeID:       nodeID,
		privateKey:   privateKey,
		publicKey:    publicKey,
		trustedPeers: make(map[string]*TrustedPeer),
	}, nil
}

// GetPublicKey returns the node's public key
func (sm *SecurityManager) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(sm.publicKey)
}

// Sign creates a digital signature for data
func (sm *SecurityManager) Sign(data []byte) (string, error) {
	signature := ed25519.Sign(sm.privateKey, data)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify checks a signature against data and peer ID
func (sm *SecurityManager) Verify(data []byte, signature string, peerID string) error {
	sm.mu.RLock()
	peer, exists := sm.trustedPeers[peerID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown peer: %s", peerID)
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	if !ed25519.Verify(peer.PublicKey, data, sig) {
		// Decrease reputation on failed verification
		sm.updateReputation(peerID, -1)
		return fmt.Errorf("signature verification failed")
	}

	// Update last verified time
	sm.mu.Lock()
	peer.LastVerified = time.Now()
	sm.mu.Unlock()

	return nil
}

// AddTrustedPeer adds a peer to the trust store
func (sm *SecurityManager) AddTrustedPeer(peerID string, publicKeyStr string, initialTrust float64) error {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return fmt.Errorf("invalid public key format: %w", err)
	}

	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.trustedPeers[peerID] = &TrustedPeer{
		ID:           peerID,
		PublicKey:    ed25519.PublicKey(publicKey),
		TrustLevel:   initialTrust,
		LastVerified: time.Now(),
		Reputation:   0,
	}

	return nil
}

// RemoveTrustedPeer removes a peer from the trust store
func (sm *SecurityManager) RemoveTrustedPeer(peerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.trustedPeers, peerID)
}

// GetTrustLevel returns the trust level for a peer
func (sm *SecurityManager) GetTrustLevel(peerID string) (float64, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	peer, exists := sm.trustedPeers[peerID]
	if !exists {
		return 0, fmt.Errorf("unknown peer: %s", peerID)
	}

	return peer.TrustLevel, nil
}

// updateReputation adjusts a peer's reputation
func (sm *SecurityManager) updateReputation(peerID string, delta int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	peer, exists := sm.trustedPeers[peerID]
	if !exists {
		return
	}

	peer.Reputation += delta

	// Adjust trust level based on reputation
	if peer.Reputation < -10 {
		peer.TrustLevel = 0.1
	} else if peer.Reputation > 10 {
		peer.TrustLevel = 0.9
	} else {
		peer.TrustLevel = 0.5 + float64(peer.Reputation)*0.04
	}
}

// SecureExecutor wraps the executor with security features
type SecureExecutor struct {
	*Executor
	securityManager *SecurityManager
	sandbox         *Sandbox
}

// NewSecureExecutor creates a new secure executor
func NewSecureExecutor(mode ExecutionMode, sm *SecurityManager, opts ...ExecutorOption) *SecureExecutor {
	return &SecureExecutor{
		Executor:        NewExecutor(mode, opts...),
		securityManager: sm,
		sandbox:         NewSandbox(),
	}
}

// ExecuteSecure executes a task with security checks
func (se *SecureExecutor) ExecuteSecure(ctx context.Context, task Task, peerID string) (interface{}, error) {
	// Check trust level
	trustLevel, err := se.securityManager.GetTrustLevel(peerID)
	if err != nil {
		return nil, fmt.Errorf("cannot execute task from untrusted peer: %w", err)
	}

	if trustLevel < 0.3 {
		return nil, fmt.Errorf("insufficient trust level: %.2f", trustLevel)
	}

	// Execute in sandbox if trust level is moderate
	if trustLevel < 0.7 {
		return se.sandbox.Execute(ctx, task)
	}

	// High trust - execute normally
	return task.Execute(ctx)
}

// Sandbox provides isolated execution environment
type Sandbox struct {
	mu sync.Mutex
}

// NewSandbox creates a new sandbox
func NewSandbox() *Sandbox {
	return &Sandbox{}
}

// Execute runs a task in a sandboxed environment
func (s *Sandbox) Execute(ctx context.Context, task Task) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Implement actual sandboxing
	// For now, just execute with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return task.Execute(ctx)
}

// CryptoSigner implements the Signer interface using SecurityManager
type CryptoSigner struct {
	sm *SecurityManager
}

// NewCryptoSigner creates a new crypto signer
func NewCryptoSigner(sm *SecurityManager) *CryptoSigner {
	return &CryptoSigner{sm: sm}
}

// Sign implements the Signer interface
func (cs *CryptoSigner) Sign(data []byte) (string, error) {
	return cs.sm.Sign(data)
}

// CryptoVerifier implements the Verifier interface using SecurityManager
type CryptoVerifier struct {
	sm *SecurityManager
}

// NewCryptoVerifier creates a new crypto verifier
func NewCryptoVerifier(sm *SecurityManager) *CryptoVerifier {
	return &CryptoVerifier{sm: sm}
}

// Verify implements the Verifier interface
func (cv *CryptoVerifier) Verify(data []byte, signature string, peerID string) error {
	return cv.sm.Verify(data, signature, peerID)
}

// WitnessProtocol implements witness-based verification
type WitnessProtocol struct {
	network      *P2PNetwork
	cache        *VerifiableCache
	minWitnesses int
	mu           sync.RWMutex
}

// NewWitnessProtocol creates a new witness protocol
func NewWitnessProtocol(network *P2PNetwork, cache *VerifiableCache, minWitnesses int) *WitnessProtocol {
	wp := &WitnessProtocol{
		network:      network,
		cache:        cache,
		minWitnesses: minWitnesses,
	}

	// Register witness handlers
	network.RegisterHandler(MessageTypeWitnessRequest, wp.handleWitnessRequest)
	network.RegisterHandler(MessageTypeWitnessResponse, wp.handleWitnessResponse)

	return wp
}

// RequestWitness requests witness verification from peers
func (wp *WitnessProtocol) RequestWitness(ctx context.Context, key string, value []byte) error {
	// Create witness request
	request := map[string]interface{}{
		"key":   key,
		"value": base64.StdEncoding.EncodeToString(value),
		"hash":  wp.calculateHash(value),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal witness request: %w", err)
	}

	// Broadcast to all peers
	return wp.network.Broadcast(&Message{
		Type:      MessageTypeWitnessRequest,
		ID:        generateMessageID(),
		From:      wp.network.nodeID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// handleWitnessRequest processes witness verification requests
func (wp *WitnessProtocol) handleWitnessRequest(ctx context.Context, msg *Message, peer *Peer) error {
	var request map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &request); err != nil {
		return err
	}

	// Verify the hash
	valueStr, _ := request["value"].(string)
	value, _ := base64.StdEncoding.DecodeString(valueStr)
	expectedHash, _ := request["hash"].(string)

	if wp.calculateHash(value) != expectedHash {
		return fmt.Errorf("hash verification failed")
	}

	// Send witness response
	response := map[string]interface{}{
		"key":      request["key"],
		"hash":     expectedHash,
		"verified": true,
		"witness":  wp.network.nodeID,
	}

	payload, _ := json.Marshal(response)

	return wp.network.SendMessage(peer, &Message{
		Type:      MessageTypeWitnessResponse,
		ID:        generateMessageID(),
		From:      wp.network.nodeID,
		To:        peer.ID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// handleWitnessResponse processes witness verification responses
func (wp *WitnessProtocol) handleWitnessResponse(ctx context.Context, msg *Message, peer *Peer) error {
	// TODO: Aggregate witness responses
	return nil
}

// calculateHash computes SHA256 hash
func (wp *WitnessProtocol) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
