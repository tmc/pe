package errors

import (
	"context"
	stderrors "errors"
	"net"
	"net/url"
	"os"
	"syscall"
	"testing"
	"time"
)

type timeoutNetError struct{}

func (timeoutNetError) Error() string   { return "timeout" }
func (timeoutNetError) Timeout() bool   { return true }
func (timeoutNetError) Temporary() bool { return true }

type plainNetError struct{}

func (plainNetError) Error() string   { return "network unavailable" }
func (plainNetError) Timeout() bool   { return false }
func (plainNetError) Temporary() bool { return true }

func TestWrapNilErrors(t *testing.T) {
	for _, fn := range []func() error{
		func() error { return WrapProvider(nil, "p", "m") },
		func() error { return WrapInference(nil, "m") },
		func() error { return WrapOptimization(nil, "algo") },
		func() error { return WrapEvaluation(nil, "case") },
		func() error { return WrapFile(nil, "p", "read") },
		func() error { return WrapModule(nil, "m", "v") },
		func() error { return WrapSecurity(nil, "check") },
		func() error { return WrapNetwork(nil, "dial") },
		func() error { return WrapContext(nil, "op") },
		func() error { return Annotate(nil, "k", "v") },
	} {
		if err := fn(); err != nil {
			t.Fatalf("nil wrapper returned %v", err)
		}
	}
}

func TestWrapProviderInferenceOptimizationEvaluation(t *testing.T) {
	for _, tt := range []struct {
		err  error
		code ErrorCode
	}{
		{stderrors.New("invalid api key"), ErrCodeProviderAuth},
		{stderrors.New("rate limit exceeded"), ErrCodeProviderRateLimit},
		{stderrors.New("quota exhausted"), ErrCodeProviderQuota},
		{stderrors.New("deadline exceeded"), ErrCodeProviderTimeout},
		{stderrors.New("service unavailable"), ErrCodeProviderUnavailable},
		{stderrors.New("other"), ErrCodeProviderAPI},
		{timeoutNetError{}, ErrCodeProviderTimeout},
	} {
		err := WrapProvider(tt.err, "openai", "gpt")
		if codeOf(err) != tt.code {
			t.Fatalf("provider %q code = %#v", tt.err, err)
		}
	}
	for _, tt := range []struct {
		err  error
		code ErrorCode
	}{
		{stderrors.New("context length too long"), ErrCodeInferenceContextLen},
		{stderrors.New("timeout"), ErrCodeInferenceTimeout},
		{stderrors.New("malformed request"), ErrCodeInferenceInvalid},
		{stderrors.New("failed"), ErrCodeInferenceFailed},
	} {
		err := WrapInference(tt.err, "model")
		if codeOf(err) != tt.code {
			t.Fatalf("inference %q code = %#v", tt.err, err)
		}
	}
	if codeOf(WrapOptimization(stderrors.New("did not converge"), "gaso")) != ErrCodeOptimizationConverged {
		t.Fatal("optimization convergence code mismatch")
	}
	if codeOf(WrapOptimization(stderrors.New("timeout"), "gaso")) != ErrCodeOptimizationTimeout {
		t.Fatal("optimization timeout code mismatch")
	}
	if codeOf(WrapEvaluation(stderrors.New("assertion failed"), "case")) != ErrCodeAssertionFailed {
		t.Fatal("evaluation assertion code mismatch")
	}
	if codeOf(WrapEvaluation(stderrors.New("deadline"), "case")) != ErrCodeEvaluationTimeout {
		t.Fatal("evaluation timeout code mismatch")
	}
}

