#!/bin/bash

# This is a demo script for the benchmark feature
# It runs a benchmark against multiple providers using the sample configuration

# Build the pe command
go build .

# Set up environment variables if needed
# export PE_USE_CGPT=true

# Run the benchmark with text format output
./pe benchmark example/benchmark-config.yaml --iterations 2 --format text

# Run with CSV output to a file
# ./pe benchmark example/benchmark-config.yaml --iterations 3 --format csv --output benchmark-results.csv

# Run with JSON output for detailed analysis
# ./pe benchmark example/benchmark-config.yaml --iterations 5 --format json --output benchmark-results.json