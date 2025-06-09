package providers

// RegisterAll registers all built-in providers
// This is called from main or test packages to avoid import cycles
func RegisterAll() {
	// The providers are already registered in init() functions
	// This function exists to ensure the package is imported
}
