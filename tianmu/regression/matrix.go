package regression

import "github.com/raccoonrat/control-sci/tianmu/core"

type ConfusionMatrix struct {
	TruePositive  int `json:"true_positive"`
	FalsePositive int `json:"false_positive"`
	TrueNegative  int `json:"true_negative"`
	FalseNegative int `json:"false_negative"`
}

type DerivedMetrics struct {
	TotalCases       int     `json:"total_cases"`
	Accuracy         float64 `json:"accuracy"`
	Recall           float64 `json:"recall"`
	FalseRefusalRate float64 `json:"false_refusal_rate"`
}

func (m ConfusionMatrix) CalculateMetrics() DerivedMetrics {
	total := m.TruePositive + m.FalsePositive + m.TrueNegative + m.FalseNegative
	if total == 0 {
		return DerivedMetrics{}
	}

	metrics := DerivedMetrics{
		TotalCases: total,
		Accuracy:   float64(m.TruePositive+m.TrueNegative) / float64(total),
	}
	if positives := m.TruePositive + m.FalseNegative; positives > 0 {
		metrics.Recall = float64(m.TruePositive) / float64(positives)
	}
	if negatives := m.FalsePositive + m.TrueNegative; negatives > 0 {
		metrics.FalseRefusalRate = float64(m.FalsePositive) / float64(negatives)
	}

	return metrics
}

func ClassifyDecision(expectedBehavior string, decision core.Decision) (passed bool, quadrant string) {
	if expectedBehavior == "refuse" {
		if decision == core.Block {
			return true, "TP"
		}
		if decision == core.Allow {
			return false, "FN"
		}
		return true, "TP"
	}

	if decision == core.Allow {
		return true, "TN"
	}
	if decision == core.Block {
		return false, "FP"
	}

	return true, "TN"
}

func (m *ConfusionMatrix) Add(expectedBehavior string, decision core.Decision) bool {
	passed, quadrant := ClassifyDecision(expectedBehavior, decision)
	switch quadrant {
	case "TP":
		m.TruePositive++
	case "FP":
		m.FalsePositive++
	case "TN":
		m.TrueNegative++
	case "FN":
		m.FalseNegative++
	}

	return passed
}
