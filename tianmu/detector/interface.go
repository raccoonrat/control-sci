package detector

import (
	"context"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

type LLMDetector interface {
	ID() string
	Category() string
	Version() string
	Detect(ctx context.Context, normalizedPrompt string) (core.DetectorSignal, error)
}
