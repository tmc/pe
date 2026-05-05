package metaprompt

import (
	"context"

	"github.com/tmc/pe/internal/llm"
)

// Generator is the provider surface required by metaprompt optimizers.
type Generator interface {
	Generate(context.Context, string, llm.GenerateOptions) (*llm.GenerateResponse, error)
}
