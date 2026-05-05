module example.com/prompts/basic

pe 1

capability {
	data allow public internal
	prompts allow local
	providers allow local hosted
	tools deny shell network
}

placement {
	run local
	workspace read
	network false
}

policy {
	composition strict
	require-typed-io true
	require-reviewed-imports true
}
