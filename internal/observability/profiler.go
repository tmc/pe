package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// ProfileType represents different types of profiling
type ProfileType string

const (
	CPUProfile    ProfileType = "cpu"
	MemoryProfile ProfileType = "memory"
	GoProfile     ProfileType = "goroutine"
	BlockProfile  ProfileType = "block"
	MutexProfile  ProfileType = "mutex"
)

// ProfileData contains profiling information
type ProfileData struct {
	Type      ProfileType   `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Filename  string        `json:"filename"`
	Stats     ProfileStats  `json:"stats"`
}

// ProfileStats contains runtime statistics
type ProfileStats struct {
	// Memory stats
	Alloc         uint64 `json:"alloc"`          // Allocated bytes
	TotalAlloc    uint64 `json:"total_alloc"`    // Total allocated bytes
	Sys           uint64 `json:"sys"`            // System bytes
	Lookups       uint64 `json:"lookups"`        // Pointer lookups
	Mallocs       uint64 `json:"mallocs"`        // Malloc operations
	Frees         uint64 `json:"frees"`          // Free operations
	HeapAlloc     uint64 `json:"heap_alloc"`     // Heap allocated bytes
	HeapSys       uint64 `json:"heap_sys"`       // Heap system bytes
	HeapIdle      uint64 `json:"heap_idle"`      // Heap idle bytes
	HeapInuse     uint64 `json:"heap_inuse"`     // Heap in-use bytes
	HeapReleased  uint64 `json:"heap_released"`  // Heap released bytes
	HeapObjects   uint64 `json:"heap_objects"`   // Heap objects
	StackInuse    uint64 `json:"stack_inuse"`    // Stack in-use bytes
	StackSys      uint64 `json:"stack_sys"`      // Stack system bytes
	
	// GC stats
	NextGC       uint64  `json:"next_gc"`        // Next GC target
	LastGC       uint64  `json:"last_gc"`        // Last GC time (ns)
	PauseTotalNs uint64  `json:"pause_total_ns"` // Total GC pause time
	PauseNs      []uint64 `json:"pause_ns"`       // Recent GC pause times
	NumGC        uint32  `json:"num_gc"`         // Number of GC cycles
	GCCPUFraction float64 `json:"gc_cpu_fraction"` // GC CPU fraction
	
	// Runtime stats
	NumGoroutines int `json:"num_goroutines"` // Number of goroutines
	NumCPU        int `json:"num_cpu"`        // Number of CPUs
}

// Profiler handles performance profiling
type Profiler struct {
	mu       sync.RWMutex
	enabled  bool
	profiles map[ProfileType]*ProfileData
	outputDir string
}

// NewProfiler creates a new profiler
func NewProfiler(outputDir string) *Profiler {
	if outputDir == "" {
		outputDir = "profiles"
	}
	
	// Ensure output directory exists
	os.MkdirAll(outputDir, 0755)
	
	return &Profiler{
		enabled:   false,
		profiles:  make(map[ProfileType]*ProfileData),
		outputDir: outputDir,
	}
}

// Start starts profiling for the specified type
func (p *Profiler) Start(profileType ProfileType) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.enabled {
		return fmt.Errorf("profiler is disabled")
	}
	
	timestamp := time.Now()
	filename := fmt.Sprintf("%s/%s_%s.prof", p.outputDir, profileType, timestamp.Format("20060102_150405"))
	
	switch profileType {
	case CPUProfile:
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create CPU profile file: %v", err)
		}
		
		if err := pprof.StartCPUProfile(file); err != nil {
			file.Close()
			return fmt.Errorf("failed to start CPU profile: %v", err)
		}
		
		p.profiles[CPUProfile] = &ProfileData{
			Type:      CPUProfile,
			Timestamp: timestamp,
			Filename:  filename,
		}
		
	case MemoryProfile:
		// Memory profiling is snapshot-based, handled in Stop()
		p.profiles[MemoryProfile] = &ProfileData{
			Type:      MemoryProfile,
			Timestamp: timestamp,
			Filename:  filename,
		}
		
	case GoProfile:
		p.profiles[GoProfile] = &ProfileData{
			Type:      GoProfile,
			Timestamp: timestamp,
			Filename:  filename,
		}
		
	case BlockProfile:
		runtime.SetBlockProfileRate(1)
		p.profiles[BlockProfile] = &ProfileData{
			Type:      BlockProfile,
			Timestamp: timestamp,
			Filename:  filename,
		}
		
	case MutexProfile:
		runtime.SetMutexProfileFraction(1)
		p.profiles[MutexProfile] = &ProfileData{
			Type:      MutexProfile,
			Timestamp: timestamp,
			Filename:  filename,
		}
		
	default:
		return fmt.Errorf("unsupported profile type: %s", profileType)
	}
	
	return nil
}

// Stop stops profiling for the specified type
func (p *Profiler) Stop(profileType ProfileType) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	profileData, exists := p.profiles[profileType]
	if !exists {
		return fmt.Errorf("profile type %s not started", profileType)
	}
	
	profileData.Duration = time.Since(profileData.Timestamp)
	profileData.Stats = p.collectStats()
	
	switch profileType {
	case CPUProfile:
		pprof.StopCPUProfile()
		
	case MemoryProfile:
		file, err := os.Create(profileData.Filename)
		if err != nil {
			return fmt.Errorf("failed to create memory profile file: %v", err)
		}
		defer file.Close()
		
		runtime.GC() // Force GC before memory profile
		if err := pprof.WriteHeapProfile(file); err != nil {
			return fmt.Errorf("failed to write memory profile: %v", err)
		}
		
	case GoProfile:
		file, err := os.Create(profileData.Filename)
		if err != nil {
			return fmt.Errorf("failed to create goroutine profile file: %v", err)
		}
		defer file.Close()
		
		if err := pprof.Lookup("goroutine").WriteTo(file, 0); err != nil {
			return fmt.Errorf("failed to write goroutine profile: %v", err)
		}
		
	case BlockProfile:
		file, err := os.Create(profileData.Filename)
		if err != nil {
			return fmt.Errorf("failed to create block profile file: %v", err)
		}
		defer file.Close()
		
		if err := pprof.Lookup("block").WriteTo(file, 0); err != nil {
			return fmt.Errorf("failed to write block profile: %v", err)
		}
		runtime.SetBlockProfileRate(0)
		
	case MutexProfile:
		file, err := os.Create(profileData.Filename)
		if err != nil {
			return fmt.Errorf("failed to create mutex profile file: %v", err)
		}
		defer file.Close()
		
		if err := pprof.Lookup("mutex").WriteTo(file, 0); err != nil {
			return fmt.Errorf("failed to write mutex profile: %v", err)
		}
		runtime.SetMutexProfileFraction(0)
	}
	
	// Save profile metadata
	metaFilename := profileData.Filename + ".json"
	if err := p.saveProfileMetadata(profileData, metaFilename); err != nil {
		return fmt.Errorf("failed to save profile metadata: %v", err)
	}
	
	delete(p.profiles, profileType)
	return nil
}

// CollectSnapshot collects a snapshot of current runtime statistics
func (p *Profiler) CollectSnapshot() ProfileStats {
	return p.collectStats()
}

// collectStats collects runtime statistics
func (p *Profiler) collectStats() ProfileStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return ProfileStats{
		// Memory stats
		Alloc:        memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		Sys:          memStats.Sys,
		Lookups:      memStats.Lookups,
		Mallocs:      memStats.Mallocs,
		Frees:        memStats.Frees,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
		HeapReleased: memStats.HeapReleased,
		HeapObjects:  memStats.HeapObjects,
		StackInuse:   memStats.StackInuse,
		StackSys:     memStats.StackSys,
		
		// GC stats
		NextGC:        memStats.NextGC,
		LastGC:        memStats.LastGC,
		PauseTotalNs:  memStats.PauseTotalNs,
		PauseNs:       memStats.PauseNs[:],
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction,
		
		// Runtime stats
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
	}
}

// saveProfileMetadata saves profile metadata to a JSON file
func (p *Profiler) saveProfileMetadata(profileData *ProfileData, filename string) error {
	data, err := json.MarshalIndent(profileData, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// Enable enables the profiler
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable disables the profiler
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled returns whether the profiler is enabled
func (p *Profiler) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// GetActiveProfiles returns currently active profiles
func (p *Profiler) GetActiveProfiles() []ProfileType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var active []ProfileType
	for profileType := range p.profiles {
		active = append(active, profileType)
	}
	
	return active
}

// ProfiledFunction wraps a function with profiling
func (p *Profiler) ProfiledFunction(ctx context.Context, name string, profileTypes []ProfileType, fn func(context.Context) error) error {
	if !p.enabled {
		return fn(ctx)
	}
	
	// Start profiling
	for _, profileType := range profileTypes {
		if err := p.Start(profileType); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start %s profile: %v\n", profileType, err)
		}
	}
	
	// Add profiling info to trace
	AddEvent(ctx, "profiling_started", fmt.Sprintf("Started profiling: %v", profileTypes), "info")
	
	// Execute function
	err := fn(ctx)
	
	// Stop profiling
	for _, profileType := range profileTypes {
		if stopErr := p.Stop(profileType); stopErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop %s profile: %v\n", profileType, stopErr)
		}
	}
	
	AddEvent(ctx, "profiling_stopped", fmt.Sprintf("Stopped profiling: %v", profileTypes), "info")
	
	return err
}

// MemoryProfiler is a simpler interface for memory monitoring
type MemoryProfiler struct {
	baseline ProfileStats
	samples  []ProfileStats
	interval time.Duration
	stopChan chan struct{}
	mu       sync.RWMutex
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(interval time.Duration) *MemoryProfiler {
	return &MemoryProfiler{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start starts memory monitoring
func (mp *MemoryProfiler) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	profiler := NewProfiler("")
	mp.baseline = profiler.collectStats()
	mp.samples = []ProfileStats{mp.baseline}
	
	go mp.monitor()
}

// Stop stops memory monitoring
func (mp *MemoryProfiler) Stop() {
	close(mp.stopChan)
}

// GetReport returns a memory usage report
func (mp *MemoryProfiler) GetReport() MemoryReport {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	if len(mp.samples) == 0 {
		return MemoryReport{}
	}
	
	latest := mp.samples[len(mp.samples)-1]
	
	return MemoryReport{
		Baseline:       mp.baseline,
		Current:        latest,
		Samples:        len(mp.samples),
		PeakAlloc:      mp.findPeak(),
		TotalAllocated: latest.TotalAlloc - mp.baseline.TotalAlloc,
		GCCycles:       int(latest.NumGC - mp.baseline.NumGC),
	}
}

// monitor runs the monitoring loop
func (mp *MemoryProfiler) monitor() {
	ticker := time.NewTicker(mp.interval)
	defer ticker.Stop()
	
	profiler := NewProfiler("")
	
	for {
		select {
		case <-ticker.C:
			stats := profiler.collectStats()
			mp.mu.Lock()
			mp.samples = append(mp.samples, stats)
			mp.mu.Unlock()
			
		case <-mp.stopChan:
			return
		}
	}
}

// findPeak finds the peak memory allocation
func (mp *MemoryProfiler) findPeak() uint64 {
	var peak uint64
	for _, sample := range mp.samples {
		if sample.Alloc > peak {
			peak = sample.Alloc
		}
	}
	return peak
}

// MemoryReport contains a memory usage report
type MemoryReport struct {
	Baseline       ProfileStats `json:"baseline"`
	Current        ProfileStats `json:"current"`
	Samples        int          `json:"samples"`
	PeakAlloc      uint64       `json:"peak_alloc"`
	TotalAllocated uint64       `json:"total_allocated"`
	GCCycles       int          `json:"gc_cycles"`
}

// Global profiler instance
var globalProfiler *Profiler
var profilerOnce sync.Once

// GetGlobalProfiler returns the global profiler instance
func GetGlobalProfiler() *Profiler {
	profilerOnce.Do(func() {
		globalProfiler = NewProfiler("profiles")
	})
	return globalProfiler
}

// InitGlobalProfiler initializes the global profiler
func InitGlobalProfiler(outputDir string) {
	globalProfiler = NewProfiler(outputDir)
}