#\!/bin/bash

# Try different variations of cgpt to see what's happening
echo "=== Test 1: Basic cgpt call with model ==="
cgpt -b openai -m gpt-4 "What is the capital of France?"

echo "=== Test 2: cgpt with --debug flag ==="
cgpt -b openai -m gpt-4 --debug "What is the capital of France?"

echo "=== Test 3: cgpt with explicit temperature ==="
cgpt -b openai -m gpt-4 --temperature 0.2 "What is the capital of France?"

echo "=== Test 4: cgpt with googleai backend ==="
cgpt -b googleai -m gemini-2.0-flash "What is the capital of France?"

