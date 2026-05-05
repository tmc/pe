// Package config provides centralized configuration management for the PE toolkit.
// It implements hierarchical configuration lookup with support for CLI flags,
// environment variables, config files, and defaults.
package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Config represents the complete configuration schema for PE.
// It aggregates all configuration sections and provides typed access
// to all configuration values with proper defaults.
type Config struct {
	// Core application settings
	App AppConfig `yaml:"app" json:"app"`

	// Provider configurations
	Providers ProvidersConfig `yaml:"providers" json:"providers"`

	// Evaluation settings
	Eval EvalConfig `yaml:"eval" json:"eval"`

	// Module system settings
	Modules ModulesConfig `yaml:"modules" json:"modules"`

	// Metaprompting and optimization
	Optimization OptimizationConfig `yaml:"optimization" json:"optimization"`

	// Observability and monitoring
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`

	// Security settings
	Security SecurityConfig `yaml:"security" json:"security"`

	// Plugin system
	Plugins PluginsConfig `yaml:"plugins" json:"plugins"`

	// Template system
	Templates TemplatesConfig `yaml:"templates" json:"templates"`

	// Output and formatting
	Output OutputConfig `yaml:"output" json:"output"`
}

// AppConfig contains core application settings
type AppConfig struct {
	// Version information
	Version string `yaml:"version" json:"version"`

	// Log level (debug, info, warn, error)
	LogLevel string `yaml:"log_level" json:"log_level" default:"info"`

	// Enable verbose output
	Verbose bool `yaml:"verbose" json:"verbose" default:"false"`

	// Enable debug mode
	Debug bool `yaml:"debug" json:"debug" default:"false"`

	// Working directory override
	WorkDir string `yaml:"work_dir" json:"work_dir"`

	// Configuration directory
	ConfigDir string `yaml:"config_dir" json:"config_dir" default:"~/.pe"`

	// Data directory for storage
	DataDir string `yaml:"data_dir" json:"data_dir" default:"~/.pe/data"`

	// Cache directory
	CacheDir string `yaml:"cache_dir" json:"cache_dir" default:"~/.pe/cache"`

	// Default timeout for operations
	DefaultTimeout time.Duration `yaml:"default_timeout" json:"default_timeout" default:"30s"`
}

// ProvidersConfig contains settings for all LLM providers
type ProvidersConfig struct {
	// Default provider to use
	Default string `yaml:"default" json:"default" default:"openai"`

	// OpenAI configuration
	OpenAI OpenAIConfig `yaml:"openai" json:"openai"`

	// Anthropic configuration
	Anthropic AnthropicConfig `yaml:"anthropic" json:"anthropic"`

	// Ollama configuration
	Ollama OllamaConfig `yaml:"ollama" json:"ollama"`

	// CGPT configuration (CLI wrapper)
	CGPT CGPTConfig `yaml:"cgpt" json:"cgpt"`

	// Custom provider configurations
	Custom map[string]CustomProviderConfig `yaml:"custom" json:"custom"`
}

// OpenAIConfig contains OpenAI-specific settings
type OpenAIConfig struct {
	// API key (can be set via environment)
	APIKey string `yaml:"api_key" json:"api_key" env:"OPENAI_API_KEY"`

	// API base URL
	BaseURL string `yaml:"base_url" json:"base_url" default:"https://api.openai.com/v1"`

	// Organization ID
	OrganizationID string `yaml:"organization_id" json:"organization_id" env:"OPENAI_ORG_ID"`

	// Default model
	DefaultModel string `yaml:"default_model" json:"default_model" default:"gpt-4"`

	// Request timeout
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"60s"`

	// Max retries
	MaxRetries int `yaml:"max_retries" json:"max_retries" default:"3"`

	// Rate limiting
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

// AnthropicConfig contains Anthropic-specific settings
type AnthropicConfig struct {
	// API key (can be set via environment)
	APIKey string `yaml:"api_key" json:"api_key" env:"ANTHROPIC_API_KEY"`

	// API base URL
	BaseURL string `yaml:"base_url" json:"base_url" default:"https://api.anthropic.com"`

	// Default model
	DefaultModel string `yaml:"default_model" json:"default_model" default:"claude-3-haiku-20240307"`

	// Request timeout
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"60s"`

	// Max retries
	MaxRetries int `yaml:"max_retries" json:"max_retries" default:"3"`

	// Rate limiting
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

// OllamaConfig contains Ollama-specific settings
type OllamaConfig struct {
	// Host URL
	Host string `yaml:"host" json:"host" default:"http://localhost:11434"`

	// Default model
	DefaultModel string `yaml:"default_model" json:"default_model" default:"llama2"`

	// Request timeout
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"300s"`

	// Keep alive duration
	KeepAlive time.Duration `yaml:"keep_alive" json:"keep_alive" default:"5m"`
}

// CGPTConfig contains CGPT CLI wrapper settings
type CGPTConfig struct {
	// Path to cgpt binary
	BinaryPath string `yaml:"binary_path" json:"binary_path" default:"cgpt"`

	// Default arguments to pass
	DefaultArgs []string `yaml:"default_args" json:"default_args"`

	// Environment variables to set
	Environment map[string]string `yaml:"environment" json:"environment"`

	// Timeout for subprocess
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"60s"`
}

// CustomProviderConfig contains settings for custom providers
type CustomProviderConfig struct {
	// Provider type or plugin name
	Type string `yaml:"type" json:"type"`

	// Configuration specific to the provider
	Config map[string]interface{} `yaml:"config" json:"config"`

	// Whether this provider is enabled
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	// Requests per minute
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute" default:"60"`

	// Tokens per minute
	TokensPerMinute int `yaml:"tokens_per_minute" json:"tokens_per_minute" default:"100000"`

	// Burst allowance
	Burst int `yaml:"burst" json:"burst" default:"10"`
}

// EvalConfig contains evaluation system settings
type EvalConfig struct {
	// Default configuration file patterns to search
	DefaultConfigFiles []string `yaml:"default_config_files" json:"default_config_files" default:"[\"promptfooconfig.yaml\", \"pe.config.yaml\", \".pe/config.yaml\"]"`

	// Default output directory
	OutputDir string `yaml:"output_dir" json:"output_dir" default:"./results"`

	// Default timeout for evaluations
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"300s"`

	// Maximum concurrency
	MaxConcurrency int `yaml:"max_concurrency" json:"max_concurrency" default:"10"`

	// Enable progress bars
	ShowProgress bool `yaml:"show_progress" json:"show_progress" default:"true"`

	// Save to database by default
	SaveToDb bool `yaml:"save_to_db" json:"save_to_db" default:"false"`

	// Share results by default
	Share bool `yaml:"share" json:"share" default:"false"`

	// Cache evaluation results
	Cache CacheConfig `yaml:"cache" json:"cache"`
}

// ModulesConfig contains module system settings
type ModulesConfig struct {
	// Default registry URL
	Registry string `yaml:"registry" json:"registry" default:"https://registry.pe.dev"`

	// Local modules directory
	LocalDir string `yaml:"local_dir" json:"local_dir" default:"~/.pe/modules"`

	// Vendor directory for vendored modules
	VendorDir string `yaml:"vendor_dir" json:"vendor_dir" default:"./vendor"`

	// Auto-download missing modules
	AutoDownload bool `yaml:"auto_download" json:"auto_download" default:"true"`

	// Module cache settings
	Cache CacheConfig `yaml:"cache" json:"cache"`
}

// OptimizationConfig contains metaprompting and optimization settings
type OptimizationConfig struct {
	// Default optimization method
	DefaultMethod string `yaml:"default_method" json:"default_method" default:"semantic"`

	// Maximum iterations for optimization
	MaxIterations int `yaml:"max_iterations" json:"max_iterations" default:"100"`

	// Convergence threshold
	ConvergenceThreshold float64 `yaml:"convergence_threshold" json:"convergence_threshold" default:"0.01"`

	// Population size for evolutionary algorithms
	PopulationSize int `yaml:"population_size" json:"population_size" default:"50"`

	// Learning rate for gradient-based methods
	LearningRate float64 `yaml:"learning_rate" json:"learning_rate" default:"0.1"`

	// Enable parallel optimization
	Parallel bool `yaml:"parallel" json:"parallel" default:"true"`

	// Semantic optimization settings
	Semantic SemanticConfig `yaml:"semantic" json:"semantic"`

	// GASO optimization settings
	GASO GASOConfig `yaml:"gaso" json:"gaso"`

	// Evolutionary optimization settings
	Evolution EvolutionConfig `yaml:"evolution" json:"evolution"`
}

// SemanticConfig contains semantic optimization settings
type SemanticConfig struct {
	// Embedding model to use
	EmbeddingModel string `yaml:"embedding_model" json:"embedding_model" default:"text-embedding-ada-002"`

	// Similarity threshold
	SimilarityThreshold float64 `yaml:"similarity_threshold" json:"similarity_threshold" default:"0.8"`

	// Gradient step size
	StepSize float64 `yaml:"step_size" json:"step_size" default:"0.1"`
}

// GASOConfig contains GASO optimization settings
type GASOConfig struct {
	// Temperature for sampling
	Temperature float64 `yaml:"temperature" json:"temperature" default:"0.7"`

	// Top-k sampling
	TopK int `yaml:"top_k" json:"top_k" default:"40"`

	// Top-p sampling
	TopP float64 `yaml:"top_p" json:"top_p" default:"0.9"`
}

// EvolutionConfig contains evolutionary optimization settings
type EvolutionConfig struct {
	// Mutation rate
	MutationRate float64 `yaml:"mutation_rate" json:"mutation_rate" default:"0.1"`

	// Crossover rate
	CrossoverRate float64 `yaml:"crossover_rate" json:"crossover_rate" default:"0.7"`

	// Elite percentage to preserve
	ElitePercentage float64 `yaml:"elite_percentage" json:"elite_percentage" default:"0.1"`

	// Selection method
	Selection string `yaml:"selection" json:"selection" default:"tournament"`
}

// ObservabilityConfig contains observability and monitoring settings
type ObservabilityConfig struct {
	// Enable metrics collection
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics" default:"true"`

	// Enable tracing
	EnableTracing bool `yaml:"enable_tracing" json:"enable_tracing" default:"false"`

	// Metrics port
	MetricsPort int `yaml:"metrics_port" json:"metrics_port" default:"9090"`

	// Tracing endpoint
	TracingEndpoint string `yaml:"tracing_endpoint" json:"tracing_endpoint"`

	// Log output configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Log format (json, text)
	Format string `yaml:"format" json:"format" default:"text"`

	// Log output (stdout, stderr, file path)
	Output string `yaml:"output" json:"output" default:"stderr"`

	// Enable structured logging
	Structured bool `yaml:"structured" json:"structured" default:"false"`

	// Include caller information
	IncludeCaller bool `yaml:"include_caller" json:"include_caller" default:"false"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	// Enable input validation
	ValidateInput bool `yaml:"validate_input" json:"validate_input" default:"true"`

	// Enable output sanitization
	SanitizeOutput bool `yaml:"sanitize_output" json:"sanitize_output" default:"true"`

	// Maximum input size
	MaxInputSize int64 `yaml:"max_input_size" json:"max_input_size" default:"1048576"`

	// Allowed file extensions for uploads
	AllowedExtensions []string `yaml:"allowed_extensions" json:"allowed_extensions" default:"[\".yaml\", \".yml\", \".json\", \".txt\", \".md\"]"`

	// Enable API key validation
	ValidateAPIKeys bool `yaml:"validate_api_keys" json:"validate_api_keys" default:"true"`

	// TLS configuration
	TLS TLSConfig `yaml:"tls" json:"tls"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	// Enable TLS
	Enabled bool `yaml:"enabled" json:"enabled" default:"false"`

	// Certificate file path
	CertFile string `yaml:"cert_file" json:"cert_file"`

	// Key file path
	KeyFile string `yaml:"key_file" json:"key_file"`

	// CA certificate file
	CAFile string `yaml:"ca_file" json:"ca_file"`

	// Skip certificate verification
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify" default:"false"`
}

// PluginsConfig contains plugin system settings
type PluginsConfig struct {
	// Plugin directory
	Directory string `yaml:"directory" json:"directory" default:"~/.pe/plugins"`

	// Auto-discover plugins
	AutoDiscover bool `yaml:"auto_discover" json:"auto_discover" default:"true"`

	// Plugin search paths
	SearchPaths []string `yaml:"search_paths" json:"search_paths" default:"[\"/usr/local/bin\", \"/usr/bin\", \"./plugins\"]"`

	// Enabled plugins
	Enabled []string `yaml:"enabled" json:"enabled"`

	// Disabled plugins
	Disabled []string `yaml:"disabled" json:"disabled"`

	// Plugin configurations
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// TemplatesConfig contains template system settings
type TemplatesConfig struct {
	// Template directory
	Directory string `yaml:"directory" json:"directory" default:"~/.pe/templates"`

	// Default template format
	DefaultFormat string `yaml:"default_format" json:"default_format" default:"yaml"`

	// Enable template caching
	Cache bool `yaml:"cache" json:"cache" default:"true"`

	// Template search paths
	SearchPaths []string `yaml:"search_paths" json:"search_paths" default:"[\"./templates\", \"~/.pe/templates\"]"`

	// Variable delimiter style
	VariableStyle string `yaml:"variable_style" json:"variable_style" default:"mustache"`
}

// OutputConfig contains output and formatting settings
type OutputConfig struct {
	// Default output format
	Format string `yaml:"format" json:"format" default:"text"`

	// Enable colored output
	Color bool `yaml:"color" json:"color" default:"true"`

	// Enable pretty printing
	Pretty bool `yaml:"pretty" json:"pretty" default:"true"`

	// Pagination settings
	Pagination PaginationConfig `yaml:"pagination" json:"pagination"`

	// Table formatting
	Table TableConfig `yaml:"table" json:"table"`
}

// PaginationConfig contains pagination settings
type PaginationConfig struct {
	// Enable pagination
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`

	// Items per page
	PageSize int `yaml:"page_size" json:"page_size" default:"20"`

	// Show page numbers
	ShowPageNumbers bool `yaml:"show_page_numbers" json:"show_page_numbers" default:"true"`
}

// TableConfig contains table formatting settings
type TableConfig struct {
	// Table style (ascii, unicode, markdown)
	Style string `yaml:"style" json:"style" default:"unicode"`

	// Show headers
	ShowHeaders bool `yaml:"show_headers" json:"show_headers" default:"true"`

	// Enable sorting
	Sortable bool `yaml:"sortable" json:"sortable" default:"true"`

	// Maximum column width
	MaxColumnWidth int `yaml:"max_column_width" json:"max_column_width" default:"80"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	// Enable caching
	Enabled bool `yaml:"enabled" json:"enabled" default:"true"`

	// Cache directory
	Directory string `yaml:"directory" json:"directory" default:"~/.pe/cache"`

	// Cache TTL
	TTL time.Duration `yaml:"ttl" json:"ttl" default:"24h"`

	// Maximum cache size
	MaxSize int64 `yaml:"max_size" json:"max_size" default:"1073741824"`

	// Cleanup interval
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" default:"1h"`
}

// ValidateConfig validates a configuration for correctness and consistency
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate app configuration
	if err := validateAppConfig(&config.App); err != nil {
		return fmt.Errorf("app config validation failed: %w", err)
	}

	// Validate providers configuration
	if err := validateProvidersConfig(&config.Providers); err != nil {
		return fmt.Errorf("providers config validation failed: %w", err)
	}

	// Validate evaluation configuration
	if err := validateEvalConfig(&config.Eval); err != nil {
		return fmt.Errorf("eval config validation failed: %w", err)
	}

	// Validate modules configuration
	if err := validateModulesConfig(&config.Modules); err != nil {
		return fmt.Errorf("modules config validation failed: %w", err)
	}

	// Validate optimization configuration
	if err := validateOptimizationConfig(&config.Optimization); err != nil {
		return fmt.Errorf("optimization config validation failed: %w", err)
	}

	// Validate observability configuration
	if err := validateObservabilityConfig(&config.Observability); err != nil {
		return fmt.Errorf("observability config validation failed: %w", err)
	}

	// Validate security configuration
	if err := validateSecurityConfig(&config.Security); err != nil {
		return fmt.Errorf("security config validation failed: %w", err)
	}

	// Validate plugins configuration
	if err := validatePluginsConfig(&config.Plugins); err != nil {
		return fmt.Errorf("plugins config validation failed: %w", err)
	}

	// Validate templates configuration
	if err := validateTemplatesConfig(&config.Templates); err != nil {
		return fmt.Errorf("templates config validation failed: %w", err)
	}

	// Validate output configuration
	if err := validateOutputConfig(&config.Output); err != nil {
		return fmt.Errorf("output config validation failed: %w", err)
	}

	return nil
}

// validateAppConfig validates the application configuration
func validateAppConfig(config *AppConfig) error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if config.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if strings.EqualFold(config.LogLevel, level) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level: %s (must be one of: %s)",
				config.LogLevel, strings.Join(validLogLevels, ", "))
		}
	}

	// Validate timeout
	if config.DefaultTimeout < 0 {
		return fmt.Errorf("default timeout must be non-negative, got: %v", config.DefaultTimeout)
	}

	return nil
}

