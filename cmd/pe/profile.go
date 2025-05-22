package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tmc/pe/internal/observability"
)

// profileCmd returns a cobra.Command for profiling and observability features
func profileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Profiling and observability tools",
		Long: `Profile provides advanced profiling and observability features including
CPU profiling, memory profiling, distributed tracing, and metrics collection.

These tools help analyze performance, identify bottlenecks, and monitor
system behavior during prompt evaluation and processing.`,
	}

	cmd.AddCommand(profileStartCmd())
	cmd.AddCommand(profileStopCmd())
	cmd.AddCommand(profileStatusCmd())
	cmd.AddCommand(profileReportCmd())
	cmd.AddCommand(traceCmd())
	cmd.AddCommand(metricsCmd())

	return cmd
}

// profileStartCmd starts profiling
func profileStartCmd() *cobra.Command {
	var profileTypes []string
	var outputDir string
	var duration string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start profiling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileStart(cmd, profileTypes, outputDir, duration)
		},
	}

	cmd.Flags().StringSliceVarP(&profileTypes, "type", "t", []string{"cpu"}, 
		"Profile types: cpu, memory, goroutine, block, mutex")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "profiles", "Output directory")
	cmd.Flags().StringVarP(&duration, "duration", "d", "", "Profile duration (e.g., 30s)")

	return cmd
}

// profileStopCmd stops profiling
func profileStopCmd() *cobra.Command {
	var profileTypes []string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop profiling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileStop(cmd, profileTypes)
		},
	}

	cmd.Flags().StringSliceVarP(&profileTypes, "type", "t", []string{}, 
		"Profile types to stop (empty = all active)")

	return cmd
}

// profileStatusCmd shows profiling status
func profileStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show profiling status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileStatus(cmd, args)
		},
	}

	return cmd
}

// profileReportCmd generates profiling reports
func profileReportCmd() *cobra.Command {
	var outputFile string
	var format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate profiling report",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileReport(cmd, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, text")

	return cmd
}

// traceCmd handles distributed tracing
func traceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Distributed tracing tools",
	}

	cmd.AddCommand(traceStartCmd())
	cmd.AddCommand(traceStopCmd())
	cmd.AddCommand(traceReportCmd())

	return cmd
}

// traceStartCmd starts tracing
func traceStartCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start distributed tracing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTraceStart(cmd, outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "traces.jsonl", "Output file")

	return cmd
}

// traceStopCmd stops tracing
func traceStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop distributed tracing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTraceStop(cmd, args)
		},
	}

	return cmd
}

// traceReportCmd generates trace reports
func traceReportCmd() *cobra.Command {
	var traceFile string
	var outputFile string
	var format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate trace report",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTraceReport(cmd, traceFile, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&traceFile, "trace-file", "t", "traces.jsonl", "Trace file to analyze")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, text")

	return cmd
}

// metricsCmd handles metrics collection
func metricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Metrics collection tools",
	}

	cmd.AddCommand(metricsStartCmd())
	cmd.AddCommand(metricsStopCmd())
	cmd.AddCommand(metricsReportCmd())

	return cmd
}

// metricsStartCmd starts metrics collection
func metricsStartCmd() *cobra.Command {
	var outputFile string
	var interval string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start metrics collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetricsStart(cmd, outputFile, interval)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "metrics.jsonl", "Output file")
	cmd.Flags().StringVarP(&interval, "interval", "i", "10s", "Collection interval")

	return cmd
}

// metricsStopCmd stops metrics collection
func metricsStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop metrics collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetricsStop(cmd, args)
		},
	}

	return cmd
}

// metricsReportCmd generates metrics reports
func metricsReportCmd() *cobra.Command {
	var outputFile string
	var format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate metrics report",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetricsReport(cmd, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, text")

	return cmd
}

// Implementation functions

