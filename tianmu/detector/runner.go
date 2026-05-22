package detector

import (
	"context"
	"errors"
	"fmt"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

type Runner struct {
	normalizer *sanitize.Normalizer
	detectors  []LLMDetector
}

func NewRunner(normalizer *sanitize.Normalizer, detectors ...LLMDetector) (*Runner, error) {
	if normalizer == nil {
		return nil, errors.New("normalizer is required")
	}
	seen := map[string]struct{}{}
	for _, detector := range detectors {
		if detector == nil {
			return nil, errors.New("detector is required")
		}
		if detector.ID() == "" {
			return nil, errors.New("detector id is required")
		}
		if detector.Category() == "" {
			return nil, errors.New("detector category is required")
		}
		if detector.Version() == "" {
			return nil, errors.New("detector version is required")
		}
		if _, ok := seen[detector.ID()]; ok {
			return nil, fmt.Errorf("duplicate detector id: %s", detector.ID())
		}
		seen[detector.ID()] = struct{}{}
	}

	return &Runner{
		normalizer: normalizer,
		detectors:  append([]LLMDetector(nil), detectors...),
	}, nil
}

func (r *Runner) Detect(ctx context.Context, prompt string) ([]core.DetectorSignal, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	normalizedPrompt := r.normalizer.Normalize(prompt)
	signals := make([]core.DetectorSignal, 0, len(r.detectors))

	for _, detector := range r.detectors {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		signal, err := detector.Detect(ctx, normalizedPrompt)
		if err != nil {
			return nil, fmt.Errorf("detector %s failed: %w", detector.ID(), err)
		}
		if signal.DetectorID == "" {
			signal.DetectorID = detector.ID()
		}
		if signal.Category == "" {
			signal.Category = detector.Category()
		}
		if signal.Version == "" {
			signal.Version = detector.Version()
		}
		signals = append(signals, signal)
	}

	return signals, nil
}