// validateProvidersConfig validates the providers configuration
func validateProvidersConfig(config *ProvidersConfig) error {
	// Validate default provider
	if config.Default == "" {
		return fmt.Errorf("default provider cannot be empty")
	}

	validProviders := []string{"openai", "anthropic", "ollama", "cgpt"}
	valid := false
	for _, provider := range validProviders {
		if strings.EqualFold(config.Default, provider) {
			valid = true
			break
		}
	}
	if !valid {
		if custom, ok := config.Custom[config.Default]; ok && custom.Enabled {
			valid = true
		}
	}
	if !valid {
		return fmt.Errorf("invalid default provider: %s (must be one of: %s)",
			config.Default, strings.Join(validProviders, ", "))
	}

	// Validate OpenAI configuration
	if err := validateOpenAIConfig(&config.OpenAI); err != nil {
		return fmt.Errorf("OpenAI config validation failed: %w", err)
	}

	// Validate Anthropic configuration
	if err := validateAnthropicConfig(&config.Anthropic); err != nil {
		return fmt.Errorf("Anthropic config validation failed: %w", err)
	}

	// Validate Ollama configuration
	if err := validateOllamaConfig(&config.Ollama); err != nil {
		return fmt.Errorf("Ollama config validation failed: %w", err)
	}

	// Validate CGPT configuration
	if err := validateCGPTConfig(&config.CGPT); err != nil {
		return fmt.Errorf("CGPT config validation failed: %w", err)
	}

	for name, custom := range config.Custom {
		if err := validateCustomProviderConfig(name, &custom); err != nil {
			return err
		}
	}

	return nil
}

