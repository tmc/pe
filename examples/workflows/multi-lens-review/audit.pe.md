---
kind: pe.workflow.v1
name: local-audit
inputs:
  topic:
    type: string
    default: error handling
safety:
  providers:
    allow: [local]
---
phase("Audit")
questions = parallel([
    plan("List one common " + args["topic"] + " mistake in Go, in one sentence.", label="q" + str(i))
    for i in range(2)
])

result = generate("Merge into a two-item checklist:\n" + "\n".join(questions))
