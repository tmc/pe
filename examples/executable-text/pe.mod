module example.com/executable-text

pe 1

metadata (
	owner release
	purpose review
)

policy (
	data allow release-notes test-results
	prompts allow ./plain.prompt ./templated.prompt ./workflow.pe.yaml
	providers allow local
	tools allow "pe cat"
	placement run local
	placement write tmp
)

# Child artifacts may narrow this policy, but cannot loosen it.
