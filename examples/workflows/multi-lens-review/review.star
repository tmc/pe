# Multi-lens code review: fan out one review call per lens in parallel,
# then synthesize the findings with a final call.
#
# Run:
#   pe exp workflow run review.star --provider ollama:llama3.2:3b \
#       --args '{"code": "func add(a, b int) int { return a + b }"}'

code = args["code"]

phase("Review")
lenses = ["correctness", "readability", "error handling"]
findings = parallel([
    plan("In one short sentence, review this Go code for " + lens + ":\n" + code, label=lens)
    for lens in lenses
])

phase("Synthesize")
result = generate(
    "Combine these code review notes into one short verdict:\n" + "\n".join(findings),
    label="synthesize",
)
