# Simple Prompt Example

This example demonstrates the most basic usage of PE - running a single prompt.

## Files
- `prompt.txt` - A simple prompt asking for an explanation of machine learning

## Running the Example

```bash
# Using OpenAI
pe run prompt.txt --provider openai

# Using Anthropic
pe run prompt.txt --provider anthropic

# Using default provider (cgpt)
pe run prompt.txt

# With streaming output
pe run prompt.txt --stream

# Save output to file
pe run prompt.txt --provider openai > output.txt
```

## What to Expect

The prompt will generate a beginner-friendly explanation of machine learning in under 200 words.

## Customization

You can modify `prompt.txt` to ask any question. Try:
- Changing the topic
- Adjusting the word limit
- Adding specific requirements
- Changing the target audience