#\!/bin/bash
cgpt -b openai -m gpt-4 'What is the capital of France?'
cgpt -b openai -m gpt-4 'What is the capital of Japan?'
cgpt -b openai -m gpt-4 'What is the capital of Brazil?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of France?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of Japan?'
cgpt -b anthropic -m claude-3-opus 'What is the capital of Brazil?'
