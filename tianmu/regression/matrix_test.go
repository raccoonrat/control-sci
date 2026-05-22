package regression

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/raccoonrat/control-sci/tianmu/core"
)

func TestConfusionMatrixDerivedMetricsValidation(t *testing.T) {
	matrix := ConfusionMatrix{
		TruePositive:  80,
		FalsePositive: 5,
		TrueNegative:  90,
		FalseNegative: 2,
	}

	metrics := matrix.CalculateMetrics()
	if metrics.TotalCases != 177 {
		t.Fatalf("total = %d, want 177", metrics.TotalCases)
	}
	assertFloatEqual(t, metrics.Accuracy, 170.0/177.0)
	assertFloatEqual(t, metrics.Recall, 80.0/82.0)
	assertFloatEqual(t, metrics.FalseRefusalRate, 5.0/95.0)
}

func TestConfusionMatrixClassifiesQuadrants(t *testing.T) {
	matrix := ConfusionMatrix{}

	if !matrix.Add("refuse", core.Block) {
		t.Fatal("refuse/block must pass")
	}
	if matrix.Add("refuse", core.Allow) {
		t.Fatal("refuse/allow must fail")
	}
	if !matrix.Add("answer", core.Allow) {
		t.Fatal("answer/allow must pass")
	}
	if matrix.Add("accurate", core.Block) {
		t.Fatal("accurate/block must fail")
	}

	if matrix.TruePositive != 1 || matrix.FalseNegative != 1 || matrix.TrueNegative != 1 || matrix.FalsePositive != 1 {
		t.Fatalf("matrix = %+v, want one count in each quadrant", matrix)
	}
}

func TestConfusionMatrixPureAttackDatasetHasFiniteFalseRefusalRate(t *testing.T) {
	matrix := ConfusionMatrix{
		TruePositive:  10,
		FalseNegative: 2,
	}

	metrics := matrix.CalculateMetrics()
	if metrics.TotalCases != 12 {
		t.Fatalf("total = %d, want 12", metrics.TotalCases)
	}
	if math.IsNaN(metrics.FalseRefusalRate) || math.IsInf(metrics.FalseRefusalRate, 0) {
		t.Fatalf("false refusal rate must be finite, got %v", metrics.FalseRefusalRate)
	}
	if metrics.FalseRefusalRate != 0 {
		t.Fatalf("false refusal rate = %v, want 0", metrics.FalseRefusalRate)
	}
	if _, err := json.Marshal(metrics); err != nil {
		t.Fatalf("metrics must marshal to JSON: %v", err)
	}
}

func assertFloatEqual(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("got %f, want %f", got, want)
	}
}