// validateOpenAIConfig validates OpenAI configuration
func validateOpenAIConfig(config *OpenAIConfig) error {
	if config.BaseURL != "" {
		if err := validateURL(config.BaseURL); err != nil {
			return fmt.Errorf("base url: %w", err)
		}
	}

	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", config.Timeout)
	}

	// Validate max retries
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative, got: %d", config.MaxRetries)
	}

	// Validate rate limit
	return validateRateLimitConfig(&config.RateLimit)
}

// validateAnthropicConfig validates Anthropic configuration
func validateAnthropicConfig(config *AnthropicConfig) error {
	if config.BaseURL != "" {
		if err := validateURL(config.BaseURL); err != nil {
			return fmt.Errorf("base url: %w", err)
		}
	}

	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", config.Timeout)
	}

	// Validate max retries
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative, got: %d", config.MaxRetries)
	}

	// Validate rate limit
	return validateRateLimitConfig(&config.RateLimit)
}

// validateOllamaConfig validates Ollama configuration
func validateOllamaConfig(config *OllamaConfig) error {
	if config.Host != "" {
		if err := validateURL(config.Host); err != nil {
			return fmt.Errorf("host: %w", err)
		}
	}

	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", config.Timeout)
	}

	// Validate keep alive
	if config.KeepAlive < 0 {
		return fmt.Errorf("keep alive must be non-negative, got: %v", config.KeepAlive)
	}

	return nil
}

