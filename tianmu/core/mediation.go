package core

const (
	ChinesePIICategory = "chinese_pii"
)

func (e *Evaluator) EvaluateAction(risk RiskEvaluation, action ActionContext) PolicyDecision {
	decision := e.Evaluate(risk)
	if decision.Decision != Allow {
		return decision
	}

	if action.SideEffect {
		return PolicyDecision{
			Decision:          AskConfirmation,
			PolicyPackVersion: e.pack.Version,
			ReasonCode:        "side_effect_action_requires_approval",
		}
	}

	if hasTriggeredCategory(risk, ChinesePIICategory, 0.85) {
		return PolicyDecision{
			Decision:          RedactThenAllow,
			PolicyPackVersion: e.pack.Version,
			ReasonCode:        "cn_privacy_leakage_mediated_to_redact",
		}
	}

	return decision
}

func hasTriggeredCategory(risk RiskEvaluation, category string, threshold float64) bool {
	for _, signal := range risk.Signals {
		if !signal.Triggered || signal.Confidence < threshold {
			continue
		}
		if signal.Category == category || signal.DetectorID == category {
			return true
		}
	}

	return false
}
