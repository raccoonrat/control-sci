package regression

import (
	"errors"
	"os"
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
	"github.com/raccoonrat/control-sci/tianmu/sanitize"
)

func TestLoadTC260JSONL(t *testing.T) {
	cases, err := LoadTC260JSONL("../../datasets/tc260/dataset_v6/dataset_tiny.jsonl", 3)
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tc260 dataset is not available")
	}
	if err != nil {
		t.Fatalf("load tc260 jsonl: %v", err)
	}

	if len(cases) != 3 {
		t.Fatalf("cases = %d, want 3", len(cases))
	}
	for _, testCase := range cases {
		if testCase.ExpectedBehavior == "" {
			t.Fatal("expected behavior must not be empty")
		}
		if testCase.Prompt == "" {
			t.Fatal("prompt must not be empty")
		}
	}
}

func TestRunTC260TinyDataset(t *testing.T) {
	cases, err := LoadTC260JSONL("../../datasets/tc260/dataset_v6/dataset_tiny.jsonl", 0)
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("tc260 dataset is not available")
	}
	if err != nil {
		t.Fatalf("load tc260 jsonl: %v", err)
	}

	evaluator, err := core.NewEvaluator(core.PolicyPack{
		Version: "tc260-oracle-policy-v1",
		Rules: []core.PolicyRule{
			{
				RiskCategory:        TC260RefusalCategory,
				ConfidenceThreshold: 0.80,
				TargetDecision:      core.Block,
				ReasonCode:          "tc260_expected_refusal",
			},
		},
	})
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}
	engine, err := core.NewEngine(core.PersonalAI, evaluator)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	results, summary, err := RunTC260Cases(engine, sanitize.NewNormalizer(), cases)
	if err != nil {
		t.Fatalf("run tc260 cases: %v", err)
	}

	if summary.Total != len(cases) {
		t.Fatalf("summary total = %d, want %d", summary.Total, len(cases))
	}
	if summary.Failed != 0 {
		t.Fatalf("summary failed = %d, want 0; first failure: %s", summary.Failed, firstFailure(results))
	}
	if summary.Passed != len(cases) {
		t.Fatalf("summary passed = %d, want %d", summary.Passed, len(cases))
	}
}

func firstFailure(results []TC260Result) string {
	for _, result := range results {
		if !result.Passed {
			return result.Failure
		}
	}

	return ""
}
