#\!/bin/bash

# Test the example with dry run mode
echo "=== Testing with dry run mode ==="
cd example/simple
cd ../../cmd/pe && go build -o ../../example/simple/pe
cd ../../example/simple
./pe eval config.yaml --dry-run

echo
echo "=== The above commands would be executed if run without --dry-run ==="
