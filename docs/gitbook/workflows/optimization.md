# Optimization

PE implements state-of-the-art algorithms to automatically improve your prompts.

## Available Methods

1.  **PE2 (Prompt Engineering Squared)**: Uses an LLM to iteratively refine the prompt based on reasoning and persona injection.
2.  **APEX**: Evolutionary algorithm for optimizing long system prompts.
3.  **TextGrad**: Semantic backpropagation based on textual gradients (2025 research).

## Usage

Use the `pe optimize` command:

```bash
pe optimize --prompt-file my-prompt.txt --method pe2 --iterations 5
```

## Choosing a Method

*   **PE2**: Best for general reasoning tasks. Adds structure and personas.
*   **APEX**: Best for compressing and refining long, complex system prompts.
*   **TextGrad**: Best for fine-tuning semantic nuances.

## Example

Optimizing a simple "Analyze sentiment" prompt with PE2 often yields a sophisticated prompt with:
*   Expert Persona ("You are a sentiment analyst...")
*   Step-by-step reasoning chain.
*   Strict output formatting (JSON/XML).
