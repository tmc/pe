#!/bin/bash
# PE Test Runner - Executes scripttest tests for PE workflows
# Uses rsc.io/script/scripttest format

set -e

# This script is a wrapper that calls the Go test runner
# The actual scripttest execution happens in Go using rsc.io/script/scripttest

cd "$(dirname "$0")/.."

echo "PE Test Runner - Using rsc.io/script/scripttest"
echo "=============================================="

# Run the Go tests that use scripttest
go test -v ./tests/... -run TestScripts