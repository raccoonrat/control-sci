package core

import "testing"

func TestEvaluateActionMediatesChinesePIIToRedact(t *testing.T) {
	evaluator, err := NewEvaluator(PolicyPack{Version: "test-policy-v1"})
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}

	decision := evaluator.EvaluateAction(
		RiskEvaluation{
			Signals: []DetectorSignal{
				{
					DetectorID: "cn-pii-fastpath",
					Category:   ChinesePIICategory,
					Version:    "cn-pii-v1",
					Confidence: 0.91,
					Triggered:  true,
				},
			},
		},
		ActionContext{ActionType: "call_tool"},
	)

	if decision.Decision != RedactThenAllow {
		t.Fatalf("decision = %q, want %q", decision.Decision, RedactThenAllow)
	}
	if decision.ReasonCode != "cn_privacy_leakage_mediated_to_redact" {
		t.Fatalf("reason = %q, want cn_privacy_leakage_mediated_to_redact", decision.ReasonCode)
	}
}

func TestEvaluateActionKeepsBlockAboveSideEffectConfirmation(t *testing.T) {
	evaluator, err := NewEvaluator(PolicyPack{
		Version: "test-policy-v1",
		Rules: []PolicyRule{
			{
				RiskCategory:        "prompt_injection",
				ConfidenceThreshold: 0.80,
				TargetDecision:      Block,
				ReasonCode:          "prompt_injection_blocked",
			},
		},
	})
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}

	decision := evaluator.EvaluateAction(
		RiskEvaluation{
			Signals: []DetectorSignal{
				{
					DetectorID: "cn-injection-fastpath",
					Category:   "prompt_injection",
					Version:    "cn-injection-v1",
					Confidence: 0.93,
					Triggered:  true,
				},
			},
		},
		ActionContext{ActionType: "call_tool", SideEffect: true},
	)

	if decision.Decision != Block {
		t.Fatalf("decision = %q, want %q", decision.Decision, Block)
	}
}
