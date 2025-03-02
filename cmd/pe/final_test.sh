#\!/bin/bash

# Commands to run each prompt with each provider/model:

cgpt -b openai -m gpt-4 'What is the capital of France?'
cgpt -b openai -m gpt-4 'What is the capital of Japan?'
cgpt -b openai -m gpt-4 'What is the capital of Brazil?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of France?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of Japan?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of Brazil?'

# Save these commands to a file and run with 'bash filename' to execute them