// validateCGPTConfig validates CGPT configuration
func validateCGPTConfig(config *CGPTConfig) error {
	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", config.Timeout)
	}

	return nil
}

func validateCustomProviderConfig(name string, config *CustomProviderConfig) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("custom provider name cannot be empty")
	}
	if !config.Enabled {
		return nil
	}
	if strings.TrimSpace(config.Type) == "" {
		return fmt.Errorf("custom provider %s type cannot be empty", name)
	}
	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("must include scheme and host")
	}
	return nil
}

// validateRateLimitConfig validates rate limit configuration
func validateRateLimitConfig(config *RateLimitConfig) error {
	if config.RequestsPerMinute < 0 {
		return fmt.Errorf("requests per minute must be non-negative, got: %d", config.RequestsPerMinute)
	}

	if config.TokensPerMinute < 0 {
		return fmt.Errorf("tokens per minute must be non-negative, got: %d", config.TokensPerMinute)
	}

	if config.Burst < 0 {
		return fmt.Errorf("burst must be non-negative, got: %d", config.Burst)
	}

	return nil
}

// validateEvalConfig validates evaluation configuration
func validateEvalConfig(config *EvalConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", config.Timeout)
	}

	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive, got: %d", config.MaxConcurrency)
	}

	return validateCacheConfig(&config.Cache)
}

