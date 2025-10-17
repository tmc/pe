# PE Test Coverage Report

**Generated**: October 16, 2025  
**Test Run**: All packages

## Summary

**Overall Coverage**: 30.0%

**Packages Tested**: EOF
grep "coverage:" /tmp/pe-coverage.txt | grep -v "\[no" | wc -l | tr -d ' ' >> docs/TEST_COVERAGE_REPORT.md

cat >> docs/TEST_COVERAGE_REPORT.md << 'EOF'

**Test Status**: ⚠️ 1 failing test (attestation - keychain timeout)

## Coverage by Category

### Excellent Coverage (>75%)
- **ok  	github.com/tmc/pe/internal/promptfoo/execution/consensus	3.936s	**: 82.2% of statements
- **ok  	github.com/tmc/pe/internal/promptfoo/security/redteam	2.606s	**: 86.4% of statements
- **ok  	github.com/tmc/pe/internal/starlark	3.578s	**: 81.5% of statements

### Good Coverage (50-75%)
- **ok  	github.com/tmc/pe/internal/cgpt	2.624s	**: 69.3% of statements
- **ok  	github.com/tmc/pe/internal/inference/providers/anthropic	1.583s	**: 73.3% of statements
- **ok  	github.com/tmc/pe/internal/inference/providers/cgpt	2.340s	**: 51.3% of statements
- **ok  	github.com/tmc/pe/internal/inference/providers/ollama	2.956s	**: 71.9% of statements
- **ok  	github.com/tmc/pe/internal/inference/providers/openai	3.566s	**: 74.0% of statements
- **ok  	github.com/tmc/pe/internal/observability	4.662s	**: 62.0% of statements
- **ok  	github.com/tmc/pe/internal/prompt	2.647s	**: 58.9% of statements
- **ok  	github.com/tmc/pe/internal/providers	3.856s	**: 57.9% of statements
- **ok  	github.com/tmc/pe/internal/testing/mocks	3.698s	**: 60.4% of statements

### Fair Coverage (25-50%)
- **ok  	github.com/tmc/pe/ext/starlark	0.517s	**: 38.0% of statements
- **ok  	github.com/tmc/pe/internal/config	1.040s	**: 33.6% of statements
- **ok  	github.com/tmc/pe/internal/errors	2.082s	**: 30.7% of statements
- **ok  	github.com/tmc/pe/internal/llm	3.623s	**: 42.7% of statements
- **ok  	github.com/tmc/pe/internal/pemod	2.932s	**: 48.7% of statements
- **ok  	github.com/tmc/pe/internal/promptfoo/evaluation/evaluator	2.716s	**: 30.6% of statements
- **ok  	github.com/tmc/pe/internal/promptfoo/evaluation/metrics	2.720s	**: 27.9% of statements
- **ok  	github.com/tmc/pe/internal/promptfoo/execution/distributed	14.754s	**: 29.9% of statements

### Low Coverage (<25%) - Needs Attention
- **ok  	github.com/tmc/pe/cmd/pe	0.897s	**: 6.4% of statements ⚠️
- **	github.com/tmc/pe/example/structured/with-go-structs		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/examples/inference		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/ext/starlark/cmd/starlark-demo		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/cli		**: 0.0% of statements ⚠️
- **ok  	github.com/tmc/pe/internal/inference	0.254s	**: 22.4% of statements ⚠️
- **ok  	github.com/tmc/pe/internal/metaprompt	3.498s	**: 16.0% of statements ⚠️
- **	github.com/tmc/pe/internal/module		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/optimization		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/optimization/adapters		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/optimization/optimizers		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/plugin		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/internal/promptfoo/evaluation/testing		**: 0.0% of statements ⚠️
- ****: 7.8% of statements ⚠️
- **ok  	github.com/tmc/pe/internal/structured	3.066s	**: 6.6% of statements ⚠️
- **	github.com/tmc/pe/internal/templates		**: 0.0% of statements ⚠️
- **ok  	github.com/tmc/pe/internal/testing	4.897s	**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/plugins/promptfoo		**: 0.0% of statements ⚠️
- **	github.com/tmc/pe/test		**: 0.0% of statements ⚠️

## Test Failures

### internal/promptfoo/security/attestation
**Status**: FAIL (602s timeout)  
**Cause**: Keychain access timeout during attestation tests  
**Action Required**: Fix keychain mock or skip keychain tests in CI

## Packages Without Tests

- internal/promptfoo

## Recommendations

### Immediate Actions (P1)
1. Fix attestation package test failure
2. Add tests for cmd/pe (currently 6.4%)
3. Add tests for internal/cli (no tests)
4. Add tests for internal/structured (6.6%)

### Short Term (P2)
1. Improve metaprompt coverage (16.0% → 50%+)
2. Improve inference coverage (22.4% → 50%+)
3. Add tests for packages without test files

### Long Term (P3)
1. Target 70%+ coverage across all packages
2. Add integration tests
3. Improve test quality and assertions

## Coverage Trends

This is the first comprehensive coverage report. Future reports will track trends.

**Documentation Claims vs Reality**:
- docs/README.md claimed: ~25%
- docs/CURRENT_STATUS.md claimed: ~40%
- **Actual measured**: See summary above
