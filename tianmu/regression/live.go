package regression

import (
	"context"
	"fmt"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

func RunLiveDetectorsRegression(
	ctx context.Context,
	engine *core.Engine,
	normalizer *sanitize.Normalizer,
	detectors []core.DetectorProxy,
	cases []TC260Case,
) ([]TC260Result, TC260Summary, ConfusionMatrix, *Profiler, error) {
	if engine == nil {
		return nil, TC260Summary{}, ConfusionMatrix{}, nil, fmt.Errorf("engine is required")
	}
	if normalizer == nil {
		return nil, TC260Summary{}, ConfusionMatrix{}, nil, fmt.Errorf("normalizer is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	results := make([]TC260Result, 0, len(cases))
	summary := TC260Summary{Total: len(cases)}
	matrix := ConfusionMatrix{}
	profiler := NewProfiler()

	for _, testCase := range cases {
		decision, err := engine.InspectAndMediate(
			ctx,
			normalizer,
			detectors,
			core.RequestContext{ProductID: "tc260", Language: testCase.Lang, InteractionType: "user_io"},
			core.IdentityContext{ActorID: "tc260-live-detector"},
			core.DataContext{DataClassification: "public", Source: "tc260_dataset", Destination: "model_context"},
			core.ActionContext{ActionType: "generate_response"},
			testCase.Prompt,
		)
		if err != nil {
			return nil, TC260Summary{}, ConfusionMatrix{}, nil, err
		}

		actual := decision.PolicyDecision.Decision
		passed := matrix.Add(testCase.ExpectedBehavior, actual)
		profiler.ProfileCase(testCase, actual, passed)
		if passed {
			summary.Passed++
		} else {
			summary.Failed++
		}

		result := TC260Result{
			Case:       testCase,
			Decision:   actual,
			ReasonCode: decision.PolicyDecision.ReasonCode,
			Passed:     passed,
		}
		if !passed {
			result.Failure = fmt.Sprintf("expected behavior %q but got decision %q", testCase.ExpectedBehavior, actual)
		}
		results = append(results, result)
	}

	return results, summary, matrix, profiler, nil
}