func runProfileStart(cmd *cobra.Command, profileTypes []string, outputDir, durationStr string) error {
	profiler := observability.GetGlobalProfiler()
	profiler.Enable()

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Start profiling for each type
	for _, typeStr := range profileTypes {
		profileType := observability.ProfileType(typeStr)
		
		if err := profiler.Start(profileType); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to start %s profiling: %v\n", typeStr, err)
			continue
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Started %s profiling\n", typeStr)
	}

	// If duration is specified, auto-stop after duration
	if durationStr != "" {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %v", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Profiling will stop automatically after %v\n", duration)
		
		go func() {
			time.Sleep(duration)
			for _, typeStr := range profileTypes {
				profileType := observability.ProfileType(typeStr)
				if err := profiler.Stop(profileType); err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Failed to stop %s profiling: %v\n", typeStr, err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Stopped %s profiling\n", typeStr)
				}
			}
		}()
	}

	return nil
}

func runProfileStop(cmd *cobra.Command, profileTypes []string) error {
	profiler := observability.GetGlobalProfiler()

	// If no types specified, stop all active profiles
	if len(profileTypes) == 0 {
		profileTypes = make([]string, 0)
		for _, activeType := range profiler.GetActiveProfiles() {
			profileTypes = append(profileTypes, string(activeType))
		}
	}

	for _, typeStr := range profileTypes {
		profileType := observability.ProfileType(typeStr)
		
		if err := profiler.Stop(profileType); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to stop %s profiling: %v\n", typeStr, err)
			continue
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Stopped %s profiling\n", typeStr)
	}

	return nil
}

func runProfileStatus(cmd *cobra.Command, args []string) error {
	profiler := observability.GetGlobalProfiler()
	
	fmt.Fprintf(cmd.OutOrStdout(), "Profiler Status:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Enabled: %t\n", profiler.IsEnabled())
	
	activeProfiles := profiler.GetActiveProfiles()
	if len(activeProfiles) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Active Profiles:\n")
		for _, profileType := range activeProfiles {
			fmt.Fprintf(cmd.OutOrStdout(), "    - %s\n", profileType)
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  No active profiles\n")
	}

	// Show current runtime stats
	stats := profiler.CollectSnapshot()
	fmt.Fprintf(cmd.OutOrStdout(), "\nCurrent Runtime Stats:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Memory Allocated: %.2f MB\n", float64(stats.Alloc)/(1024*1024))
	fmt.Fprintf(cmd.OutOrStdout(), "  Total Allocated: %.2f MB\n", float64(stats.TotalAlloc)/(1024*1024))
	fmt.Fprintf(cmd.OutOrStdout(), "  System Memory: %.2f MB\n", float64(stats.Sys)/(1024*1024))
	fmt.Fprintf(cmd.OutOrStdout(), "  Goroutines: %d\n", stats.NumGoroutines)
	fmt.Fprintf(cmd.OutOrStdout(), "  GC Cycles: %d\n", stats.NumGC)
	fmt.Fprintf(cmd.OutOrStdout(), "  GC CPU Fraction: %.4f\n", stats.GCCPUFraction)

	return nil
}

func runProfileReport(cmd *cobra.Command, outputFile, format string) error {
	profiler := observability.GetGlobalProfiler()
	stats := profiler.CollectSnapshot()

	report := map[string]interface{}{
		"timestamp": time.Now(),
		"stats":     stats,
		"active_profiles": profiler.GetActiveProfiles(),
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		
		if outputFile != "" {
			return os.WriteFile(outputFile, data, 0644)
		}
		
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

	case "text":
		output := fmt.Sprintf(`Profiling Report
Generated: %s

Memory Statistics:
  Allocated: %.2f MB
  Total Allocated: %.2f MB
  System: %.2f MB
  Heap Allocated: %.2f MB
  Heap System: %.2f MB
  Heap In Use: %.2f MB
  Heap Idle: %.2f MB

Runtime Statistics:
  Goroutines: %d
  GC Cycles: %d
  GC CPU Fraction: %.4f
  
Active Profiles: %v
`,
			time.Now().Format(time.RFC3339),
			float64(stats.Alloc)/(1024*1024),
			float64(stats.TotalAlloc)/(1024*1024),
			float64(stats.Sys)/(1024*1024),
			float64(stats.HeapAlloc)/(1024*1024),
			float64(stats.HeapSys)/(1024*1024),
			float64(stats.HeapInuse)/(1024*1024),
			float64(stats.HeapIdle)/(1024*1024),
			stats.NumGoroutines,
			stats.NumGC,
			stats.GCCPUFraction,
			profiler.GetActiveProfiles(),
		)

		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(output), 0644)
		}
		
		fmt.Fprint(cmd.OutOrStdout(), output)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func runTraceStart(cmd *cobra.Command, outputFile string) error {
	writer, err := observability.NewFileTraceWriter(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create trace writer: %v", err)
	}

	observability.InitGlobalTracer(writer)
	fmt.Fprintf(cmd.OutOrStdout(), "Started tracing to: %s\n", outputFile)
	
	return nil
}