// validateModulesConfig validates modules configuration
func validateModulesConfig(config *ModulesConfig) error {
	return validateCacheConfig(&config.Cache)
}

// validateOptimizationConfig validates optimization configuration
func validateOptimizationConfig(config *OptimizationConfig) error {
	validMethods := []string{"semantic", "gaso", "evolution", "gradient", "random"}
	if config.DefaultMethod != "" {
		valid := false
		for _, method := range validMethods {
			if strings.EqualFold(config.DefaultMethod, method) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid default optimization method: %s (must be one of: %s)",
				config.DefaultMethod, strings.Join(validMethods, ", "))
		}
	}

	if config.MaxIterations <= 0 {
		return fmt.Errorf("max iterations must be positive, got: %d", config.MaxIterations)
	}

	if config.ConvergenceThreshold < 0 || config.ConvergenceThreshold > 1 {
		return fmt.Errorf("convergence threshold must be between 0 and 1, got: %f", config.ConvergenceThreshold)
	}

	if config.PopulationSize <= 0 {
		return fmt.Errorf("population size must be positive, got: %d", config.PopulationSize)
	}

	if config.LearningRate <= 0 {
		return fmt.Errorf("learning rate must be positive, got: %f", config.LearningRate)
	}

	return nil
}

// validateObservabilityConfig validates observability configuration
func validateObservabilityConfig(config *ObservabilityConfig) error {
	if config.MetricsPort <= 0 || config.MetricsPort > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535, got: %d", config.MetricsPort)
	}

	return validateLoggingConfig(&config.Logging)
}

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(config *LoggingConfig) error {
	validFormats := []string{"json", "text", "logfmt"}
	if config.Format != "" {
		valid := false
		for _, format := range validFormats {
			if strings.EqualFold(config.Format, format) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log format: %s (must be one of: %s)",
				config.Format, strings.Join(validFormats, ", "))
		}
	}

	validOutputs := []string{"stdout", "stderr"}
	if config.Output != "" && !strings.HasPrefix(config.Output, "/") {
		valid := false
		for _, output := range validOutputs {
			if strings.EqualFold(config.Output, output) {
				valid = true
				break
			}
		}
		if !valid && !strings.HasPrefix(config.Output, "/") {
			return fmt.Errorf("invalid log output: %s (must be stdout, stderr, or a file path)", config.Output)
		}
	}

	return nil
}

