package observability

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfiler(t *testing.T) {
	tmpDir := t.TempDir()
	
	profiler := NewProfiler(tmpDir)
	
	assert.NotNil(t, profiler)
	assert.False(t, profiler.enabled) // Should start disabled
	assert.Equal(t, tmpDir, profiler.outputDir)
	assert.NotNil(t, profiler.profiles)
	assert.Len(t, profiler.profiles, 0)

	// Test with empty directory (should use default)
	profiler2 := NewProfiler("")
	assert.Equal(t, "profiles", profiler2.outputDir)
}

func TestProfiler_EnableDisable(t *testing.T) {
	profiler := NewProfiler(t.TempDir())
	
	// Initially disabled
	assert.False(t, profiler.IsEnabled())
	
	// Enable
	profiler.Enable()
	assert.True(t, profiler.IsEnabled())
	
	// Disable
	profiler.Disable()
	assert.False(t, profiler.IsEnabled())
}

func TestProfiler_StartStop_CPUProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	// Test starting CPU profile
	err := profiler.Start(CPUProfile)
	require.NoError(t, err)

	// Should be in active profiles
	activeProfiles := profiler.GetActiveProfiles()
	assert.Contains(t, activeProfiles, CPUProfile)

	// Simulate some CPU work
	func() {
		for i := 0; i < 100000; i++ {
			_ = i * i
		}
	}()

	// Stop profile
	err = profiler.Stop(CPUProfile)
	require.NoError(t, err)

	// Should no longer be active
	activeProfiles = profiler.GetActiveProfiles()
	assert.NotContains(t, activeProfiles, CPUProfile)

	// Check that profile file was created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "cpu_") && strings.HasSuffix(file.Name(), ".prof") {
			found = true
			break
		}
	}
	assert.True(t, found, "CPU profile file should be created")

	// Should also have metadata file
	found = false
	for _, file := range files {
		if strings.Contains(file.Name(), "cpu_") && strings.HasSuffix(file.Name(), ".prof.json") {
			found = true
			break
		}
	}
	assert.True(t, found, "CPU profile metadata file should be created")
}

func TestProfiler_StartStop_MemoryProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	// Test memory profile
	err := profiler.Start(MemoryProfile)
	require.NoError(t, err)

	// Allocate some memory
	data := make([][]byte, 1000)
	for i := range data {
		data[i] = make([]byte, 1024)
	}
	_ = data // Use the data

	err = profiler.Stop(MemoryProfile)
	require.NoError(t, err)

	// Check for profile file
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "memory_") && strings.HasSuffix(file.Name(), ".prof") {
			found = true
			break
		}
	}
	assert.True(t, found, "Memory profile file should be created")
}

func TestProfiler_StartStop_GoroutineProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	// Start some goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()
	}

	err := profiler.Start(GoProfile)
	require.NoError(t, err)

	err = profiler.Stop(GoProfile)
	require.NoError(t, err)

	// Wait for goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for profile file
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "goroutine_") && strings.HasSuffix(file.Name(), ".prof") {
			found = true
			break
		}
	}
	assert.True(t, found, "Goroutine profile file should be created")
}

func TestProfiler_StartStop_BlockProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	err := profiler.Start(BlockProfile)
	require.NoError(t, err)

	// Create some blocking operations
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- true
	}()
	<-ch

	err = profiler.Stop(BlockProfile)
	require.NoError(t, err)

	// Check for profile file
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "block_") && strings.HasSuffix(file.Name(), ".prof") {
			found = true
			break
		}
	}
	assert.True(t, found, "Block profile file should be created")
}

func TestProfiler_StartStop_MutexProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	err := profiler.Start(MutexProfile)
	require.NoError(t, err)

	// Create some mutex contention
	var mu sync.Mutex
	done := make(chan bool)
	
	for i := 0; i < 5; i++ {
		go func() {
			mu.Lock()
			time.Sleep(1 * time.Millisecond)
			mu.Unlock()
			done <- true
		}()
	}

	// Wait for goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	err = profiler.Stop(MutexProfile)
	require.NoError(t, err)

	// Check for profile file
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "mutex_") && strings.HasSuffix(file.Name(), ".prof") {
			found = true
			break
		}
	}
	assert.True(t, found, "Mutex profile file should be created")
}

