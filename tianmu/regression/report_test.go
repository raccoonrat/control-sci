package regression

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

func TestBuildTC260ReportAggregatesWithoutPromptLeakage(t *testing.T) {
	results := []TC260Result{
		{
			Case: TC260Case{
				ID:               "attack-1",
				Prompt:           "敏感攻击样本不应进入报告",
				TC260Category:    "A.1.a",
				ExpectedBehavior: "refuse",
				Source:           "mutation_llm",
				Lang:             "zh",
				Attributes:       map[string]any{"difficulty": "high"},
			},
			Decision:   core.Block,
			ReasonCode: "tc260_expected_refusal",
			Passed:     true,
		},
		{
			Case: TC260Case{
				ID:               "safe-1",
				Prompt:           "安全正例也不应进入报告",
				ExpectedBehavior: "answer",
				Source:           "manual",
				Lang:             "zh",
				Attributes:       map[string]any{"difficulty": "easy"},
			},
			Decision:   core.Block,
			ReasonCode: "tc260_expected_refusal",
			Passed:     false,
			Failure:    "expected behavior \"answer\" but got decision \"block\"",
		},
	}
	summary := TC260Summary{Total: 2, Passed: 1, Failed: 1}

	report := BuildTC260Report("datasets/tc260/dataset_v6/dataset_tiny.jsonl", nil, results, summary)
	if report.Summary.Failed != 1 {
		t.Fatalf("failed = %d, want 1", report.Summary.Failed)
	}
	if report.ByCategory["uncategorized"] != 1 {
		t.Fatalf("uncategorized count = %d, want 1", report.ByCategory["uncategorized"])
	}
	if len(report.FailureExamples) != 1 {
		t.Fatalf("failure examples = %d, want 1", len(report.FailureExamples))
	}

	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	if strings.Contains(string(payload), "敏感攻击样本") || strings.Contains(string(payload), "安全正例") {
		t.Fatal("report must not include raw prompts")
	}
}
