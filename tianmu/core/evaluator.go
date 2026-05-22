package core

import "errors"

type PolicyRule struct {
	RiskCategory        string   `json:"risk_category"`
	ConfidenceThreshold float64  `json:"confidence_threshold"`
	TargetDecision      Decision `json:"target_decision"`
	ReasonCode          string   `json:"reason_code"`
}

type PolicyPack struct {
	Version string       `json:"version"`
	Rules   []PolicyRule `json:"rules"`
}

type Evaluator struct {
	pack PolicyPack
}

func NewEvaluator(pack PolicyPack) (*Evaluator, error) {
	if pack.Version == "" {
		return nil, errors.New("policy pack version is required")
	}
	for _, rule := range pack.Rules {
		if rule.RiskCategory == "" {
			return nil, errors.New("policy rule risk category is required")
		}
		if rule.ReasonCode == "" {
			return nil, errors.New("policy rule reason code is required")
		}
		if rule.ConfidenceThreshold < 0 || rule.ConfidenceThreshold > 1 {
			return nil, errors.New("policy rule confidence threshold must be between 0 and 1")
		}
	}

	return &Evaluator{pack: pack}, nil
}

func (e *Evaluator) Evaluate(risk RiskEvaluation) PolicyDecision {
	selected := PolicyDecision{
		Decision:          Allow,
		PolicyPackVersion: e.pack.Version,
		ReasonCode:        "pass_default",
	}

	for _, signal := range risk.Signals {
		if !signal.Triggered {
			continue
		}

		for _, rule := range e.pack.Rules {
			if signal.Category != rule.RiskCategory || signal.Confidence < rule.ConfidenceThreshold {
				continue
			}

			candidate := PolicyDecision{
				Decision:          rule.TargetDecision,
				PolicyPackVersion: e.pack.Version,
				ReasonCode:        rule.ReasonCode,
			}
			if outranks(candidate.Decision, selected.Decision) {
				selected = candidate
			}
		}
	}

	return selected
}

func outranks(left Decision, right Decision) bool {
	return decisionRank(left) > decisionRank(right)
}

func decisionRank(decision Decision) int {
	switch decision {
	case Block:
		return 100
	case Escalate:
		return 90
	case AskConfirmation:
		return 80
	case Rewrite:
		return 70
	case RedactThenAllow:
		return 60
	case LogOnly:
		return 20
	case Allow:
		return 0
	default:
		return -1
	}
}