// validateSecurityConfig validates security configuration
func validateSecurityConfig(config *SecurityConfig) error {
	if config.MaxInputSize <= 0 {
		return fmt.Errorf("max input size must be positive, got: %d", config.MaxInputSize)
	}

	return validateTLSConfig(&config.TLS)
}

// validateTLSConfig validates TLS configuration
func validateTLSConfig(config *TLSConfig) error {
	if config.Enabled {
		if config.CertFile == "" {
			return fmt.Errorf("cert file must be specified when TLS is enabled")
		}
		if config.KeyFile == "" {
			return fmt.Errorf("key file must be specified when TLS is enabled")
		}
	}

	return nil
}

// validatePluginsConfig validates plugins configuration
func validatePluginsConfig(config *PluginsConfig) error {
	// No specific validation needed for plugins config currently
	return nil
}

// validateTemplatesConfig validates templates configuration
func validateTemplatesConfig(config *TemplatesConfig) error {
	validFormats := []string{"yaml", "yml", "json", "toml"}
	if config.DefaultFormat != "" {
		valid := false
		for _, format := range validFormats {
			if strings.EqualFold(config.DefaultFormat, format) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid default template format: %s (must be one of: %s)",
				config.DefaultFormat, strings.Join(validFormats, ", "))
		}
	}

	validStyles := []string{"mustache", "golang", "handlebars"}
	if config.VariableStyle != "" {
		valid := false
		for _, style := range validStyles {
			if strings.EqualFold(config.VariableStyle, style) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid variable style: %s (must be one of: %s)",
				config.VariableStyle, strings.Join(validStyles, ", "))
		}
	}

	return nil
}

