package errors

import (
	"github.com/tmc/pe/internal/observability"
)

// Integration reports and logs errors at package boundaries.
type Integration struct {
	reporter *ErrorReporter
	logger   observability.Logger
}

// NewIntegration returns an error integration using reporter and logger.
func NewIntegration(reporter *ErrorReporter, logger observability.Logger) *Integration {
	if reporter == nil {
		reporter = NewReporter()
	}
	if logger == nil {
		logger = observability.GetGlobalLogger()
	}
	return &Integration{
		reporter: reporter,
		logger:   logger,
	}
}

// Reporter returns the integration's reporter.
func (i *Integration) Reporter() *ErrorReporter {
	if i == nil {
		return nil
	}
	return i.reporter
}

// Handle reports and logs err, returning it unchanged.
func (i *Integration) Handle(err error) error {
	return i.record(err)
}

// Provider wraps, reports, and logs provider errors.
func (i *Integration) Provider(provider, model string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := preserveOrWrap(err, func(error) error {
		return WrapProvider(err, provider, model)
	})
	addContext(wrapped, "provider", provider)
	addContext(wrapped, "model", model)
	return i.record(wrapped)
}

// Command wraps, reports, and logs command errors.
func (i *Integration) Command(command string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := preserveOrWrap(err, func(err error) error {
		return Wrap(err, ErrCodeInternal, "command failed").WithComponent("command")
	})
	addContext(wrapped, "command", command)
	return i.record(wrapped)
}

// Optimization wraps, reports, and logs optimization errors.
func (i *Integration) Optimization(method string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := preserveOrWrap(err, func(error) error {
		return WrapOptimization(err, method)
	})
	addContext(wrapped, "method", method)
	return i.record(wrapped)
}

// Evaluation wraps, reports, and logs evaluation errors.
func (i *Integration) Evaluation(suite string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := preserveOrWrap(err, func(error) error {
		return WrapEvaluation(err, suite)
	})
	addContext(wrapped, "suite", suite)
	return i.record(wrapped)
}

// Module wraps, reports, and logs module errors.
func (i *Integration) Module(module string, err error) error {
	if err == nil {
		return nil
	}
	wrapped := preserveOrWrap(err, func(error) error {
		return WrapModule(err, module, "")
	})
	if peErr := asPEError(wrapped); peErr != nil && peErr.Component == "" {
		peErr.WithComponent("module")
	}
	addContext(wrapped, "module", module)
	return i.record(wrapped)
}

func (i *Integration) record(err error) error {
	if err == nil {
		return nil
	}
	if i == nil {
		return err
	}
	if i.reporter != nil {
		i.reporter.Report(err)
	}
	if i.logger != nil {
		i.logger.Error("error reported",
			observability.String("code", string(GetCode(err))),
			observability.String("component", GetComponent(err)),
			observability.String("severity", string(GetSeverity(err))),
			observability.Bool("retryable", IsRetryable(err)),
			observability.Any("context", GetContext(err)),
			observability.Error(err),
		)
	}
	return err
}

func preserveOrWrap(err error, wrap func(error) error) error {
	if asPEError(err) != nil {
		return err
	}
	return wrap(err)
}

func addContext(err error, key string, value interface{}) {
	if value == nil || value == "" {
		return
	}
	if peErr := asPEError(err); peErr != nil {
		peErr.WithContext(key, value)
	}
}
