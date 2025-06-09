package distributed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached response with verification metadata
type CacheEntry struct {
	Key       string        `json:"key"`
	Value     []byte        `json:"value"`
	Hash      string        `json:"hash"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
	Witnesses []string      `json:"witnesses"` // Peer IDs that have verified this entry
	Signature string        `json:"signature"` // Cryptographic signature
}

// WitnessData represents witness verification data
type WitnessData struct {
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
	Hash      string    `json:"hash"` // Hash of the witnessed data
}

// VerifiableCache provides cryptographically verifiable caching
type VerifiableCache struct {
	mu        sync.RWMutex
	entries   map[string]*CacheEntry
	peerID    string
	peers     []string
	signer    Signer
	verifier  Verifier
	store     *ContentAddressedStore
	witnesses map[string][]WitnessData // Key -> witness data
}

// Signer provides cryptographic signing
type Signer interface {
	Sign(data []byte) (string, error)
}

// Verifier provides signature verification
type Verifier interface {
	Verify(data []byte, signature string, peerID string) error
}

// NewVerifiableCache creates a new verifiable cache instance
func NewVerifiableCache(peerID string, signer Signer, verifier Verifier) *VerifiableCache {
	return &VerifiableCache{
		entries:   make(map[string]*CacheEntry),
		peerID:    peerID,
		signer:    signer,
		verifier:  verifier,
		store:     NewContentAddressedStore(),
		witnesses: make(map[string][]WitnessData),
	}
}

// Set adds an entry to the cache with cryptographic verification
func (c *VerifiableCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate content hash
	hash := c.calculateHash(value)

	// Create entry
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		Hash:      hash,
		Timestamp: time.Now(),
		TTL:       ttl,
		Witnesses: []string{c.peerID},
	}

	// Sign the entry data (without signature field)
	signData := map[string]interface{}{
		"key":       entry.Key,
		"value":     entry.Value,
		"hash":      entry.Hash,
		"timestamp": entry.Timestamp,
		"ttl":       entry.TTL,
		"witnesses": entry.Witnesses,
	}

	entryData, err := json.Marshal(signData)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	signature, err := c.signer.Sign(entryData)
	if err != nil {
		return fmt.Errorf("failed to sign entry: %w", err)
	}
	entry.Signature = signature

	c.entries[key] = entry

	// Broadcast to peers for witness verification
	go c.broadcastForWitness(ctx, entry)

	return nil
}

// Get retrieves an entry from the cache with verification
func (c *VerifiableCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		// Try to fetch from peers
		return c.fetchFromPeers(ctx, key)
	}

	// Check if entry is expired
	if c.isExpired(entry) {
		delete(c.entries, key)
		return nil, fmt.Errorf("cache entry expired")
	}

	// Verify content hash
	if c.calculateHash(entry.Value) != entry.Hash {
		return nil, fmt.Errorf("cache entry corrupted: hash mismatch")
	}

	// Verify signature (reconstruct sign data without signature field)
	signData := map[string]interface{}{
		"key":       entry.Key,
		"value":     entry.Value,
		"hash":      entry.Hash,
		"timestamp": entry.Timestamp,
		"ttl":       entry.TTL,
		"witnesses": entry.Witnesses,
	}

	entryData, _ := json.Marshal(signData)
	if err := c.verifier.Verify(entryData, entry.Signature, entry.Witnesses[0]); err != nil {
		return nil, fmt.Errorf("cache entry signature invalid: %w", err)
	}

	return entry.Value, nil
}

// calculateHash computes SHA256 hash of data
func (c *VerifiableCache) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// isExpired checks if a cache entry has expired
func (c *VerifiableCache) isExpired(entry *CacheEntry) bool {
	if entry.TTL == 0 {
		return false // No expiration
	}
	return time.Since(entry.Timestamp) > entry.TTL
}

// broadcastForWitness sends entry to peers for witness verification
func (c *VerifiableCache) broadcastForWitness(ctx context.Context, entry *CacheEntry) {
	// TODO: Implement P2P broadcast
	// For now, we'll just simulate witness verification
}

// fetchFromPeers attempts to retrieve a cache entry from peer nodes
func (c *VerifiableCache) fetchFromPeers(ctx context.Context, key string) ([]byte, error) {
	// TODO: Implement P2P fetch
	return nil, fmt.Errorf("key not found in cache or peers")
}

// AddPeer adds a peer to the cache network
func (c *VerifiableCache) AddPeer(peerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.peers = append(c.peers, peerID)
}

// Export exports the cache for sharing with other nodes
func (c *VerifiableCache) Export() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Filter out expired entries
	validEntries := make([]*CacheEntry, 0)
	for _, entry := range c.entries {
		if !c.isExpired(entry) {
			validEntries = append(validEntries, entry)
		}
	}

	return json.Marshal(validEntries)
}

// Import imports cache entries from another node
func (c *VerifiableCache) Import(data []byte) error {
	var entries []*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	imported := 0
	for _, entry := range entries {
		// Verify entry before importing
		if c.calculateHash(entry.Value) != entry.Hash {
			continue // Skip corrupted entries
		}

		// TODO: Verify signature against known peers

		// Only import if not expired
		if !c.isExpired(entry) {
			c.entries[entry.Key] = entry
			imported++
		}
	}

	return nil
}

// Stats returns cache statistics
func (c *VerifiableCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		TotalEntries: len(c.entries),
		PeerCount:    len(c.peers),
	}

	for _, entry := range c.entries {
		if c.isExpired(entry) {
			stats.ExpiredEntries++
		} else {
			stats.ValidEntries++
		}
		stats.TotalSize += len(entry.Value)
	}

	return stats
}

// GetWithWitnesses retrieves a cache entry along with its witness data
func (c *VerifiableCache) GetWithWitnesses(key string) (*CacheEntry, []WitnessData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, nil, fmt.Errorf("key not found: %s", key)
	}

	// Check if entry is expired
	if c.isExpired(entry) {
		return nil, nil, fmt.Errorf("cache entry expired")
	}

	// Get witness data
	witnesses := c.witnesses[key]

	return entry, witnesses, nil
}

// addWitness adds witness verification data for a cache entry
func (c *VerifiableCache) addWitness(key string, witness WitnessData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[key]; !exists {
		return // Entry doesn't exist
	}

	// Add witness data
	c.witnesses[key] = append(c.witnesses[key], witness)

	// Update entry's witness list if not already present
	entry := c.entries[key]
	found := false
	for _, w := range entry.Witnesses {
		if w == witness.PeerID {
			found = true
			break
		}
	}
	if !found {
		entry.Witnesses = append(entry.Witnesses, witness.PeerID)
	}
}

// VerifyEntry verifies the cryptographic signature of a cache entry
func (c *VerifiableCache) VerifyEntry(entry *CacheEntry) error {
	// Verify content hash
	if c.calculateHash(entry.Value) != entry.Hash {
		return fmt.Errorf("cache entry corrupted: hash mismatch")
	}

	// Verify signature (reconstruct sign data without signature field)
	signData := map[string]interface{}{
		"key":       entry.Key,
		"value":     entry.Value,
		"hash":      entry.Hash,
		"timestamp": entry.Timestamp,
		"ttl":       entry.TTL,
		"witnesses": entry.Witnesses,
	}

	entryData, err := json.Marshal(signData)
	if err != nil {
		return fmt.Errorf("failed to marshal entry data: %w", err)
	}

	// Verify against the first witness (original signer)
	if len(entry.Witnesses) == 0 {
		return fmt.Errorf("no witnesses for entry")
	}

	if err := c.verifier.Verify(entryData, entry.Signature, entry.Witnesses[0]); err != nil {
		return fmt.Errorf("cache entry signature invalid: %w", err)
	}

	return nil
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	ExpiredEntries int
	TotalSize      int
	PeerCount      int
}

// ContentAddressedStore provides content-addressed storage
type ContentAddressedStore struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

// NewContentAddressedStore creates a new content-addressed store
func NewContentAddressedStore() *ContentAddressedStore {
	return &ContentAddressedStore{
		objects: make(map[string][]byte),
	}
}

// Put stores data and returns its content address (hash)
func (s *ContentAddressedStore) Put(data []byte) string {
	hash := sha256.Sum256(data)
	address := hex.EncodeToString(hash[:])

	s.mu.Lock()
	defer s.mu.Unlock()

	s.objects[address] = data
	return address
}

// Get retrieves data by its content address
func (s *ContentAddressedStore) Get(address string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.objects[address]
	if !exists {
		return nil, fmt.Errorf("object not found: %s", address)
	}

	// Verify content matches address
	hash := sha256.Sum256(data)
	if hex.EncodeToString(hash[:]) != address {
		return nil, fmt.Errorf("content corruption detected")
	}

	return data, nil
}

// Has checks if an object exists in the store
func (s *ContentAddressedStore) Has(address string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.objects[address]
	return exists
}

// List returns all key-value pairs in the store
func (s *ContentAddressedStore) List() []struct {
	Key   string
	Value []byte
} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []struct {
		Key   string
		Value []byte
	}
	for k, v := range s.objects {
		items = append(items, struct {
			Key   string
			Value []byte
		}{Key: k, Value: v})
	}
	return items
}

// Set stores a key-value pair (used by cache sync)
func (s *ContentAddressedStore) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = data
	return nil
}

// DistributedCache provides a high-level API for the distributed cache system
type DistributedCache struct {
	path     string
	cache    *VerifiableCache
	store    *ContentAddressedStore
	security *SecurityManager
	mu       sync.RWMutex
	stats    *CacheStatsDetail
}

// CacheStatsDetail provides detailed cache statistics
type CacheStatsDetail struct {
	TotalEntries        int
	TotalSize           int64
	HitRate             float64
	MissRate            float64
	WitnessCount        int
	VerifiedEntries     int
	AvgEntrySize        int64
	OldestEntry         time.Time
	NewestEntry         time.Time
	EntryTypes          map[string]int
	WitnessDistribution map[int]int
	hits                int64
	misses              int64
}

// CacheEntryDetail provides detailed information about a cache entry
type CacheEntryDetail struct {
	Hash      string
	Type      string
	CreatedAt time.Time
	TTL       time.Duration
	Data      []byte
	Signature []byte
	Witnesses []WitnessInfo
}

// WitnessInfo contains witness verification information
type WitnessInfo struct {
	NodeID    string
	Timestamp time.Time
	Signature string
}

// VerificationResults contains cache verification results
type VerificationResults struct {
	TotalEntries     int
	ValidEntries     int
	InvalidEntries   int
	CorruptedEntries int
}

// NewDistributedCache creates a new distributed cache instance
func NewDistributedCache(path string) (*DistributedCache, error) {
	security, err := NewSecurityManager("local")
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}

	cache := NewVerifiableCache("local", security, security)
	store := NewContentAddressedStore()

	return &DistributedCache{
		path:     path,
		cache:    cache,
		store:    store,
		security: security,
		stats: &CacheStatsDetail{
			EntryTypes:          make(map[string]int),
			WitnessDistribution: make(map[int]int),
		},
	}, nil
}

// GetStats returns basic cache statistics
func (dc *DistributedCache) GetStats() (*CacheStatsDetail, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Get stats from the underlying cache
	stats := dc.cache.Stats()

	// Create fresh stats object
	result := &CacheStatsDetail{
		TotalEntries:        stats.TotalEntries,
		TotalSize:           int64(stats.TotalSize),
		VerifiedEntries:     stats.ValidEntries,
		WitnessCount:        stats.PeerCount,
		hits:                dc.stats.hits,
		misses:              dc.stats.misses,
		EntryTypes:          dc.stats.EntryTypes,
		WitnessDistribution: dc.stats.WitnessDistribution,
	}

	if result.hits+result.misses > 0 {
		result.HitRate = float64(result.hits) / float64(result.hits+result.misses)
		result.MissRate = 1.0 - result.HitRate
	}

	return result, nil
}

// List returns a list of cache entries matching the filter
func (dc *DistributedCache) List(filter string) ([]*CacheEntryDetail, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	entries := make([]*CacheEntryDetail, 0)

	// For now, return all entries
	// TODO: Implement proper filtering
	for key, entry := range dc.cache.entries {
		if !dc.cache.isExpired(entry) {
			detail := &CacheEntryDetail{
				Hash:      entry.Hash,
				Type:      "prompt", // TODO: Determine actual type
				CreatedAt: entry.Timestamp,
				TTL:       entry.TTL,
				Data:      entry.Value,
				Witnesses: make([]WitnessInfo, len(entry.Witnesses)),
			}

			for i, witness := range entry.Witnesses {
				detail.Witnesses[i] = WitnessInfo{
					NodeID:    witness,
					Timestamp: entry.Timestamp,
				}
			}

			entries = append(entries, detail)
			_ = key // TODO: Use key in filtering
		}
	}

	return entries, nil
}

// Get retrieves a cache entry by hash
func (dc *DistributedCache) Get(hash string) (*CacheEntryDetail, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Find entry by hash
	for _, entry := range dc.cache.entries {
		if entry.Hash == hash || entry.Hash[:12] == hash {
			dc.stats.hits++
			return &CacheEntryDetail{
				Hash:      entry.Hash,
				Type:      "prompt",
				CreatedAt: entry.Timestamp,
				TTL:       entry.TTL,
				Data:      entry.Value,
				Signature: []byte(entry.Signature),
				Witnesses: make([]WitnessInfo, len(entry.Witnesses)),
			}, nil
		}
	}

	dc.stats.misses++
	return nil, fmt.Errorf("entry not found")
}

// Set stores a cache entry
func (dc *DistributedCache) Set(key string, entry *CacheEntryDetail) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Use verifiable cache to set with signing
	ctx := context.Background()
	return dc.cache.Set(ctx, key, entry.Data, entry.TTL)
}

// ImportBundle imports a cache bundle
func (dc *DistributedCache) ImportBundle(data []byte, verify bool) (imported, failed int, err error) {
	// TODO: Implement proper bundle format
	if err := dc.cache.Import(data); err != nil {
		return 0, 0, err
	}

	return len(dc.cache.entries), 0, nil
}

// ExportBundle exports cache entries to a bundle
func (dc *DistributedCache) ExportBundle(filter string, sign bool) ([]byte, error) {
	// TODO: Implement proper bundle format with filtering
	return dc.cache.Export()
}

// VerifyIntegrity verifies the integrity of all cache entries
func (dc *DistributedCache) VerifyIntegrity() (*VerificationResults, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	results := &VerificationResults{
		TotalEntries: len(dc.cache.entries),
	}

	for _, entry := range dc.cache.entries {
		// Check hash
		if dc.cache.calculateHash(entry.Value) != entry.Hash {
			results.CorruptedEntries++
			continue
		}

		// Check signature
		signData := map[string]interface{}{
			"key":       entry.Key,
			"value":     entry.Value,
			"hash":      entry.Hash,
			"timestamp": entry.Timestamp,
			"ttl":       entry.TTL,
			"witnesses": entry.Witnesses,
		}

		entryData, _ := json.Marshal(signData)
		if err := dc.security.Verify(entryData, entry.Signature, entry.Witnesses[0]); err != nil {
			results.InvalidEntries++
		} else {
			results.ValidEntries++
		}
	}

	return results, nil
}

// Clear removes cache entries matching the pattern
func (dc *DistributedCache) Clear(pattern string) (int, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	cleared := 0

	if pattern == "" {
		// Clear all
		cleared = len(dc.cache.entries)
		dc.cache.entries = make(map[string]*CacheEntry)
	} else {
		// TODO: Implement pattern matching
		for key := range dc.cache.entries {
			delete(dc.cache.entries, key)
			cleared++
		}
	}

	return cleared, nil
}

// GetDetailedStats returns detailed cache statistics
func (dc *DistributedCache) GetDetailedStats() (*CacheStatsDetail, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	stats := dc.stats
	stats.TotalEntries = len(dc.cache.entries)

	var totalSize int64
	var oldestTime, newestTime time.Time

	for _, entry := range dc.cache.entries {
		size := int64(len(entry.Value))
		totalSize += size

		if oldestTime.IsZero() || entry.Timestamp.Before(oldestTime) {
			oldestTime = entry.Timestamp
		}
		if newestTime.IsZero() || entry.Timestamp.After(newestTime) {
			newestTime = entry.Timestamp
		}

		// Track witness distribution
		witnessCount := len(entry.Witnesses)
		stats.WitnessDistribution[witnessCount]++

		// Track entry types (simplified for now)
		stats.EntryTypes["prompt"]++
	}

	stats.TotalSize = totalSize
	if stats.TotalEntries > 0 {
		stats.AvgEntrySize = totalSize / int64(stats.TotalEntries)
	}
	stats.OldestEntry = oldestTime
	stats.NewestEntry = newestTime

	return stats, nil
}
