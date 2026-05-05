package config

import "fmt"

// ValidationIssue describes one configuration validation failure.
type ValidationIssue struct {
	Path    string `json:"path" yaml:"path"`
	Message string `json:"message" yaml:"message"`
}

// ValidationReport summarizes configuration validation.
type ValidationReport struct {
	Valid  bool              `json:"valid" yaml:"valid"`
	Issues []ValidationIssue `json:"issues,omitempty" yaml:"issues,omitempty"`
}

// Validator validates PE configuration.
type Validator struct {
	validators []func(*Config) []ValidationIssue
}

// NewValidator returns a validator with the built-in schema checks.
func NewValidator(validators ...func(*Config) []ValidationIssue) *Validator {
	v := &Validator{
		validators: []func(*Config) []ValidationIssue{
			schemaValidationIssues,
			requiredValidationIssues,
			typeValidationIssues,
		},
	}
	v.validators = append(v.validators, validators...)
	return v
}

// Validate returns a report instead of failing at the first error.
func (v *Validator) Validate(config *Config) ValidationReport {
	var issues []ValidationIssue
	for _, validate := range v.validators {
		issues = append(issues, validate(config)...)
	}
	return ValidationReport{
		Valid:  len(issues) == 0,
		Issues: issues,
	}
}

// ValidateConfigReport validates config with the default validator.
func ValidateConfigReport(config *Config) ValidationReport {
	return NewValidator().Validate(config)
}

func schemaValidationIssues(config *Config) []ValidationIssue {
	if err := ValidateConfig(config); err != nil {
		return []ValidationIssue{{Path: "config", Message: err.Error()}}
	}
	return nil
}

func requiredValidationIssues(config *Config) []ValidationIssue {
	if config == nil {
		return nil
	}
	var issues []ValidationIssue
	if config.Providers.Default == "" {
		issues = append(issues, ValidationIssue{Path: "providers.default", Message: "required field is empty"})
	}
	if config.App.DefaultTimeout < 0 {
		issues = append(issues, ValidationIssue{Path: "app.default_timeout", Message: "duration must be non-negative"})
	}
	if config.Eval.MaxConcurrency <= 0 {
		issues = append(issues, ValidationIssue{Path: "eval.max_concurrency", Message: "value must be positive"})
	}
	return issues
}

func typeValidationIssues(config *Config) []ValidationIssue {
	if config == nil {
		return nil
	}
	var issues []ValidationIssue
	if config.Output.Format == "" {
		issues = append(issues, ValidationIssue{Path: "output.format", Message: "format must be a string value"})
	}
	if config.Optimization.LearningRate <= 0 {
		issues = append(issues, ValidationIssue{Path: "optimization.learning_rate", Message: fmt.Sprintf("value must be positive, got %g", config.Optimization.LearningRate)})
	}
	return issues
}
