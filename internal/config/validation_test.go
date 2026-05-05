package config

import (
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				App: AppConfig{
					LogLevel: "invalid",
				},
				Providers: ProvidersConfig{
					Default: "openai",
				},
				Eval: EvalConfig{
					MaxConcurrency: 1,
					Cache:          CacheConfig{MaxSize: 1},
				},
				Modules: ModulesConfig{
					Cache: CacheConfig{MaxSize: 1},
				},
				Optimization: OptimizationConfig{
					MaxIterations:  1,
					PopulationSize: 1,
					LearningRate:   0.1,
				},
				Observability: ObservabilityConfig{
					MetricsPort: 8080,
				},
				Security: SecurityConfig{
					MaxInputSize: 1024,
				},
				Output: OutputConfig{
					Pagination: PaginationConfig{PageSize: 10},
					Table:      TableConfig{MaxColumnWidth: 80},
				},
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "negative timeout",
			config: &Config{
				App: AppConfig{
					DefaultTimeout: -1 * time.Second,
				},
				Providers: ProvidersConfig{
					Default: "openai",
				},
				Eval: EvalConfig{
					MaxConcurrency: 1,
					Cache:          CacheConfig{MaxSize: 1},
				},
				Modules: ModulesConfig{
					Cache: CacheConfig{MaxSize: 1},
				},
				Optimization: OptimizationConfig{
					MaxIterations:  1,
					PopulationSize: 1,
					LearningRate:   0.1,
				},
				Observability: ObservabilityConfig{
					MetricsPort: 8080,
				},
				Security: SecurityConfig{
					MaxInputSize: 1024,
				},
				Output: OutputConfig{
					Pagination: PaginationConfig{PageSize: 10},
					Table:      TableConfig{MaxColumnWidth: 80},
				},
			},
			expectError: true,
			errorMsg:    "timeout must be non-negative",
		},
		{
			name: "invalid provider",
			config: &Config{
				App: AppConfig{},
				Providers: ProvidersConfig{
					Default: "invalid",
				},
				Eval: EvalConfig{
					MaxConcurrency: 1,
					Cache:          CacheConfig{MaxSize: 1},
				},
				Modules: ModulesConfig{
					Cache: CacheConfig{MaxSize: 1},
				},
				Optimization: OptimizationConfig{
					MaxIterations:  1,
					PopulationSize: 1,
					LearningRate:   0.1,
				},
				Observability: ObservabilityConfig{
					MetricsPort: 8080,
				},
				Security: SecurityConfig{
					MaxInputSize: 1024,
				},
				Output: OutputConfig{
					Pagination: PaginationConfig{PageSize: 10},
					Table:      TableConfig{MaxColumnWidth: 80},
				},
			},
			expectError: true,
			errorMsg:    "invalid default provider",
		},
		{
			name: "zero max concurrency",
			config: &Config{
				App: AppConfig{},
				Providers: ProvidersConfig{
					Default: "openai",
				},
				Eval: EvalConfig{
					MaxConcurrency: 0,
					Cache:          CacheConfig{MaxSize: 1},
				},
				Modules: ModulesConfig{
					Cache: CacheConfig{MaxSize: 1},
				},
				Optimization: OptimizationConfig{
					MaxIterations:  1,
					PopulationSize: 1,
					LearningRate:   0.1,
				},
				Observability: ObservabilityConfig{
					MetricsPort: 8080,
				},
				Security: SecurityConfig{
					MaxInputSize: 1024,
				},
				Output: OutputConfig{
					Pagination: PaginationConfig{PageSize: 10},
					Table:      TableConfig{MaxColumnWidth: 80},
				},
			},
			expectError: true,
			errorMsg:    "max concurrency must be positive",
		},
		{
			name: "invalid metrics port",
			config: &Config{
				App: AppConfig{},
				Providers: ProvidersConfig{
					Default: "openai",
				},
				Eval: EvalConfig{
					MaxConcurrency: 1,
					Cache:          CacheConfig{MaxSize: 1},
				},
				Modules: ModulesConfig{
					Cache: CacheConfig{MaxSize: 1},
				},
				Optimization: OptimizationConfig{
					MaxIterations:  1,
					PopulationSize: 1,
					LearningRate:   0.1,
				},
				Observability: ObservabilityConfig{
					MetricsPort: 70000,
				},
				Security: SecurityConfig{
					MaxInputSize: 1024,
				},
				Output: OutputConfig{
					Pagination: PaginationConfig{PageSize: 10},
					Table:      TableConfig{MaxColumnWidth: 80},
				},
			},
			expectError: true,
			errorMsg:    "metrics port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, but got no error", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidateAppConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *AppConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      &AppConfig{LogLevel: "info", DefaultTimeout: 30 * time.Second},
			expectError: false,
		},
		{
			name:        "invalid log level",
			config:      &AppConfig{LogLevel: "invalid"},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name:        "negative timeout",
			config:      &AppConfig{DefaultTimeout: -1 * time.Second},
			expectError: true,
			errorMsg:    "timeout must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, but got no error", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
			}
		})
	}
}

func TestValidateProvidersConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *ProvidersConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      &ProvidersConfig{Default: "openai"},
			expectError: false,
		},
		{
			name:        "empty default",
			config:      &ProvidersConfig{Default: ""},
			expectError: true,
			errorMsg:    "default provider cannot be empty",
		},
		{
			name:        "invalid provider",
			config:      &ProvidersConfig{Default: "invalid"},
			expectError: true,
			errorMsg:    "invalid default provider",
		},
		{
			name: "custom default provider",
			config: &ProvidersConfig{
				Default: "local",
				Custom: map[string]CustomProviderConfig{
					"local": {Type: "openai-compatible", Enabled: true},
				},
			},
			expectError: false,
		},
		{
			name: "custom provider missing type",
			config: &ProvidersConfig{
				Default: "openai",
				Custom: map[string]CustomProviderConfig{
					"local": {Enabled: true},
				},
			},
			expectError: true,
			errorMsg:    "type cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProvidersConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, but got no error", tt.errorMsg)
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
			}
		})
	}
}
