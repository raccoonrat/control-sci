package core

import (
	"encoding/json"
	"testing"
)

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

func TestEvaluatorSelectsHighestImpactDecision(t *testing.T) {
	engine := newTestEngine(t)

	decision, err := engine.MediateInbound(
		RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "user_io"},
		IdentityContext{ActorID: "user-dev-001"},
		DataContext{DataClassification: "personal_sensitive", ContainsPII: true, Source: "user_input", Destination: "model_context"},
		ActionContext{ActionType: "generate_response"},
		[]DetectorSignal{
			{
				DetectorID: "cn-pii-fastpath",
				Category:   "chinese_pii",
				Version:    "cn-pii-v1",
				Confidence: 0.90,
				Triggered:  true,
			},
			{
				DetectorID: "cn-injection-fastpath",
				Category:   "prompt_injection",
				Version:    "cn-injection-v1",
				Confidence: 0.88,
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
}

func TestControlDecisionObjectMarshalsReleaseEvidence(t *testing.T) {
	engine := newTestEngine(t)

	decision, err := engine.MediateInbound(
		RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "user_io"},
		IdentityContext{ActorID: "user-dev-001"},
		DataContext{DataClassification: "public", Source: "user_input", Destination: "model_context"},
		ActionContext{ActionType: "generate_response"},
		nil,
	)
	if err != nil {
		t.Fatalf("mediate inbound: %v", err)
	}

	payload, err := json.Marshal(decision)
	if err != nil {
		t.Fatalf("marshal decision object: %v", err)
	}

	var decoded struct {
		ReleaseEvidence ReleaseEvidenceLite `json:"release_evidence"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal decision object: %v", err)
	}

	if decoded.ReleaseEvidence.EvidenceLevel != "release_evidence_lite" {
		t.Fatalf("evidence level = %q, want release_evidence_lite", decoded.ReleaseEvidence.EvidenceLevel)
	}
	if decoded.ReleaseEvidence.TraceID == "" {
		t.Fatal("trace id must survive JSON serialization")
	}
	if !decoded.ReleaseEvidence.RegressionPassTag {
		t.Fatal("regression pass tag must survive JSON serialization")
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

func BenchmarkMediateInboundFastPath(b *testing.B) {
	engine := newBenchmarkEngine(b)
	req := RequestContext{ProductID: "Qira", Language: "zh-CN", InteractionType: "user_io"}
	identity := IdentityContext{ActorID: "user-dev-001"}
	data := DataContext{DataClassification: "public", Source: "user_input", Destination: "model_context"}
	action := ActionContext{ActionType: "generate_response"}
	signals := []DetectorSignal{
		{
			DetectorID: "cn-injection-fastpath",
			Category:   "prompt_injection",
			Version:    "cn-injection-v1",
			Confidence: 0.92,
			Triggered:  true,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decision, err := engine.MediateInbound(req, identity, data, action, signals)
		if err != nil {
			b.Fatalf("mediate inbound: %v", err)
		}
		if decision.PolicyDecision.Decision != Block {
			b.Fatalf("decision = %q, want %q", decision.PolicyDecision.Decision, Block)
		}
	}
}

func newBenchmarkEngine(b *testing.B) *Engine {
	b.Helper()

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
		b.Fatalf("new evaluator: %v", err)
	}

	engine, err := NewEngine(PersonalAI, evaluator)
	if err != nil {
		b.Fatalf("new engine: %v", err)
	}

	return engine
}