func TestWrapFileModuleSecurityNetworkContext(t *testing.T) {
	if codeOf(WrapFile(os.ErrNotExist, "missing", "read")) != ErrCodeFileNotFound {
		t.Fatal("not-exist code mismatch")
	}
	for _, tt := range []struct {
		err  error
		code ErrorCode
	}{
		{&os.PathError{Op: "open", Path: "x", Err: syscall.EACCES}, ErrCodeSecurityAuth},
		{&os.PathError{Op: "write", Path: "x", Err: syscall.ENOSPC}, ErrCodeFileWrite},
		{&os.PathError{Op: "open", Path: "x", Err: syscall.EMFILE}, ErrCodeFileRead},
	} {
		if got := codeOf(WrapFile(tt.err, "x", "read")); got != tt.code {
			t.Fatalf("file code = %s, want %s", got, tt.code)
		}
	}
	if codeOf(WrapFile(stderrors.New("plain"), "x", "other")) != ErrCodeFileRead {
		t.Fatal("file default code mismatch")
	}
	for _, tt := range []struct {
		err  error
		code ErrorCode
	}{
		{stderrors.New("not found 404"), ErrCodeModuleNotFound},
		{stderrors.New("dependency required"), ErrCodeModuleDependency},
		{stderrors.New("registry download failed"), ErrCodeModuleRegistry},
		{stderrors.New("invalid"), ErrCodeModuleInvalid},
	} {
		if got := codeOf(WrapModule(tt.err, "mod", "v1")); got != tt.code {
			t.Fatalf("module code = %s, want %s", got, tt.code)
		}
	}
	for _, tt := range []struct {
		err  error
		code ErrorCode
		risk string
	}{
		{stderrors.New("malicious injection"), ErrCodeSecurityValidation, "CRITICAL"},
		{stderrors.New("unauthorized"), ErrCodeSecurityAuth, "HIGH"},
		{stderrors.New("policy violation"), ErrCodeSecurityPolicy, "HIGH"},
		{stderrors.New("other"), ErrCodeSecurityValidation, "MEDIUM"},
	} {
		err := WrapSecurity(tt.err, "check")
		sec := err.(*SecurityError)
		if sec.Code != tt.code || sec.RiskLevel != tt.risk {
			t.Fatalf("security = %#v", sec)
		}
	}
	for _, tt := range []struct {
		err       error
		code      ErrorCode
		retryable bool
	}{
		{timeoutNetError{}, ErrCodeNetworkTimeout, true},
		{plainNetError{}, ErrCodeNetworkUnavailable, true},
		{&net.DNSError{}, ErrCodeNetworkUnavailable, true},
		{&url.Error{Op: "get", URL: "x", Err: stderrors.New("bad")}, ErrCodeNetworkUnavailable, true},
		{stderrors.New("plain"), ErrCodeNetworkUnavailable, false},
	} {
		var pe *PEError
		if !stderrors.As(WrapNetwork(tt.err, "get"), &pe) || pe.Code != tt.code || pe.Retryable != tt.retryable {
			t.Fatalf("network = %#v", pe)
		}
	}
	if codeOf(WrapContext(context.Canceled, "op")) != ErrCodeInternal {
		t.Fatal("canceled code mismatch")
	}
	if codeOf(WrapContext(context.DeadlineExceeded, "op")) != ErrCodeNetworkTimeout {
		t.Fatal("deadline code mismatch")
	}
	if codeOf(WrapContext(stderrors.New("other"), "op")) != ErrCodeInternal {
		t.Fatal("context default code mismatch")
	}
}

func TestChainAnnotateAndWithTimeout(t *testing.T) {
	if Chain(nil, nil) != nil {
		t.Fatal("nil chain should be nil")
	}
	one := stderrors.New("one")
	if Chain(nil, one) != one {
		t.Fatal("single chain should return original")
	}
	chained := Chain(one, stderrors.New("two"))
	var pe *PEError
	if !stderrors.As(chained, &pe) || pe.Context["error_count"] != 2 {
		t.Fatalf("chain = %#v", pe)
	}
	annotated := Annotate(one, "key", "value")
	if !stderrors.As(annotated, &pe) || pe.Context["key"] != "value" {
		t.Fatalf("annotated = %#v", pe)
	}
	annotated = Annotate(pe, "next", 2)
	if !stderrors.As(annotated, &pe) || pe.Context["next"] != 2 {
		t.Fatalf("annotated pe = %#v", pe)
	}
	if err := WithTimeout(context.Background(), time.Second, "op", func(ctx context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if codeOf(WithTimeout(context.Background(), time.Second, "op", func(ctx context.Context) error { return context.Canceled })) != ErrCodeInternal {
		t.Fatal("with timeout fn error mismatch")
	}
	err := WithTimeout(context.Background(), time.Millisecond, "slow", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if codeOf(err) != ErrCodeNetworkTimeout {
		t.Fatalf("with timeout = %v", err)
	}
}

func codeOf(err error) ErrorCode {
	type peCarrier interface {
		error
		Unwrap() error
	}
	switch e := err.(type) {
	case *PEError:
		return e.Code
	case *ProviderError:
		return e.Code
	case *InferenceError:
		return e.Code
	case *OptimizationError:
		return e.Code
	case *EvaluationError:
		return e.Code
	case *FileError:
		return e.Code
	case *ModuleError:
		return e.Code
	case *SecurityError:
		return e.Code
	}
	var pe *PEError
	if !stderrors.As(err, &pe) {
		return ""
	}
	return pe.Code
}
