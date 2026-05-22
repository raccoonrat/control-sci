package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

type DetectorProxy interface {
	ID() string
	Category() string
	Version() string
	Detect(ctx context.Context, normalizedPrompt string) (DetectorSignal, error)
}

func (e *Engine) InspectAndMediate(
	ctx context.Context,
	normalizer *sanitize.Normalizer,
	detectors []DetectorProxy,
	req RequestContext,
	identity IdentityContext,
	data DataContext,
	action ActionContext,
	rawPrompt string,
) (*ControlDecisionObject, error) {
	if normalizer == nil {
		return nil, errors.New("normalizer is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedPrompt := normalizer.Normalize(rawPrompt)
	signals := make([]DetectorSignal, 0, len(detectors))
	for _, detector := range detectors {
		if detector == nil {
			return nil, errors.New("detector is required")
		}
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

	return e.MediateInbound(req, identity, data, action, signals)
}