func runTraceStop(cmd *cobra.Command, args []string) error {
	tracer := observability.GetGlobalTracer()
	tracer.Disable()
	
	if err := tracer.Close(); err != nil {
		return fmt.Errorf("failed to close tracer: %v", err)
	}
	
	fmt.Fprintf(cmd.OutOrStdout(), "Stopped tracing\n")
	return nil
}

func runTraceReport(cmd *cobra.Command, traceFile, outputFile, format string) error {
	// For now, just indicate that trace analysis would be implemented here
	fmt.Fprintf(cmd.OutOrStdout(), "Trace analysis not yet implemented\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Trace file: %s\n", traceFile)
	
	if _, err := os.Stat(traceFile); err != nil {
		return fmt.Errorf("trace file not found: %s", traceFile)
	}
	
	// In a real implementation, this would parse the trace file and generate reports
	report := map[string]interface{}{
		"trace_file": traceFile,
		"status":     "analysis_pending",
		"message":    "Trace analysis functionality will be implemented",
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	if outputFile != "" {
		return os.WriteFile(outputFile, data, 0644)
	}
	
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runMetricsStart(cmd *cobra.Command, outputFile, intervalStr string) error {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return fmt.Errorf("invalid interval: %v", err)
	}

	writer, err := observability.NewJSONMetricsWriter(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create metrics writer: %v", err)
	}

	observability.InitGlobalMetrics(writer)
	
	fmt.Fprintf(cmd.OutOrStdout(), "Started metrics collection to: %s\n", outputFile)
	fmt.Fprintf(cmd.OutOrStdout(), "Collection interval: %v\n", interval)
	
	return nil
}

func runMetricsStop(cmd *cobra.Command, args []string) error {
	metrics := observability.GetGlobalMetrics()
	metrics.Disable()
	
	if err := metrics.Close(); err != nil {
		return fmt.Errorf("failed to close metrics: %v", err)
	}
	
	fmt.Fprintf(cmd.OutOrStdout(), "Stopped metrics collection\n")
	return nil
}

func runMetricsReport(cmd *cobra.Command, outputFile, format string) error {
	metrics := observability.GetGlobalMetrics()
	report := metrics.GenerateReport()

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		
		if outputFile != "" {
			return os.WriteFile(outputFile, data, 0644)
		}
		
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

	case "text":
		output := fmt.Sprintf(`Metrics Report
Generated: %s

Summary:
  Total Metrics: %d
  Counters: %d
  Gauges: %d
  Histograms: %d
  Summaries: %d

Counters:
`,
			report.Timestamp.Format(time.RFC3339),
			report.Summary.TotalMetrics,
			report.Summary.TotalCounters,
			report.Summary.TotalGauges,
			report.Summary.TotalHistograms,
			report.Summary.TotalSummaries,
		)

		for name, value := range report.Counters {
			output += fmt.Sprintf("  %s: %.2f\n", name, value)
		}

		output += "\nGauges:\n"
		for name, value := range report.Gauges {
			output += fmt.Sprintf("  %s: %.2f\n", name, value)
		}

		if outputFile != "" {
			return os.WriteFile(outputFile, []byte(output), 0644)
		}
		
		fmt.Fprint(cmd.OutOrStdout(), output)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}