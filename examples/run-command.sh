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

# Disable streaming when a single buffered response is easier to inspect
pe run "Write a haiku about coding" --stream=false

# Streaming mode
pe run "Tell me a story about a robot" --stream

# Provider-specific model selection belongs in evaluation configs.

# Pipeline usage
echo "What are the main benefits of Go?" | pe run -
pe run "List 5 programming languages" | grep -i python

# From gist (when implemented)
# pe run gist:username/my-prompt