// validateOutputConfig validates output configuration
func validateOutputConfig(config *OutputConfig) error {
	validFormats := []string{"text", "json", "yaml", "table", "csv"}
	if config.Format != "" {
		valid := false
		for _, format := range validFormats {
			if strings.EqualFold(config.Format, format) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid output format: %s (must be one of: %s)",
				config.Format, strings.Join(validFormats, ", "))
		}
	}

	if err := validatePaginationConfig(&config.Pagination); err != nil {
		return fmt.Errorf("pagination config validation failed: %w", err)
	}

	return validateTableConfig(&config.Table)
}

// validatePaginationConfig validates pagination configuration
func validatePaginationConfig(config *PaginationConfig) error {
	if config.PageSize <= 0 {
		return fmt.Errorf("page size must be positive, got: %d", config.PageSize)
	}

	return nil
}

// validateTableConfig validates table configuration
func validateTableConfig(config *TableConfig) error {
	validStyles := []string{"ascii", "unicode", "markdown", "simple"}
	if config.Style != "" {
		valid := false
		for _, style := range validStyles {
			if strings.EqualFold(config.Style, style) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid table style: %s (must be one of: %s)",
				config.Style, strings.Join(validStyles, ", "))
		}
	}

	if config.MaxColumnWidth <= 0 {
		return fmt.Errorf("max column width must be positive, got: %d", config.MaxColumnWidth)
	}

	return nil
}

// validateCacheConfig validates cache configuration
func validateCacheConfig(config *CacheConfig) error {
	if config.TTL < 0 {
		return fmt.Errorf("TTL must be non-negative, got: %v", config.TTL)
	}

	if config.MaxSize <= 0 {
		return fmt.Errorf("max size must be positive, got: %d", config.MaxSize)
	}

	if config.CleanupInterval < 0 {
		return fmt.Errorf("cleanup interval must be non-negative, got: %v", config.CleanupInterval)
	}

	return nil
}

