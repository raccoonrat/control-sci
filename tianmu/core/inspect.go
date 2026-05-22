package core

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	signals := make([]DetectorSignal, len(detectors))
	errs := make([]error, len(detectors))
	for _, detector := range detectors {
		if detector == nil {
			return nil, errors.New("detector is required")
		}
	}

	var wg sync.WaitGroup
	for idx, detector := range detectors {
		wg.Add(1)
		go func(idx int, detector DetectorProxy) {
			defer wg.Done()
			if err := ctx.Err(); err != nil {
				errs[idx] = err
				return
			}
			signal, err := detector.Detect(ctx, normalizedPrompt)
			if err != nil {
				errs[idx] = fmt.Errorf("detector %s failed: %w", detector.ID(), err)
				return
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
			signals[idx] = signal
		}(idx, detector)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}

	return e.MediateInbound(req, identity, data, action, signals)
}
