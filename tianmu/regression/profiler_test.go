package regression

import (
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

func TestProfilerAggregatesCategoryAndDifficulty(t *testing.T) {
	profiler := NewProfiler()

	profiler.ProfileCase(TC260Case{
		ID:               "attack-1",
		TC260Category:    "A.1.a",
		ExpectedBehavior: "refuse",
		Attributes:       map[string]any{"difficulty": "hard"},
	}, core.Allow, false)
	profiler.ProfileCase(TC260Case{
		ID:               "safe-1",
		ExpectedBehavior: "answer",
		Attributes:       map[string]any{"difficulty": "easy"},
	}, core.Block, false)

	if profiler.CategoryStats["A.1.a"].FalseNegative != 1 {
		t.Fatalf("A.1.a FN = %d, want 1", profiler.CategoryStats["A.1.a"].FalseNegative)
	}
	if profiler.CategoryStats["positive_control"].FalsePositive != 1 {
		t.Fatalf("positive_control FP = %d, want 1", profiler.CategoryStats["positive_control"].FalsePositive)
	}
	if profiler.DifficultyStats["hard"].Failed != 1 || profiler.DifficultyStats["easy"].Failed != 1 {
		t.Fatalf("difficulty stats = %+v", profiler.DifficultyStats)
	}
}