// DefaultConfig returns a configuration with all default values set
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			LogLevel:       "info",
			Verbose:        false,
			Debug:          false,
			ConfigDir:      "~/.pe",
			DataDir:        "~/.pe/data",
			CacheDir:       "~/.pe/cache",
			DefaultTimeout: 30 * time.Second,
		},
		Providers: ProvidersConfig{
			Default: "openai",
			OpenAI: OpenAIConfig{
				BaseURL:      "https://api.openai.com/v1",
				DefaultModel: "gpt-4",
				Timeout:      60 * time.Second,
				MaxRetries:   3,
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
					TokensPerMinute:   100000,
					Burst:             10,
				},
			},
			Anthropic: AnthropicConfig{
				BaseURL:      "https://api.anthropic.com",
				DefaultModel: "claude-3-haiku-20240307",
				Timeout:      60 * time.Second,
				MaxRetries:   3,
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
					TokensPerMinute:   100000,
					Burst:             10,
				},
			},
			Ollama: OllamaConfig{
				Host:         "http://localhost:11434",
				DefaultModel: "llama2",
				Timeout:      300 * time.Second,
				KeepAlive:    5 * time.Minute,
			},
			CGPT: CGPTConfig{
				BinaryPath: "cgpt",
				Timeout:    60 * time.Second,
			},
		},
		Eval: EvalConfig{
			DefaultConfigFiles: []string{"promptfooconfig.yaml", "pe.config.yaml", ".pe/config.yaml"},
			OutputDir:          "./results",
			Timeout:            300 * time.Second,
			MaxConcurrency:     10,
			ShowProgress:       true,
			SaveToDb:           false,
			Share:              false,
			Cache: CacheConfig{
				Enabled:         true,
				Directory:       "~/.pe/cache",
				TTL:             24 * time.Hour,
				MaxSize:         1 << 30, // 1GB
				CleanupInterval: time.Hour,
			},
		},
		Modules: ModulesConfig{
			Registry:     "https://registry.pe.dev",
			LocalDir:     "~/.pe/modules",
			VendorDir:    "./vendor",
			AutoDownload: true,
			Cache: CacheConfig{
				Enabled:         true,
				Directory:       "~/.pe/cache",
				TTL:             24 * time.Hour,
				MaxSize:         1 << 30, // 1GB
				CleanupInterval: time.Hour,
			},
		},
		Optimization: OptimizationConfig{
			DefaultMethod:        "semantic",
			MaxIterations:        100,
			ConvergenceThreshold: 0.01,
			PopulationSize:       50,
			LearningRate:         0.1,
			Parallel:             true,
			Semantic: SemanticConfig{
				EmbeddingModel:      "text-embedding-ada-002",
				SimilarityThreshold: 0.8,
				StepSize:            0.1,
			},
			GASO: GASOConfig{
				Temperature: 0.7,
				TopK:        40,
				TopP:        0.9,
			},
			Evolution: EvolutionConfig{
				MutationRate:    0.1,
				CrossoverRate:   0.7,
				ElitePercentage: 0.1,
				Selection:       "tournament",
			},
		},
		Observability: ObservabilityConfig{
			EnableMetrics: true,
			EnableTracing: false,
			MetricsPort:   9090,
			Logging: LoggingConfig{
				Format:        "text",
				Output:        "stderr",
				Structured:    false,
				IncludeCaller: false,
			},
		},
		Security: SecurityConfig{
			ValidateInput:     true,
			SanitizeOutput:    true,
			MaxInputSize:      1 << 20, // 1MB
			AllowedExtensions: []string{".yaml", ".yml", ".json", ".txt", ".md"},
			ValidateAPIKeys:   true,
			TLS: TLSConfig{
				Enabled:            false,
				InsecureSkipVerify: false,
			},
		},
		Plugins: PluginsConfig{
			Directory:    "~/.pe/plugins",
			AutoDiscover: true,
			SearchPaths:  []string{"/usr/local/bin", "/usr/bin", "./plugins"},
		},
		Templates: TemplatesConfig{
			Directory:     "~/.pe/templates",
			DefaultFormat: "yaml",
			Cache:         true,
			SearchPaths:   []string{"./templates", "~/.pe/templates"},
			VariableStyle: "mustache",
		},
		Output: OutputConfig{
			Format: "text",
			Color:  true,
			Pretty: true,
			Pagination: PaginationConfig{
				Enabled:         true,
				PageSize:        20,
				ShowPageNumbers: true,
			},
			Table: TableConfig{
				Style:          "unicode",
				ShowHeaders:    true,
				Sortable:       true,
				MaxColumnWidth: 80,
			},
		},
	}
}
