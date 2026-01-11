# Pipelines

PE supports Unix-style composition, allowing you to chain prompts and data operations.

## The Pipe Operator

Just like standard Unix tools (`grep`, `sed`), PE commands read from Stdin and write to Stdout.

```bash
cat data.txt | pe ask "Summarize" | pe ask "Translate to French"
```

## Stream Processing Tools

*   `pe ask`: Runs a prompt for each line of input or the entire buffer.
*   `pe filter`: Filters JSON objects in a stream based on criteria.
*   `pe reduce`: Aggregates a stream of data into a single result (e.g., summarize all reviews).

## Example: Feedback Analysis Pipeline

```bash
cat customer_reviews.db \
  | pe ask "Extract sentiment and key topics (JSON)" \
  | pe filter --query "sentiment == 'negative'" \
  | pe reduce "Summarize the top 3 complaints"
```
