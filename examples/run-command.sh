#!/bin/bash
# Example usage of PE's run command with inference API

# Simple prompt
pe run "What is 2+2?"

# From file
echo "Explain quantum computing in simple terms" > prompt.txt
pe run prompt.txt

# With variables
pe run "Translate {{.Text}} to {{.Language}}" \
  --var Text="Hello world" \
  --var Language="French"

# With specific temperature (using default model)
pe run "Write a haiku about coding" \
  --temperature 0.9

# Streaming mode
pe run "Tell me a story about a robot" \
  --stream \
  --max-tokens 200

# With system prompt
pe run "Explain recursion" \
  --system "You are a computer science teacher. Use simple examples."

# Using different providers (when implemented)
# pe run "Hello" --provider ollama --model llama2
# pe run "Hello" --provider anthropic --model claude-3-haiku

# Pipeline usage
echo "What are the main benefits of Go?" | pe run -
pe run "List 5 programming languages" | grep -i python

# From gist (when implemented)
# pe run gist:username/my-prompt