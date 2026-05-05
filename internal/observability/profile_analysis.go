package observability

import (
	"context"
	"sync"
	"time"
)

// ProfileAnalysis summarizes runtime profile data.
type ProfileAnalysis struct {
	HeapObjects   uint64   `json:"heap_objects"`
	HeapAlloc     uint64   `json:"heap_alloc"`
	GCCycles      uint32   `json:"gc_cycles"`
	GCCPUFraction float64  `json:"gc_cpu_fraction"`
	Goroutines    int      `json:"goroutines"`
	Findings      []string `json:"findings,omitempty"`
}

// AnalyzeProfileStats returns a small diagnostic summary for stats.
func AnalyzeProfileStats(stats ProfileStats) ProfileAnalysis {
	analysis := ProfileAnalysis{
		HeapObjects:   stats.HeapObjects,
		HeapAlloc:     stats.HeapAlloc,
		GCCycles:      stats.NumGC,
		GCCPUFraction: stats.GCCPUFraction,
		Goroutines:    stats.NumGoroutines,
	}
	if stats.GCCPUFraction > 0.10 {
		analysis.Findings = append(analysis.Findings, "high gc cpu fraction")
	}
	if stats.NumGoroutines > 1000 {
		analysis.Findings = append(analysis.Findings, "high goroutine count")
	}
	if stats.HeapObjects > 1_000_000 {
		analysis.Findings = append(analysis.Findings, "high heap object count")
	}
	return analysis
}

// ContinuousProfiler periodically records runtime snapshots.
type ContinuousProfiler struct {
	profiler *Profiler
	ticker   *time.Ticker
	done     chan struct{}
	mu       sync.RWMutex
	samples  []ProfileStats
}

// NewContinuousProfiler creates a continuous profiler.
func NewContinuousProfiler(profiler *Profiler, interval time.Duration) *ContinuousProfiler {
	if profiler == nil {
		profiler = NewProfiler("")
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &ContinuousProfiler{
		profiler: profiler,
		ticker:   time.NewTicker(interval),
		done:     make(chan struct{}),
	}
}

// Start begins collecting snapshots until ctx is canceled or Stop is called.
func (p *ContinuousProfiler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.mu.Lock()
				p.samples = append(p.samples, p.profiler.CollectSnapshot())
				p.mu.Unlock()
			case <-ctx.Done():
				p.Stop()
				return
			case <-p.done:
				return
			}
		}
	}()
}

// Stop stops collection.
func (p *ContinuousProfiler) Stop() {
	p.ticker.Stop()
	select {
	case <-p.done:
	default:
		close(p.done)
	}
}

// Samples returns collected snapshots.
func (p *ContinuousProfiler) Samples() []ProfileStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	samples := make([]ProfileStats, len(p.samples))
	copy(samples, p.samples)
	return samples
}
