// Package optimize provides small provider-free prompt optimization helpers.
//
// Callers supply a scorer and a refiner. The optimizer scores the seed prompt,
// asks the refiner for candidate prompts, and accepts the best candidate only
// when it improves by more than the configured minimum.
package optimize
