#\!/bin/bash

# Function to run a command with proper error handling
run_cgpt() {
  local backend=$1
  local model=$2
  local prompt=$3
  
  echo "# Running: $backend model $model"
  
  # Capture both stdout and stderr
  output=$(cgpt -b "$backend" -m "$model" "$prompt" 2>&1)
  status=$?
  
  # Check if command succeeded
  if [ $status -eq 0 ]; then
    echo "$output"
  else
    echo "ERROR: Failed to run cgpt with $backend model $model"
    echo "Error details: $output"
  fi
  echo
}

echo '# === openai provider with model gpt-4 ==='
echo

run_cgpt "openai" "gpt-4" "What is the capital of France?"
run_cgpt "openai" "gpt-4" "What is the capital of Japan?"
run_cgpt "openai" "gpt-4" "What is the capital of Brazil?"

echo '# === anthropic provider with model claude-3-opus ==='
echo

run_cgpt "anthropic" "claude-3-opus" "What is the capital of France?"
run_cgpt "anthropic" "claude-3-opus" "What is the capital of Japan?"
run_cgpt "anthropic" "claude-3-opus" "What is the capital of Brazil?"

echo "# === End of commands ==="
echo "# All commands have completed execution"
echo "# Any errors encountered have been reported above"