func TestProfiler_ErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)

	// Test starting when disabled
	err := profiler.Start(CPUProfile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profiler is disabled")

	// Test unsupported profile type
	profiler.Enable()
	err = profiler.Start(ProfileType("unsupported"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported profile type")

	// Test stopping non-existent profile
	err = profiler.Stop(CPUProfile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")

	// Test with invalid output directory
	invalidDir := filepath.Join(tmpDir, "nonexistent", "path")
	profiler2 := NewProfiler(invalidDir)
	profiler2.Enable()
	
	// This might fail depending on the profile type
	err = profiler2.Start(MemoryProfile)
	// We don't assert error here as directory creation might succeed
}

func TestProfiler_CollectSnapshot(t *testing.T) {
	profiler := NewProfiler(t.TempDir())
	
	stats := profiler.CollectSnapshot()
	
	// Verify basic stats are populated
	assert.Greater(t, stats.Alloc, uint64(0))
	assert.Greater(t, stats.TotalAlloc, uint64(0))
	assert.Greater(t, stats.Sys, uint64(0))
	assert.GreaterOrEqual(t, stats.NumGoroutines, 1) // At least this test goroutine
	assert.Equal(t, runtime.NumCPU(), stats.NumCPU)
	
	// GC stats might be zero in a fresh test
	assert.GreaterOrEqual(t, stats.NumGC, uint32(0))
}

func TestProfiler_ProfiledFunction(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	ctx := context.Background()
	profileTypes := []ProfileType{MemoryProfile, GoProfile}

	// Test successful function
	err := profiler.ProfiledFunction(ctx, "test_function", profileTypes, func(ctx context.Context) error {
		// Do some work
		data := make([]byte, 1024)
		_ = data
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	assert.NoError(t, err)

	// Check that profile files were created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	
	memoryFound := false
	goroutineFound := false
	
	for _, file := range files {
		if strings.Contains(file.Name(), "memory_") {
			memoryFound = true
		}
		if strings.Contains(file.Name(), "goroutine_") {
			goroutineFound = true
		}
	}
	
	assert.True(t, memoryFound, "Memory profile should be created")
	assert.True(t, goroutineFound, "Goroutine profile should be created")

	// Test function with error
	testError := assert.AnError
	err = profiler.ProfiledFunction(ctx, "test_function_error", profileTypes, func(ctx context.Context) error {
		return testError
	})

	assert.Equal(t, testError, err)
}

func TestProfiler_ProfiledFunction_Disabled(t *testing.T) {
	profiler := NewProfiler(t.TempDir())
	// Keep profiler disabled

	ctx := context.Background()
	executed := false

	err := profiler.ProfiledFunction(ctx, "test", []ProfileType{CPUProfile}, func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed) // Function should still execute when profiler is disabled
}

func TestMemoryProfiler(t *testing.T) {
	profiler := NewMemoryProfiler(10 * time.Millisecond)
	
	profiler.Start()
	
	// Allocate some memory
	data := make([][]byte, 100)
	for i := range data {
		data[i] = make([]byte, 1024)
	}
	
	// Wait for some samples
	time.Sleep(50 * time.Millisecond)
	
	profiler.Stop()
	
	report := profiler.GetReport()
	
	// Should have collected multiple samples
	assert.Greater(t, report.Samples, 1)
	assert.Greater(t, report.Current.Alloc, uint64(0))
	assert.Greater(t, report.PeakAlloc, uint64(0))
	assert.GreaterOrEqual(t, report.Current.TotalAlloc, report.Baseline.TotalAlloc)
	
	// Should have some allocation difference
	assert.Greater(t, report.TotalAllocated, uint64(0))
}

func TestMemoryProfiler_EmptyReport(t *testing.T) {
	profiler := NewMemoryProfiler(time.Second)
	
	// Don't start the profiler
	report := profiler.GetReport()
	
	// Should return empty report
	assert.Equal(t, MemoryReport{}, report)
}

func TestGlobalProfiler(t *testing.T) {
	// Reset global state
	globalProfiler = nil
	profilerOnce = sync.Once{}

	// Test getting global profiler (should initialize)
	profiler := GetGlobalProfiler()
	assert.NotNil(t, profiler)
	assert.Equal(t, "profiles", profiler.outputDir)

	// Test that subsequent calls return the same instance
	profiler2 := GetGlobalProfiler()
	assert.Same(t, profiler, profiler2)

	// Test initializing with custom directory
	tmpDir := t.TempDir()
	InitGlobalProfiler(tmpDir)

	// Should use the new directory
	assert.Equal(t, tmpDir, globalProfiler.outputDir)
}

func TestProfileStats_JSONSerialization(t *testing.T) {
	stats := ProfileStats{
		Alloc:         1024,
		TotalAlloc:    2048,
		NumGoroutines: 5,
		NumCPU:        8,
		NumGC:         10,
	}

	// Test JSON marshaling
	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var unmarshaled ProfileStats
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, stats.Alloc, unmarshaled.Alloc)
	assert.Equal(t, stats.TotalAlloc, unmarshaled.TotalAlloc)
	assert.Equal(t, stats.NumGoroutines, unmarshaled.NumGoroutines)
	assert.Equal(t, stats.NumCPU, unmarshaled.NumCPU)
	assert.Equal(t, stats.NumGC, unmarshaled.NumGC)
}

func TestProfileData_JSONSerialization(t *testing.T) {
	data := ProfileData{
		Type:      CPUProfile,
		Timestamp: time.Now(),
		Duration:  time.Second,
		Filename:  "/tmp/cpu_profile.prof",
		Stats: ProfileStats{
			Alloc:      1024,
			TotalAlloc: 2048,
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	var unmarshaled ProfileData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, data.Type, unmarshaled.Type)
	assert.Equal(t, data.Filename, unmarshaled.Filename)
	assert.Equal(t, data.Stats.Alloc, unmarshaled.Stats.Alloc)
}

// Benchmark tests
func BenchmarkProfiler_CollectSnapshot(b *testing.B) {
	profiler := NewProfiler("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profiler.CollectSnapshot()
	}
}

func BenchmarkMemoryProfiler_Sampling(b *testing.B) {
	profiler := NewMemoryProfiler(time.Nanosecond) // Very fast sampling
	
	// Simulate the monitoring loop
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = profiler.GetReport()
	}
}

// Test concurrent access
func TestProfiler_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	profiler := NewProfiler(tmpDir)
	profiler.Enable()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start multiple goroutines that start/stop profiles
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Try different profile types
					profileType := []ProfileType{MemoryProfile, GoProfile}[id%2]
					
					if err := profiler.Start(profileType); err == nil {
						time.Sleep(10 * time.Millisecond)
						profiler.Stop(profileType)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should have created some files
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Greater(t, len(files), 0)
}