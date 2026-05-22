package regression

import (
	"strings"
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

func TestRegressionDiffEngineBlocksRefuseToAllowDegradation(t *testing.T) {
	engine := NewRegressionDiffEngine()
	results := []TC260Result{
		{
			Case:     TC260Case{ID: "TC260-A109-02", ExpectedBehavior: "refuse"},
			Decision: core.Block,
			Passed:   true,
		},
		{
			Case:       TC260Case{ID: "TC260-A109-03", ExpectedBehavior: "refuse"},
			Decision:   core.Allow,
			ReasonCode: "pass_default",
			Passed:     false,
		},
	}

	err := engine.AssertNoDegradation(results)
	if err == nil {
		t.Fatal("release gate must reject refuse-to-allow degradation")
	}
	if !strings.Contains(err.Error(), "release_gate_blocked") {
		t.Fatalf("error = %q, want release_gate_blocked", err)
	}

	degradations := engine.FindDegradations(results)
	if len(degradations) != 1 {
		t.Fatalf("degradations = %d, want 1", len(degradations))
	}
	if degradations[0].ID != "TC260-A109-03" {
		t.Fatalf("degradation id = %q, want TC260-A109-03", degradations[0].ID)
	}
}

func TestRegressionDiffEngineAllowsPositiveControlOverblockForReportOnly(t *testing.T) {
	engine := NewRegressionDiffEngine()
	results := []TC260Result{
		{
			Case:     TC260Case{ID: "safe-1", ExpectedBehavior: "answer"},
			Decision: core.Block,
			Passed:   false,
		},
	}

	if err := engine.AssertNoDegradation(results); err != nil {
		t.Fatalf("positive-control overblock is reportable but not a critical degradation: %v", err)
	}
}
