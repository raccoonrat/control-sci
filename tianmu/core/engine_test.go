package core

import "testing"

func TestMediateInboundBlocksPromptInjection(t *testing.T) {
	engine := newTestEngine(t)

	decision, err := engine.MediateInbound(
		RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "user_io"},
		IdentityContext{ActorID: "user-dev-001"},
		DataContext{DataClassification: "public", Source: "user_input", Destination: "model_context"},
		ActionContext{ActionType: "generate_response"},
		[]DetectorSignal{
			{
				DetectorID: "cn-injection-fastpath",
				Category:   "prompt_injection",
				Version:    "cn-injection-v1",
				Confidence: 0.92,
				Triggered:  true,
			},
		},
	)
	if err != nil {
		t.Fatalf("mediate inbound: %v", err)
	}

	if decision.PolicyDecision.Decision != Block {
		t.Fatalf("decision = %q, want %q", decision.PolicyDecision.Decision, Block)
	}
	if decision.PolicyDecision.ReasonCode != "cn_prompt_injection_blocked" {
		t.Fatalf("reason = %q, want cn_prompt_injection_blocked", decision.PolicyDecision.ReasonCode)
	}
	if decision.ReleaseEvidence.TraceID == "" {
		t.Fatal("trace id must be generated")
	}
}

func TestMediateInboundRequiresConfirmationForSideEffects(t *testing.T) {
	engine := newTestEngine(t)

	decision, err := engine.MediateInbound(
		RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "agent_loop"},
		IdentityContext{ActorID: "user-dev-001"},
		DataContext{DataClassification: "personal_sensitive", Source: "user_input", Destination: "external_api"},
		ActionContext{ActionType: "call_tool", ToolName: "send_message", SideEffect: true},
		nil,
	)
	if err != nil {
		t.Fatalf("mediate inbound: %v", err)
	}

	if decision.PolicyDecision.Decision != AskConfirmation {
		t.Fatalf("decision = %q, want %q", decision.PolicyDecision.Decision, AskConfirmation)
	}
	if decision.PolicyDecision.ReasonCode != "side_effect_action_requires_approval" {
		t.Fatalf("reason = %q, want side_effect_action_requires_approval", decision.PolicyDecision.ReasonCode)
	}
}

func TestSummarizeRiskDeduplicatesCategoriesAndVersions(t *testing.T) {
	risk := summarizeRisk([]DetectorSignal{
		{Category: "prompt_injection", Version: "detector-v1", Confidence: 0.7, Triggered: true},
		{Category: "prompt_injection", Version: "detector-v1", Confidence: 0.9, Triggered: true},
		{Category: "chinese_pii", Version: "pii-v1", Confidence: 0.8, Triggered: false},
	})

	if len(risk.RiskCategories) != 1 || risk.RiskCategories[0] != "prompt_injection" {
		t.Fatalf("risk categories = %#v, want only prompt_injection", risk.RiskCategories)
	}
	if len(risk.DetectorVersions) != 2 {
		t.Fatalf("detector versions = %#v, want two unique versions", risk.DetectorVersions)
	}
	if risk.MaxRiskScore != 0.9 {
		t.Fatalf("max risk score = %v, want 0.9", risk.MaxRiskScore)
	}
}

func newTestEngine(t *testing.T) *Engine {
	t.Helper()

	evaluator, err := NewEvaluator(PolicyPack{
		Version: "personal-ai-cn-policy-v1",
		Rules: []PolicyRule{
			{
				RiskCategory:        "prompt_injection",
				ConfidenceThreshold: 0.80,
				TargetDecision:      Block,
				ReasonCode:          "cn_prompt_injection_blocked",
			},
			{
				RiskCategory:        "chinese_pii",
				ConfidenceThreshold: 0.85,
				TargetDecision:      RedactThenAllow,
				ReasonCode:          "cn_privacy_redaction_triggered",
			},
		},
	})
	if err != nil {
		t.Fatalf("new evaluator: %v", err)
	}

	engine, err := NewEngine(PersonalAI, evaluator)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	return engine
}
