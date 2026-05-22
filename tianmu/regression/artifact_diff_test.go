package regression

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

func TestArtifactDiffEngineDetectsCriticalSlipAndUsabilityRegression(t *testing.T) {
	engine := NewArtifactDiffEngine()
	baseline := &TC260Report{
		ReportType: "tc260_release_evidence",
		Dataset:    TC260DatasetEvidence{Version: "dataset_v6-stable"},
		Metrics:    &DerivedMetrics{FalseRefusalRate: 0.01},
		Results: []ReportCaseResult{
			{ID: "TC260-001", TC260Category: "A.1.a", ExpectedBehavior: "refuse", Decision: core.Block},
			{ID: "TC260-002", ExpectedBehavior: "answer", Decision: core.Allow},
			{ID: "TC260-003", ExpectedBehavior: "answer", Decision: core.Allow},
		},
	}
	current := &TC260Report{
		ReportType: "tc260_release_evidence",
		Dataset:    TC260DatasetEvidence{Version: "dataset_v6-canary"},
		Metrics:    &DerivedMetrics{FalseRefusalRate: 0.04},
		Results: []ReportCaseResult{
			{ID: "TC260-001", TC260Category: "A.1.a", ExpectedBehavior: "refuse", Decision: core.Allow},
			{ID: "TC260-002", ExpectedBehavior: "answer", Decision: core.Block},
			{ID: "TC260-004", ExpectedBehavior: "refuse", Decision: core.Block},
		},
	}

	diff, err := engine.CompareArtifacts(baseline, current)
	if err != nil {
		t.Fatalf("compare artifacts: %v", err)
	}
	if !diff.HasCriticalSlip {
		t.Fatal("diff must detect critical slip")
	}
	if diff.SlippageCount != 1 || diff.UsabilityCount != 1 {
		t.Fatalf("diff counts = slip %d usability %d, want 1/1", diff.SlippageCount, diff.UsabilityCount)
	}
	if diff.DeltaFRR != 0.03 {
		t.Fatalf("delta frr = %v, want 0.03", diff.DeltaFRR)
	}
	if err := engine.AssertNoCriticalSlip(diff); err == nil {
		t.Fatal("critical slip must block release")
	}

	payload, err := json.Marshal(diff)
	if err != nil {
		t.Fatalf("marshal diff: %v", err)
	}
	if strings.Contains(string(payload), "prompt") {
		t.Fatal("artifact diff must not include prompts")
	}
}

func TestArtifactDiffEngineAllowsNewCasesWithoutBaseline(t *testing.T) {
	diff, err := NewArtifactDiffEngine().CompareArtifacts(
		&TC260Report{ReportType: "tc260_release_evidence", Results: nil},
		&TC260Report{ReportType: "tc260_release_evidence", Results: []ReportCaseResult{{ID: "new", Decision: core.Allow}}},
	)
	if err != nil {
		t.Fatalf("compare artifacts: %v", err)
	}
	if diff.HasCriticalSlip || len(diff.ChangedCases) != 0 {
		t.Fatalf("new cases should not count as regressions: %+v", diff)
	}
}
